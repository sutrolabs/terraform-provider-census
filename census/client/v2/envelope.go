package v2

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

// Pagination mirrors the `pagination` object returned by v2 list endpoints
// (apidocs/v2/components/common.yaml#/Pagination).
type Pagination struct {
	TotalRecords int  `json:"total_records"`
	PerPage      int  `json:"per_page"`
	Page         int  `json:"page"`
	PrevPage     *int `json:"prev_page"`
	NextPage     *int `json:"next_page"`
	LastPage     int  `json:"last_page"`
}

// ListOptions are the common page/per_page/order query parameters accepted
// by v2 list endpoints.
type ListOptions struct {
	Page    int
	PerPage int
	Order   string
}

func (o ListOptions) params() map[string]string {
	params := make(map[string]string)
	if o.Page > 0 {
		params["page"] = strconv.Itoa(o.Page)
	}
	if o.PerPage > 0 {
		params["per_page"] = strconv.Itoa(o.PerPage)
	}
	if o.Order != "" {
		params["order"] = o.Order
	}
	return params
}

// buildURL appends query parameters to path. path is relative — the base URL
// is added separately by Client.do.
func buildURL(path string, params map[string]string) string {
	if len(params) == 0 {
		return path
	}

	q := url.Values{}
	for k, v := range params {
		q.Set(k, v)
	}
	return path + "?" + q.Encode()
}

// envelope is the `{status, data}` shape nearly every v2 endpoint uses on
// success.
type envelope[T any] struct {
	Status string `json:"status"`
	Data   T      `json:"data"`
}

// paginatedEnvelope is the list-endpoint variant, additionally carrying
// pagination metadata.
type paginatedEnvelope[T any] struct {
	Status     string     `json:"status"`
	Data       []T        `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// readBody reads and closes resp.Body, mapping any HTTP error status to an
// *APIError. On success it returns the raw body bytes for the caller to
// unmarshal.
func readBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, newAPIError(resp.StatusCode, body)
	}

	return body, nil
}

// decode reads resp's body into an envelope[T] and returns its Data. `what`
// describes the resource for the error message if the API returns a
// successful status with no body at all — see errEmptyBody.
func decode[T any](resp *http.Response, what string) (T, error) {
	var zero T

	body, err := readBody(resp)
	if err != nil {
		return zero, err
	}
	if len(body) == 0 {
		return zero, errEmptyBody(what)
	}

	var env envelope[T]
	if err := json.Unmarshal(body, &env); err != nil {
		return zero, fmt.Errorf("failed to decode response JSON: %w\n\nRaw API response:\n%s", err, string(body))
	}

	return env.Data, nil
}

// decodeList is decode's list-endpoint counterpart, also returning
// pagination metadata.
func decodeList[T any](resp *http.Response, what string) ([]T, Pagination, error) {
	var zero Pagination

	body, err := readBody(resp)
	if err != nil {
		return nil, zero, err
	}
	if len(body) == 0 {
		return nil, zero, errEmptyBody(what)
	}

	var env paginatedEnvelope[T]
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, zero, fmt.Errorf("failed to decode response JSON: %w\n\nRaw API response:\n%s", err, string(body))
	}

	return env.Data, env.Pagination, nil
}

// drain reads and discards resp's body, surfacing only an HTTP-error-status
// mapping. Use for calls whose response body carries nothing the caller
// needs.
func drain(resp *http.Response) error {
	_, err := readBody(resp)
	return err
}

// Get issues a GET request to path and decodes the `{status, data}` envelope
// into T. `what` describes the resource, used only in the empty-body error
// message (e.g. "source 123").
func Get[T any](ctx context.Context, c *Client, path string, what string) (T, error) {
	var zero T
	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return zero, err
	}
	return decode[T](resp, what)
}

// List issues a GET request to path with opts applied as query parameters
// and decodes the paginated envelope into a slice of T.
func List[T any](ctx context.Context, c *Client, path string, opts ListOptions, what string) ([]T, Pagination, error) {
	resp, err := c.do(ctx, http.MethodGet, buildURL(path, opts.params()), nil)
	if err != nil {
		return nil, Pagination{}, err
	}
	return decodeList[T](resp, what)
}

// Create issues a POST request with body and decodes the `{status, data}`
// envelope into T.
func Create[T any](ctx context.Context, c *Client, path string, body interface{}, what string) (T, error) {
	var zero T
	resp, err := c.do(ctx, http.MethodPost, path, body)
	if err != nil {
		return zero, err
	}
	return decode[T](resp, what)
}

// Update issues a PATCH request with body and decodes the `{status, data}`
// envelope into T.
func Update[T any](ctx context.Context, c *Client, path string, body interface{}, what string) (T, error) {
	var zero T
	resp, err := c.do(ctx, http.MethodPatch, path, body)
	if err != nil {
		return zero, err
	}
	return decode[T](resp, what)
}

// Delete issues a DELETE request to path. v2's delete responses only ever
// confirm `{"status": "deleted"}` — callers that don't need to assert that
// explicitly can ignore it and just check the returned error.
func Delete(ctx context.Context, c *Client, path string) error {
	resp, err := c.do(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	return drain(resp)
}
