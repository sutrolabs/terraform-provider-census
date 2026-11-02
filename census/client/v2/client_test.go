package v2

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	c, err := NewClient(&Config{
		PersonalAccessToken: "test-token",
		BaseURL:             server.URL,
	})
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	return c
}

func TestNewClient_Validation(t *testing.T) {
	if _, err := NewClient(nil); err == nil {
		t.Fatal("expected error for nil config")
	}
	if _, err := NewClient(&Config{BaseURL: "https://example.com"}); err == nil {
		t.Fatal("expected error for missing token")
	}
	if _, err := NewClient(&Config{PersonalAccessToken: "t"}); err == nil {
		t.Fatal("expected error for missing base URL")
	}
}

type testSource struct {
	ID      int    `json:"id"`
	GroupID string `json:"group_id"`
}

func TestGet_AuthHeaderAndEnvelope(t *testing.T) {
	var gotAuth, gotPath string

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"success","data":{"id":4,"group_id":"wandering_aimlessly"}}`))
	})

	source, err := Get[testSource](context.Background(), c, "/sources/4", "source 4")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if gotAuth != "Bearer test-token" {
		t.Fatalf("expected bearer auth header, got %q", gotAuth)
	}
	if gotPath != "/sources/4" {
		t.Fatalf("expected path /sources/4, got %q", gotPath)
	}
	if source.ID != 4 || source.GroupID != "wandering_aimlessly" {
		t.Fatalf("unexpected decoded source: %+v", source)
	}
}

func TestGet_NotFound(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status":404}`))
	})

	_, err := Get[testSource](context.Background(), c, "/sources/999", "source 999")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !IsNotFoundError(err) {
		t.Fatalf("expected IsNotFoundError to be true, got %v", err)
	}
	if IsConflictError(err) {
		t.Fatal("expected IsConflictError to be false for a 404")
	}
}

func TestGet_Conflict(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"status":409}`))
	})

	_, err := Get[testSource](context.Background(), c, "/sources/4", "source 4")
	if err == nil {
		t.Fatal("expected error for 409 response")
	}
	if !IsConflictError(err) {
		t.Fatalf("expected IsConflictError to be true, got %v", err)
	}
	if IsNotFoundError(err) {
		t.Fatal("expected IsNotFoundError to be false for a 409")
	}
}

func TestNewAPIError_MessageFromJSON(t *testing.T) {
	apiErr := newAPIError(http.StatusBadRequest, []byte(`{"status":400,"message":"Invalid type: not_a_real_type"}`))

	if apiErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", apiErr.StatusCode)
	}
	if apiErr.Message != "Invalid type: not_a_real_type" {
		t.Fatalf("expected message extracted from JSON, got %q", apiErr.Message)
	}
	if !strings.Contains(apiErr.Error(), "Invalid type: not_a_real_type") {
		t.Fatalf("expected Error() to include the message, got %q", apiErr.Error())
	}
}

func TestNewAPIError_FallsBackToRawBodyWhenNoMessageField(t *testing.T) {
	// Matches apidocs/v2/components/responses.yaml's actual NotFound shape —
	// a bare {"status": <code>} with no "message" field at all.
	body := []byte(`{"status":404}`)

	apiErr := newAPIError(http.StatusNotFound, body)

	if apiErr.Message != string(body) {
		t.Fatalf("expected raw body fallback when no message field present, got %q", apiErr.Message)
	}
}

func TestNewAPIError_FallsBackToRawBodyWhenNotJSON(t *testing.T) {
	// A proxy/gateway in front of the real API returning a non-JSON error
	// page is a realistic failure mode, not just a hypothetical.
	body := []byte("<html><body>502 Bad Gateway</body></html>")

	apiErr := newAPIError(http.StatusBadGateway, body)

	if apiErr.Message != string(body) {
		t.Fatalf("expected raw body fallback for non-JSON body, got %q", apiErr.Message)
	}
}

func TestGet_EmptyBodyIsError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Deliberately no body.
	})

	_, err := Get[testSource](context.Background(), c, "/sources/4", "source 4")
	if err == nil {
		t.Fatal("expected error for empty body")
	}
	if IsNotFoundError(err) {
		t.Fatal("empty body should not be classified as not found")
	}
	if !strings.Contains(err.Error(), "no data") {
		t.Fatalf("expected empty-body error message, got %q", err.Error())
	}
}

func TestList_Pagination(t *testing.T) {
	var gotQuery string

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"success","data":[{"id":1,"group_id":"a"},{"id":2,"group_id":"b"}],"pagination":{"total_records":2,"per_page":25,"page":1,"prev_page":null,"next_page":null,"last_page":1}}`))
	})

	sources, pagination, err := List[testSource](context.Background(), c, "/sources", ListOptions{Page: 2, PerPage: 10, Order: "asc"}, "sources")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(sources) != 2 {
		t.Fatalf("expected 2 sources, got %d", len(sources))
	}
	if pagination.TotalRecords != 2 {
		t.Fatalf("expected total_records 2, got %d", pagination.TotalRecords)
	}
	if !strings.Contains(gotQuery, "page=2") || !strings.Contains(gotQuery, "per_page=10") || !strings.Contains(gotQuery, "order=asc") {
		t.Fatalf("expected page/per_page/order query params, got %q", gotQuery)
	}
}

