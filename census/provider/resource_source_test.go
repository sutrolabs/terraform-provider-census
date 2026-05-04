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

	if got := resourceSource().Schema["sync_engine"].Default; got != nil {
		t.Fatalf("expected sync_engine schema default to be unset, got %#v", got)
	}

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

func TestResourceSourceCreate_UsesBasicSyncEngineWhenConfigured(t *testing.T) {
	t.Parallel()

	capturedSyncEngine, diags, d := runSourceCreateTest(t, "basic")
	if diags.HasError() {
		t.Fatalf("expected source creation to succeed, got diagnostics: %#v", diags)
	}

	if capturedSyncEngine != "basic" {
		t.Fatalf("expected create request to use basic sync engine, got %q", capturedSyncEngine)
	}

	if got := d.Get("sync_engine").(string); got != "basic" {
		t.Fatalf("expected state to retain basic sync engine, got %q", got)
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

func TestResourceSourceSchema_SyncEngineIsForceNew(t *testing.T) {
	t.Parallel()

	syncEngineField := resourceSource().Schema["sync_engine"]
	if syncEngineField == nil {
		t.Fatal("expected sync_engine field to exist on census_source")
	}

	if !syncEngineField.ForceNew {
		t.Fatal("expected sync_engine to be ForceNew because the Census API does not allow in-place sync engine changes")
	}

	if !syncEngineField.Computed {
		t.Fatal("expected sync_engine to be Computed so existing sources can preserve their current engine when the argument is omitted")
	}
}

func TestResourceSourceCreate_OmitsWarehouseWritebackByDefault(t *testing.T) {
	t.Parallel()

	captured, diags, d := runSourceCreateRequestTest(t, nil)
	if diags.HasError() {
		t.Fatalf("expected source creation to succeed, got diagnostics: %#v", diags)
	}

	if captured.WarehouseWritebackRetentionInDays != nil {
		t.Fatalf("expected warehouse_writeback_retention_in_days to be omitted, got %v", *captured.WarehouseWritebackRetentionInDays)
	}

	if got, ok := d.GetOk("warehouse_writeback_retention_in_days"); ok {
		t.Fatalf("expected state to omit warehouse_writeback_retention_in_days, got %v", got)
	}
}

func TestResourceSourceCreate_SendsWarehouseWritebackRetention(t *testing.T) {
	t.Parallel()

	retention := 30
	captured, diags, d := runSourceCreateRequestTest(t, &retention)
	if diags.HasError() {
		t.Fatalf("expected source creation to succeed, got diagnostics: %#v", diags)
	}

	if captured.WarehouseWritebackRetentionInDays == nil {
		t.Fatalf("expected create request to include warehouse_writeback_retention_in_days, got nil")
	}
	if *captured.WarehouseWritebackRetentionInDays != retention {
		t.Fatalf("expected create request retention %d, got %d", retention, *captured.WarehouseWritebackRetentionInDays)
	}

	if got := d.Get("warehouse_writeback_retention_in_days").(int); got != retention {
		t.Fatalf("expected state retention %d, got %d", retention, got)
	}
}

func TestResourceSourceRead_PopulatesWarehouseWritebackFromAPI(t *testing.T) {
	t.Parallel()

	const (
		workspaceID = 69962
		sourceID    = 2280673
		retention   = 45
	)

	apiClient := newSourceTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/workspaces/69962/api_key":
			writeJSON(t, w, http.StatusOK, map[string]string{"api_key": "workspace-token"})
		case "/sources/2280673":
			if r.Method != http.MethodGet {
				t.Fatalf("unexpected method for %s: %s", r.URL.Path, r.Method)
			}
			resp := buildSourceResponse(sourceID, "advanced")
			data := resp["data"].(map[string]interface{})
			data["warehouse_writeback_retention_in_days"] = retention
			writeJSON(t, w, http.StatusOK, resp)
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

	if got := d.Get("warehouse_writeback_retention_in_days").(int); got != retention {
		t.Fatalf("expected read to set warehouse_writeback_retention_in_days to %d, got %d", retention, got)
	}
}

func TestResourceSourceSchema_WarehouseWritebackIsUpdateable(t *testing.T) {
	t.Parallel()

	field := resourceSource().Schema["warehouse_writeback_retention_in_days"]
	if field == nil {
		t.Fatal("expected warehouse_writeback_retention_in_days field to exist on census_source")
	}
	if field.ForceNew {
		t.Fatal("expected warehouse_writeback_retention_in_days to be updateable in place (not ForceNew); the Census source update endpoint accepts changes to retention")
	}
	if field.Type != schema.TypeInt {
		t.Fatalf("expected warehouse_writeback_retention_in_days to be TypeInt, got %v", field.Type)
	}
}

func TestGetConfiguredWarehouseWritebackRetention(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		configured interface{}
		want       *int
	}{
		{name: "omitted returns nil"},
		{name: "zero returns nil", configured: 0},
		{name: "negative returns nil", configured: -1},
		{name: "positive returns pointer", configured: 30, want: intPtr(30)},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			raw := map[string]interface{}{
				"workspace_id":      "69962",
				"name":              "Warehouse Source",
				"type":              "snowflake",
				"connection_config": map[string]interface{}{"account": "acct"},
			}
			if tc.configured != nil {
				raw["warehouse_writeback_retention_in_days"] = tc.configured
			}

			d := schema.TestResourceDataRaw(t, resourceSource().Schema, raw)
			got := getConfiguredWarehouseWritebackRetention(d)
			switch {
			case tc.want == nil && got == nil:
				// ok
			case tc.want == nil && got != nil:
				t.Fatalf("expected nil, got pointer to %d", *got)
			case tc.want != nil && got == nil:
				t.Fatalf("expected pointer to %d, got nil", *tc.want)
			case *tc.want != *got:
				t.Fatalf("expected %d, got %d", *tc.want, *got)
			}
		})
	}
}

