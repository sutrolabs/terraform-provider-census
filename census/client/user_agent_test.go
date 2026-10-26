package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMakeRequest_UserAgentHeader(t *testing.T) {
	const want = "terraform-provider-census/0.2.20"

	var got string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
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
	if _, err := c.makeRequest(context.Background(), http.MethodGet, "/ping", nil, TokenTypePersonal); err != nil {
		t.Fatalf("makeRequest: %v", err)
	}
	if got != want {
		t.Errorf("User-Agent = %q, want %q", got, want)
	}
}
