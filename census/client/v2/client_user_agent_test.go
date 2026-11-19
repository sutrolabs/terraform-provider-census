package v2

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGet_UserAgentHeader(t *testing.T) {
	const want = "terraform-provider-census/0.2.20"

	var got string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"success","data":{"id":1,"group_id":"a"}}`))
	}))
	defer server.Close()

	c, err := NewClient(&Config{
		PersonalAccessToken: "test-token",
		BaseURL:             server.URL,
		HTTPClient:          server.Client(),
		UserAgent:           want,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if _, err := Get[testSource](context.Background(), c, "/sources/1", "source 1"); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != want {
		t.Errorf("User-Agent = %q, want %q", got, want)
	}
}
