package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config with personal token",
			config: &Config{
				PersonalAccessToken: "test-token",
				BaseURL:             "https://api.test.com",
				Region:              "us",
			},
			wantErr: false,
		},
		{
			name: "valid config with workspace token",
			config: &Config{
				WorkspaceAccessToken: "test-token",
				BaseURL:              "https://api.test.com",
				Region:               "us",
			},
			wantErr: false,
		},
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name: "missing base URL",
			config: &Config{
				PersonalAccessToken: "test-token",
				Region:              "us",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("NewClient() returned nil client when no error was expected")
			}
		})
	}
}

func TestClient_buildURL(t *testing.T) {
	client := &Client{
		config: &Config{
			BaseURL: "https://api.test.com",
		},
	}

	tests := []struct {
		name     string
		path     string
		params   map[string]string
		expected string
	}{
		{
			name:     "path only",
			path:     "/workspaces",
			params:   nil,
			expected: "https://api.test.com/workspaces",
		},
		{
			name: "path with parameters",
			path: "/workspaces",
			params: map[string]string{
				"page":     "1",
				"per_page": "25",
			},
			expected: "https://api.test.com/workspaces?page=1&per_page=25",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.buildURL(tt.path, tt.params)
			if result != tt.expected {
				t.Errorf("buildURL() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		apiError *APIError
		expected string
	}{
		{
			name: "error with message",
			apiError: &APIError{
				StatusCode: 400,
				Message:    "Bad request",
			},
			expected: "Census API error (status 400): Bad request",
		},
		{
			name: "error without message",
			apiError: &APIError{
				StatusCode: 500,
			},
			expected: "Census API error (status 500)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.apiError.Error()
			if result != tt.expected {
				t.Errorf("APIError.Error() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestListOptions_ToParams(t *testing.T) {
	tests := []struct {
		name     string
		opts     *ListOptions
		expected map[string]string
	}{
		{
			name: "all options set",
			opts: &ListOptions{
				Page:    2,
				PerPage: 50,
				Order:   "asc",
			},
			expected: map[string]string{
				"page":     "2",
				"per_page": "50",
				"order":    "asc",
			},
		},
		{
			name: "partial options",
			opts: &ListOptions{
				Page: 1,
			},
			expected: map[string]string{
				"page": "1",
			},
		},
		{
			name:     "no options",
			opts:     &ListOptions{},
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.opts.ToParams()

			// Check that all expected keys are present with correct values
			for key, expectedValue := range tt.expected {
				if actualValue, exists := result[key]; !exists || actualValue != expectedValue {
					t.Errorf("ToParams()[%s] = %v, want %v", key, actualValue, expectedValue)
				}
			}

			// Check that no unexpected keys are present
			if len(result) != len(tt.expected) {
				t.Errorf("ToParams() returned %d params, want %d", len(result), len(tt.expected))
			}
		})
	}
}

func TestClient_makeRequest(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check headers
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type: application/json, got: %s", r.Header.Get("Content-Type"))
		}

		if r.Header.Get("User-Agent") != "terraform-provider-census" {
			t.Errorf("Expected User-Agent: terraform-provider-census, got: %s", r.Header.Get("User-Agent"))
		}

		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Expected Authorization: Bearer test-token, got: %s", r.Header.Get("Authorization"))
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "success"}`))
	}))
	defer server.Close()

	client := &Client{
		config: &Config{
			PersonalAccessToken: "test-token",
			BaseURL:             server.URL,
		},
		httpClient: server.Client(),
	}

	resp, err := client.makeRequest(context.Background(), http.MethodGet, "/test", nil, TokenTypePersonal)
	if err != nil {
		t.Fatalf("makeRequest() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}
}
