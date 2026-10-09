package v2

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// APIError represents an HTTP-level error response from the Census v2 API.
// v2's documented error-body shapes are inconsistent as of this writing —
// some endpoints specify a `message` field (e.g. BadRequest400 in
// apidocs/v2/components/responses.yaml), while others — notably 404 and
// 409 — are left as empty, undocumented placeholders in that same file. So
// Message here is best-effort: populated from a `message` field if present,
// otherwise the raw response body. Callers should not assume either case is
// confirmed API behavior; verify against a real staging response before
// relying on Message's content for anything beyond logging.
type APIError struct {
	StatusCode int
	Message    string
	Body       []byte
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("Census API error (status %d): %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("Census API error (status %d)", e.StatusCode)
}

func newAPIError(statusCode int, body []byte) *APIError {
	apiErr := &APIError{StatusCode: statusCode, Body: body}

	var parsed struct {
		Message string `json:"message"`
	}
	if json.Unmarshal(body, &parsed) == nil && parsed.Message != "" {
		apiErr.Message = parsed.Message
	} else if len(body) > 0 {
		apiErr.Message = string(body)
	}

	return apiErr
}

// IsNotFoundError reports whether err is an *APIError for a 404 response.
func IsNotFoundError(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}

// IsConflictError reports whether err is an *APIError for a 409 response.
func IsConflictError(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusConflict
}

// errEmptyBody reports that the API responded with a successful status but
// an empty body — there is no confirmed case where this represents a
// genuine "not found" or "deleted" (this mirrors the v1 client's equivalent
// fix in census/client/client.go's errEmptyResponse, carried forward here
// since nothing in apidocs/v2 rules the scenario out for v2 either). Callers
// should treat it as an unexpected hard error, not infer resource absence.
func errEmptyBody(what string) error {
	return fmt.Errorf("Census API returned a successful response with no data for %s — this is unexpected and may indicate an API or network issue", what)
}
