package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMakeRequest_UserAgentHeader(t *testing.T) {
	tests := []struct {
		name      string
		userAgent string
		want      string
	}{
		{"unset falls back to bare string", "", "terraform-provider-census"},
		{"versioned", UserAgent("0.2.20", "1.9.0"), "terraform-provider-census/0.2.20 (terraform/1.9.0)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
				UserAgent:           tt.userAgent,
			})
			if err != nil {
				t.Fatalf("NewClient: %v", err)
			}
			if _, err := c.makeRequest(context.Background(), http.MethodGet, "/ping", nil, TokenTypePersonal); err != nil {
				t.Fatalf("makeRequest: %v", err)
			}
			if got != tt.want {
				t.Errorf("User-Agent = %q, want %q", got, tt.want)
			}
		})
	}
}