type testWriteRequest struct {
	Name string `json:"name"`
}

func TestCreate_SendsMethodPathAndBody(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody []byte

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"created","data":{"id":4,"group_id":"wandering_aimlessly"}}`))
	})

	source, err := Create[testSource](context.Background(), c, "/sources", testWriteRequest{Name: "my source"}, "source")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Fatalf("expected POST, got %s", gotMethod)
	}
	if gotPath != "/sources" {
		t.Fatalf("expected path /sources, got %q", gotPath)
	}
	if !strings.Contains(string(gotBody), `"name":"my source"`) {
		t.Fatalf("expected request body to contain the marshaled name field, got %s", gotBody)
	}
	if source.ID != 4 || source.GroupID != "wandering_aimlessly" {
		t.Fatalf("unexpected decoded source: %+v", source)
	}
}

func TestUpdate_SendsMethodPathAndBody(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody []byte

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"updated","data":{"id":4,"group_id":"wandering_aimlessly"}}`))
	})

	source, err := Update[testSource](context.Background(), c, "/sources/4", testWriteRequest{Name: "renamed"}, "source 4")
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if gotMethod != http.MethodPatch {
		t.Fatalf("expected PATCH, got %s", gotMethod)
	}
	if gotPath != "/sources/4" {
		t.Fatalf("expected path /sources/4, got %q", gotPath)
	}
	if !strings.Contains(string(gotBody), `"name":"renamed"`) {
		t.Fatalf("expected request body to contain the marshaled name field, got %s", gotBody)
	}
	if source.ID != 4 {
		t.Fatalf("expected decoded source ID 4, got %d", source.ID)
	}
}

func TestDoWithRetry_RetriesOn429(t *testing.T) {
	attempts := 0

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"success","data":{"id":1,"group_id":"a"}}`))
	})

	_, err := Get[testSource](context.Background(), c, "/sources/1", "source 1")
	if err != nil {
		t.Fatalf("expected retry to succeed, got error: %v", err)
	}
	if attempts != 2 {
		t.Fatalf("expected exactly 2 attempts, got %d", attempts)
	}
}

func TestParseRetryAfter_Seconds(t *testing.T) {
	delay, err := parseRetryAfter("5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if delay != 5*time.Second {
		t.Fatalf("expected 5s, got %v", delay)
	}
}

func TestParseRetryAfter_HTTPDate(t *testing.T) {
	future := time.Now().Add(10 * time.Second).UTC().Format(time.RFC1123)

	delay, err := parseRetryAfter(future)
	if err != nil {
		t.Fatalf("unexpected error parsing RFC1123 date: %v", err)
	}
	if delay <= 0 || delay > 11*time.Second {
		t.Fatalf("expected a delay close to 10s, got %v", delay)
	}
}

func TestParseRetryAfter_PastHTTPDate(t *testing.T) {
	past := time.Now().Add(-10 * time.Second).UTC().Format(time.RFC1123)

	delay, err := parseRetryAfter(past)
	if err != nil {
		t.Fatalf("unexpected error parsing past RFC1123 date: %v", err)
	}
	if delay != 0 {
		t.Fatalf("expected 0 delay for a date already in the past, got %v", delay)
	}
}

func TestParseRetryAfter_InvalidFormat(t *testing.T) {
	if _, err := parseRetryAfter("not-a-valid-value"); err == nil {
		t.Fatal("expected error for an unparseable Retry-After header")
	}
}

func TestRetryDelay_HonorsRetryAfterHeader(t *testing.T) {
	resp := &http.Response{Header: http.Header{"Retry-After": []string{"3"}}}
	deadline := time.Now().Add(retryBudgetTimeout)

	delay := retryDelay(resp, 1, deadline)

	if delay != 3*time.Second {
		t.Fatalf("expected Retry-After to be honored exactly, got %v", delay)
	}
}

func TestRetryDelay_FallsBackToBackoffOnInvalidHeader(t *testing.T) {
	// An unparseable Retry-After must not error or stall the retry loop — it
	// should silently fall through to the exponential backoff computation.
	resp := &http.Response{Header: http.Header{"Retry-After": []string{"not-a-valid-value"}}}
	deadline := time.Now().Add(retryBudgetTimeout)

	delay := retryDelay(resp, 1, deadline)

	if delay <= 0 {
		t.Fatalf("expected a positive backoff delay, got %v", delay)
	}
	if delay > maxRetryDelay {
		t.Fatalf("expected delay capped at maxRetryDelay, got %v", delay)
	}
}

func TestDelete(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"deleted"}`))
	})

	if err := Delete(context.Background(), c, "/sources/1"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestGetSourceGroupID(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sources/4" {
			t.Fatalf("expected /sources/4, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"success","data":{"id":4,"group_id":"wandering_aimlessly"}}`))
	})

	groupID, err := GetSourceGroupID(context.Background(), c, 4)
	if err != nil {
		t.Fatalf("GetSourceGroupID failed: %v", err)
	}
	if groupID != "wandering_aimlessly" {
		t.Fatalf("expected group_id wandering_aimlessly, got %q", groupID)
	}
}

func TestGetSourceGroupID_MissingGroupID(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"success","data":{"id":4,"group_id":""}}`))
	})

	if _, err := GetSourceGroupID(context.Background(), c, 4); err == nil {
		t.Fatal("expected error when group_id is empty")
	}
}
