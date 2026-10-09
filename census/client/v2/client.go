// Package v2 is a self-contained Go client for the Census v2 API
// (https://app.getcensus.com/api/v2). It has no dependency on the v1 client
// in census/client — v1 is slated for deletion once every resource has been
// migrated to this package, so nothing in this package may import it.
//
// v2 is personal-access-token-only: there is no workspace-token concept at
// all, unlike v1.
package v2

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"
)

// BaseURL is the only v2 API root — no regional variants exist, unlike v1.
const BaseURL = "https://app.getcensus.com/api/v2"

const (
	defaultHTTPTimeout = 60 * time.Second
	retryBudgetTimeout = 5 * time.Minute // total time budget for retrying 429s
	initialRetryDelay  = 1 * time.Second
	maxRetryDelay      = 90 * time.Second
	backoffMultiplier  = 2.0
	jitterFactor       = 0.2
)

// Config configures a Client.
type Config struct {
	// PersonalAccessToken authenticates every request. Required.
	PersonalAccessToken string
	// BaseURL is the API root. Always "https://app.getcensus.com/api/v2" in
	// production — v2 has no regional endpoints, unlike v1.
	BaseURL string
	// HTTPClient overrides the default *http.Client. Mainly for tests; nil
	// uses a default client with defaultHTTPTimeout.
	HTTPClient *http.Client
}

// Client is a Census v2 API client.
type Client struct {
	config     *Config
	httpClient *http.Client
}

// NewClient creates a Client from Config.
func NewClient(config *Config) (*Client, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if config.PersonalAccessToken == "" {
		return nil, fmt.Errorf("personal access token is required")
	}
	if config.BaseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}

	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultHTTPTimeout}
	}

	return &Client{config: config, httpClient: httpClient}, nil
}

// do sends a request with a JSON-encodable body (nil for none) and returns
// the raw response. Callers decode it with decode/decodeList/drain.
func (c *Client) do(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, c.config.BaseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "terraform-provider-census")
	req.Header.Set("Authorization", "Bearer "+c.config.PersonalAccessToken)

	return c.doWithRetry(ctx, req, bodyBytes)
}

// doWithRetry executes req, retrying on 429 with exponential backoff and
// jitter (honoring Retry-After when present) until retryBudgetTimeout is
// exhausted. Any other status, or a non-HTTP error, returns immediately.
func (c *Client) doWithRetry(ctx context.Context, req *http.Request, bodyBytes []byte) (*http.Response, error) {
	deadline := time.Now().Add(retryBudgetTimeout)
	attempt := 0

	for {
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("request timed out after %d attempts: rate limited (429)", attempt)
		}
		attempt++

		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			req.GetBody = func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader(bodyBytes)), nil
			}
			req.ContentLength = int64(len(bodyBytes))
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusTooManyRequests {
			return resp, nil
		}

		delay := retryDelay(resp, attempt, deadline)
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		if time.Now().Add(delay).After(deadline) {
			return nil, fmt.Errorf("rate limit retry delay (%v) would exceed timeout after %d attempts", delay, attempt)
		}

		timer := time.NewTimer(delay)
		select {
		case <-timer.C:
		case <-ctx.Done():
			timer.Stop()
			return nil, fmt.Errorf("request cancelled during rate limit retry: %w", ctx.Err())
		}
	}
}

// retryDelay prefers the Retry-After header; falls back to exponential
// backoff with jitter, capped so it never exceeds deadline.
func retryDelay(resp *http.Response, attempt int, deadline time.Time) time.Duration {
	if header := resp.Header.Get("Retry-After"); header != "" {
		if delay, err := parseRetryAfter(header); err == nil {
			if remaining := time.Until(deadline); delay > remaining {
				return remaining
			}
			return delay
		}
	}

	delay := initialRetryDelay
	for i := 1; i < attempt; i++ {
		delay = time.Duration(float64(delay) * backoffMultiplier)
		if delay > maxRetryDelay {
			delay = maxRetryDelay
			break
		}
	}

	jitter := float64(delay) * jitterFactor * (2*rand.Float64() - 1)
	delay += time.Duration(jitter)

	if remaining := time.Until(deadline); delay > remaining {
		return remaining
	}
	return delay
}

// parseRetryAfter parses a Retry-After header per RFC 7231 (either a
// delay-seconds integer or an HTTP-date).
func parseRetryAfter(header string) (time.Duration, error) {
	if seconds, err := strconv.ParseInt(header, 10, 64); err == nil {
		return time.Duration(seconds) * time.Second, nil
	}

	for _, layout := range []string{time.RFC1123, time.RFC850, time.ANSIC} {
		if t, err := time.Parse(layout, header); err == nil {
			if delay := time.Until(t); delay > 0 {
				return delay, nil
			}
			return 0, nil
		}
	}

	return 0, fmt.Errorf("invalid Retry-After format: %s", header)
}
