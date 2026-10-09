package v2

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
