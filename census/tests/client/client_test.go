package client_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sutrolabs/terraform-provider-census/census/client"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		config  *client.Config
		wantErr bool
	}{
		{
			name: "valid config with personal token",
			config: &client.Config{
				PersonalAccessToken: "test-token",
				BaseURL:             "https://api.test.com",
				Region:              "us",
			},
			wantErr: false,
		},
		{
			name: "valid config with workspace token",
			config: &client.Config{
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
			config: &client.Config{
				PersonalAccessToken: "test-token",
				Region:              "us",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiClient, err := client.NewClient(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && apiClient == nil {
				t.Error("NewClient() returned nil client when no error was expected")
			}
		})
	}
}

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		apiError *client.APIError
		expected string
	}{
		{
			name: "error with message",
			apiError: &client.APIError{
				StatusCode: 400,
				Message:    "Bad request",
			},
			expected: "Census API error (status 400): Bad request",
		},
		{
			name: "error without message",
			apiError: &client.APIError{
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
		opts     *client.ListOptions
		expected map[string]string
	}{
		{
			name: "all options set",
			opts: &client.ListOptions{
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
			opts: &client.ListOptions{
				Page: 1,
			},
			expected: map[string]string{
				"page": "1",
			},
		},
		{
			name:     "no options",
			opts:     &client.ListOptions{},
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

	apiClient, err := client.NewClient(&client.Config{
		PersonalAccessToken: "test-token",
		BaseURL:             server.URL,
		HTTPClient:          server.Client(),
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Note: makeRequest is not exported, so we can't test it directly from external package
	// This test would need to be adapted to test exported methods instead
	_ = apiClient

	t.Skip("Skipping test for unexported method - would need to test via exported methods")
}