func intPtr(v int) *int { return &v }

func TestGetConfiguredSourceSyncEngine(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		configured interface{}
		want       string
	}{
		{
			name: "omitted defaults to advanced on create",
			want: "advanced",
		},
		{
			name:       "explicit basic is preserved",
			configured: "basic",
			want:       "basic",
		},
		{
			name:       "explicit advanced is preserved",
			configured: "advanced",
			want:       "advanced",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			raw := map[string]interface{}{
				"workspace_id":      "69962",
				"name":              "Warehouse Source",
				"type":              "snowflake",
				"connection_config": map[string]interface{}{"account": "acct"},
			}
			if tc.configured != nil {
				raw["sync_engine"] = tc.configured
			}

			d := schema.TestResourceDataRaw(t, resourceSource().Schema, raw)
			if got := getConfiguredSourceSyncEngine(d); got != tc.want {
				t.Fatalf("expected configured sync engine %q, got %q", tc.want, got)
			}
		})
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

func runSourceCreateRequestTest(t *testing.T, retention *int) (client.SourceConnection, diag.Diagnostics, *schema.ResourceData) {
	t.Helper()

	const (
		workspaceID = 69962
		sourceID    = 2280673
	)

	var captured client.SourceConnection

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
			captured = req.Connection

			resp := buildSourceResponse(sourceID, req.Connection.SyncEngine)
			if req.Connection.WarehouseWritebackRetentionInDays != nil {
				resp["data"].(map[string]interface{})["warehouse_writeback_retention_in_days"] = *req.Connection.WarehouseWritebackRetentionInDays
			}
			writeJSON(t, w, http.StatusOK, resp)
		case "/sources/2280673":
			if r.Method != http.MethodGet {
				t.Fatalf("unexpected method for %s: %s", r.URL.Path, r.Method)
			}
			resp := buildSourceResponse(sourceID, captured.SyncEngine)
			if captured.WarehouseWritebackRetentionInDays != nil {
				resp["data"].(map[string]interface{})["warehouse_writeback_retention_in_days"] = *captured.WarehouseWritebackRetentionInDays
			}
			writeJSON(t, w, http.StatusOK, resp)
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
	if retention != nil {
		raw["warehouse_writeback_retention_in_days"] = *retention
	}

	d := schema.TestResourceDataRaw(t, resourceSource().Schema, raw)
	diags := resourceSourceCreate(context.Background(), d, apiClient)

	return captured, diags, d
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
