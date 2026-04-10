package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/sutrolabs/terraform-provider-census/census/client"
)

func TestResourceSourceCreate_DefaultsSyncEngineToAdvanced(t *testing.T) {
	t.Parallel()

	capturedSyncEngine, diags, d := runSourceCreateTest(t, "")
	if diags.HasError() {
		t.Fatalf("expected source creation to succeed, got diagnostics: %#v", diags)
	}

	if capturedSyncEngine != "advanced" {
		t.Fatalf("expected create request to default to advanced sync engine, got %q", capturedSyncEngine)
	}

	if got := d.Get("sync_engine").(string); got != "advanced" {
		t.Fatalf("expected state to keep default advanced sync engine, got %q", got)
	}
}

func TestResourceSourceCreate_UsesConfiguredSyncEngine(t *testing.T) {
	t.Parallel()

	capturedSyncEngine, diags, d := runSourceCreateTest(t, "advanced")
	if diags.HasError() {
		t.Fatalf("expected source creation to succeed, got diagnostics: %#v", diags)
	}

	if capturedSyncEngine != "advanced" {
		t.Fatalf("expected create request to use advanced sync engine, got %q", capturedSyncEngine)
	}

	if got := d.Get("sync_engine").(string); got != "advanced" {
		t.Fatalf("expected state to retain advanced sync engine, got %q", got)
	}
}

func TestResourceSourceRead_UpdatesSyncEngineFromAPI(t *testing.T) {
	t.Parallel()

	const (
		workspaceID = 69962
		sourceID    = 2280673
	)

	apiClient := newSourceTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/workspaces/69962/api_key":
			writeJSON(t, w, http.StatusOK, map[string]string{"api_key": "workspace-token"})
		case "/sources/2280673":
			if r.Method != http.MethodGet {
				t.Fatalf("unexpected method for %s: %s", r.URL.Path, r.Method)
			}
			writeJSON(t, w, http.StatusOK, buildSourceResponse(sourceID, "advanced"))
		default:
			http.NotFound(w, r)
		}
	})

	d := schema.TestResourceDataRaw(t, resourceSource().Schema, map[string]interface{}{
		"workspace_id":      strconv.Itoa(workspaceID),
		"name":              "Warehouse Source",
		"type":              "snowflake",
		"connection_config": map[string]interface{}{"account": "acct"},
	})
	d.SetId(strconv.Itoa(sourceID))

	diags := resourceSourceRead(context.Background(), d, apiClient)
	if diags.HasError() {
		t.Fatalf("expected source read to succeed, got diagnostics: %#v", diags)
	}

	if got := d.Get("sync_engine").(string); got != "advanced" {
		t.Fatalf("expected read to update sync_engine from API response, got %q", got)
	}
}

func runSourceCreateTest(t *testing.T, syncEngine string) (string, diag.Diagnostics, *schema.ResourceData) {
	t.Helper()

	const (
		workspaceID = 69962
		sourceID    = 2280673
	)

	var capturedSyncEngine string

	apiClient := newSourceTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/workspaces/69962/api_key":
			writeJSON(t, w, http.StatusOK, map[string]string{"api_key": "workspace-token"})
		case "/source_types":
			writeJSON(t, w, http.StatusOK, map[string]interface{}{
				"status": "success",
				"data": []map[string]interface{}{
					{
						"service_name": "snowflake",
						"configuration_fields": map[string]interface{}{
							"fields": []map[string]interface{}{},
						},
					},
				},
			})
		case "/sources":
			if r.Method != http.MethodPost {
				t.Fatalf("unexpected method for %s: %s", r.URL.Path, r.Method)
			}

			var req client.CreateSourceRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("failed to decode create source request: %v", err)
			}
			capturedSyncEngine = req.Connection.SyncEngine

			writeJSON(t, w, http.StatusOK, buildSourceResponse(sourceID, req.Connection.SyncEngine))
		case "/sources/2280673":
			if r.Method != http.MethodGet {
				t.Fatalf("unexpected method for %s: %s", r.URL.Path, r.Method)
			}
			writeJSON(t, w, http.StatusOK, buildSourceResponse(sourceID, capturedSyncEngine))
		default:
			http.NotFound(w, r)
		}
	})

	raw := map[string]interface{}{
		"workspace_id":      strconv.Itoa(workspaceID),
		"name":              "Warehouse Source",
		"type":              "snowflake",
		"connection_config": map[string]interface{}{"account": "acct"},
	}
	if syncEngine != "" {
		raw["sync_engine"] = syncEngine
	}

	d := schema.TestResourceDataRaw(t, resourceSource().Schema, raw)
	diags := resourceSourceCreate(context.Background(), d, apiClient)

	return capturedSyncEngine, diags, d
}

func newSourceTestClient(t *testing.T, handler http.HandlerFunc) *client.Client {
	t.Helper()

	server := httptest.NewServer(handler)
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

func buildSourceResponse(sourceID int, syncEngine string) map[string]interface{} {
	return map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"id":          sourceID,
			"name":        "Warehouse Source",
			"type":        "snowflake",
			"sync_engine": syncEngine,
			"connection": map[string]interface{}{
				"sync_engine": syncEngine,
			},
			"status":      "connected",
			"test_status": "passed",
			"created_at":  "2026-04-09T22:00:00Z",
			"updated_at":  "2026-04-09T22:05:00Z",
		},
	}
}

func writeJSON(t *testing.T, w http.ResponseWriter, statusCode int, payload interface{}) {
	t.Helper()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		t.Fatalf("failed to encode response: %v", err)
	}
}
