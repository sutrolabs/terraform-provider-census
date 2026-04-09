package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/sutrolabs/terraform-provider-census/census/client"
)

func TestResourceReads_GatewayTimeoutPreservesState(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		resourcePath    string
		expectedSummary string
		buildData       func(t *testing.T) *schema.ResourceData
		read            func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics
	}{
		{
			name:            "destination",
			resourcePath:    "/destinations/2311185",
			expectedSummary: "failed to get destination",
			buildData:       testDestinationResourceData,
			read:            resourceDestinationRead,
		},
		{
			name:            "source",
			resourcePath:    "/sources/2280673",
			expectedSummary: "failed to get source",
			buildData:       testSourceResourceData,
			read:            resourceSourceRead,
		},
		{
			name:            "dataset",
			resourcePath:    "/datasets/44001",
			expectedSummary: "failed to get dataset",
			buildData:       testDatasetResourceData,
			read:            resourceDatasetRead,
		},
		{
			name:            "sync",
			resourcePath:    "/syncs/3503053",
			expectedSummary: "failed to get sync",
			buildData:       testSyncResourceData,
			read:            resourceSyncRead,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			apiClient := newReadTestClient(t, tc.resourcePath, http.StatusGatewayTimeout)
			d := tc.buildData(t)
			originalID := d.Id()

			diags := tc.read(context.Background(), d, apiClient)
			if !diags.HasError() {
				t.Fatalf("expected diagnostics for a 504 %s read", tc.name)
			}

			if d.Id() != originalID {
				t.Fatalf("expected %s ID to be preserved after 504, got %q", tc.name, d.Id())
			}

			if got := diags[0].Summary; !strings.Contains(got, tc.expectedSummary) {
				t.Fatalf("expected gateway timeout error summary containing %q, got %q", tc.expectedSummary, got)
			}
		})
	}
}

func TestResourceDestinationRead_NotFoundClearsState(t *testing.T) {
	t.Parallel()

	apiClient := newReadTestClient(t, "/destinations/2311185", http.StatusNotFound)
	d := testDestinationResourceData(t)

	diags := resourceDestinationRead(context.Background(), d, apiClient)
	if diags.HasError() {
		t.Fatalf("expected no diagnostics for a missing destination, got %#v", diags)
	}

	if d.Id() != "" {
		t.Fatalf("expected destination ID to be cleared after 404, got %q", d.Id())
	}
}

func newReadTestClient(t *testing.T, resourcePath string, statusCode int) *client.Client {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/workspaces/69962/api_key":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"api_key":"workspace-token"}`))
		case resourcePath:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(statusCode)
			_, _ = w.Write([]byte(`{"message":"request failed"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	apiClient, err := client.NewClient(&client.Config{
		BaseURL:             server.URL,
		PersonalAccessToken: "personal-token",
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	return apiClient
}

func testDestinationResourceData(t *testing.T) *schema.ResourceData {
	t.Helper()

	d := schema.TestResourceDataRaw(t, resourceDestination().Schema, map[string]interface{}{
		"workspace_id":      "69962",
		"name":              "Test Salesforce Destination",
		"type":              "salesforce",
		"connection_config": map[string]interface{}{"access_token": "secret"},
	})
	d.SetId("2311185")

	return d
}

func testSourceResourceData(t *testing.T) *schema.ResourceData {
	t.Helper()

	d := schema.TestResourceDataRaw(t, resourceSource().Schema, map[string]interface{}{
		"workspace_id":      "69962",
		"name":              "Test Source",
		"type":              "snowflake",
		"connection_config": map[string]interface{}{"account": "acct"},
	})
	d.SetId("2280673")

	return d
}

func testDatasetResourceData(t *testing.T) *schema.ResourceData {
	t.Helper()

	d := schema.TestResourceDataRaw(t, resourceDataset().Schema, map[string]interface{}{
		"workspace_id": "69962",
		"name":         "Test Dataset",
		"type":         "sql",
		"query":        "select 1",
		"source_id":    123,
	})
	d.SetId("44001")

	return d
}

func testSyncResourceData(t *testing.T) *schema.ResourceData {
	t.Helper()

	d := schema.TestResourceDataRaw(t, ResourceSync().Schema, map[string]interface{}{
		"workspace_id": "69962",
		"label":        "Test Sync",
		"operation":    "upsert",
	})
	d.SetId("3503053")

	return d
}
