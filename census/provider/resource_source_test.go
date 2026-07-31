package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/sutrolabs/terraform-provider-census/census/client"
)

func TestResourceSourceSchema_WriteOnlyConnectionConfig(t *testing.T) {
	t.Parallel()

	schemaMap := resourceSource().Schema

	field := schemaMap["connection_config_wo"]
	if field == nil {
		t.Fatal("expected connection_config_wo field to exist on census_source")
	}
	if !field.WriteOnly {
		t.Fatal("expected connection_config_wo to be WriteOnly so its value is never persisted in state")
	}
	if !field.Sensitive {
		t.Fatal("expected connection_config_wo to be Sensitive")
	}
	if !field.Optional {
		t.Fatal("expected connection_config_wo to be Optional (backwards compatible; existing configs omit it)")
	}
	if field.Computed {
		t.Fatal("expected connection_config_wo not to be Computed (WriteOnly cannot be Computed)")
	}
	if field.ForceNew {
		t.Fatal("expected connection_config_wo not to be ForceNew (WriteOnly cannot be ForceNew)")
	}
	// SDKv2 forbids WriteOnly on TypeMap/TypeList/TypeSet, so the generic
	// write-only object is carried as a jsonencode-d string.
	if field.Type != schema.TypeString {
		t.Fatalf("expected connection_config_wo to be TypeString, got %v", field.Type)
	}
	if len(field.RequiredWith) != 1 || field.RequiredWith[0] != "connection_config_wo_version" {
		t.Fatalf("expected connection_config_wo to require connection_config_wo_version so updates are triggerable, got %#v", field.RequiredWith)
	}

	version := schemaMap["connection_config_wo_version"]
	if version == nil {
		t.Fatal("expected connection_config_wo_version field to exist on census_source")
	}
	if version.Type != schema.TypeInt {
		t.Fatalf("expected connection_config_wo_version to be TypeInt, got %v", version.Type)
	}
	if !version.Optional {
		t.Fatal("expected connection_config_wo_version to be Optional")
	}
	if version.WriteOnly {
		t.Fatal("expected connection_config_wo_version NOT to be WriteOnly; it must be tracked in state to detect changes")
	}
}

func TestResourceSourceSchema_ConnectionConfigStillRequired(t *testing.T) {
	t.Parallel()

	// Backwards compatibility: connection_config must remain a Required,
	// Sensitive map so existing configurations keep working unchanged.
	field := resourceSource().Schema["connection_config"]
	if field == nil {
		t.Fatal("expected connection_config field to exist on census_source")
	}
	if !field.Required {
		t.Fatal("expected connection_config to remain Required for backwards compatibility")
	}
	if !field.Sensitive {
		t.Fatal("expected connection_config to remain Sensitive")
	}
	if field.Type != schema.TypeMap {
		t.Fatalf("expected connection_config to remain TypeMap, got %v", field.Type)
	}
}

func TestApplyWriteOnlyCredentials_MergesAndOverridesConnectionConfig(t *testing.T) {
	t.Parallel()

	rawConfig := sourceRawConfig(map[string]cty.Value{
		"connection_config_wo": cty.StringVal(`{
			"password": "wo-password",
			"private_key_pkcs8": "wo-private-key",
			"private_key_passphrase": "wo-passphrase"
		}`),
	})

	credentials := map[string]interface{}{
		"account":  "acct",
		"password": "config-password", // should be overridden by write-only value
	}

	if err := applyWriteOnlyCredentialsFromConfig(rawConfig, credentials); err != nil {
		t.Fatalf("expected merge to succeed, got %v", err)
	}

	if got := credentials["account"]; got != "acct" {
		t.Fatalf("expected non-secret account to be preserved, got %v", got)
	}
	if got := credentials["password"]; got != "wo-password" {
		t.Fatalf("expected write-only password to override connection_config value, got %v", got)
	}
	if got := credentials["private_key_pkcs8"]; got != "wo-private-key" {
		t.Fatalf("expected private_key_pkcs8 to be populated from write-only value, got %v", got)
	}
	if got := credentials["private_key_passphrase"]; got != "wo-passphrase" {
		t.Fatalf("expected private_key_passphrase to be populated from write-only value, got %v", got)
	}
}

func TestApplyWriteOnlyCredentials_DecodesNestedJSONObject(t *testing.T) {
	t.Parallel()

	// A credential whose value is itself a JSON document (e.g. a BigQuery
	// service account) must be delivered to Census as a real object even when
	// it is supplied as a JSON string (jsonencode({ service_account_key = file(...) })).
	serviceAccount := `{"type":"service_account","project_id":"p"}`
	rawConfig := sourceRawConfig(map[string]cty.Value{
		"connection_config_wo": cty.StringVal(`{"service_account_key": ` + strconv.Quote(serviceAccount) + `}`),
	})

	credentials := map[string]interface{}{}
	if err := applyWriteOnlyCredentialsFromConfig(rawConfig, credentials); err != nil {
		t.Fatalf("expected merge to succeed, got %v", err)
	}

	obj, ok := credentials["service_account_key"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected service_account_key to be decoded into an object, got %T (%v)", credentials["service_account_key"], credentials["service_account_key"])
	}
	if obj["type"] != "service_account" || obj["project_id"] != "p" {
		t.Fatalf("unexpected decoded object: %#v", obj)
	}
}

func TestApplyWriteOnlyCredentials_DoesNotRetypeScalarSecrets(t *testing.T) {
	t.Parallel()

	// Opaque scalar secrets must be sent to Census exactly as written, even
	// when they happen to parse as a JSON number, bool, or null. Only nested
	// JSON objects/arrays are decoded.
	rawConfig := sourceRawConfig(map[string]cty.Value{
		"connection_config_wo": cty.StringVal(`{
			"numeric":  "123456",
			"boolish":  "true",
			"nullish":  "null",
			"leading0": "0123",
			"plain":    "s3cr3t"
		}`),
	})

	credentials := map[string]interface{}{}
	if err := applyWriteOnlyCredentialsFromConfig(rawConfig, credentials); err != nil {
		t.Fatalf("expected merge to succeed, got %v", err)
	}

	want := map[string]string{
		"numeric":  "123456",
		"boolish":  "true",
		"nullish":  "null",
		"leading0": "0123",
		"plain":    "s3cr3t",
	}
	for key, expected := range want {
		got, ok := credentials[key].(string)
		if !ok {
			t.Fatalf("expected %q to remain a string, got %T (%v)", key, credentials[key], credentials[key])
		}
		if got != expected {
			t.Fatalf("expected %q to stay %q, got %q", key, expected, got)
		}
	}
}

func TestApplyWriteOnlyCredentials_UnsetOrEmptyIsNoOp(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		attrs map[string]cty.Value
	}{
		{name: "absent"},
		{name: "null", attrs: map[string]cty.Value{"connection_config_wo": cty.NullVal(cty.String)}},
		{name: "unknown", attrs: map[string]cty.Value{"connection_config_wo": cty.UnknownVal(cty.String)}},
		{name: "empty string", attrs: map[string]cty.Value{"connection_config_wo": cty.StringVal("   ")}},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			credentials := map[string]interface{}{"account": "acct", "password": "config-password"}
			if err := applyWriteOnlyCredentialsFromConfig(sourceRawConfig(tc.attrs), credentials); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if credentials["password"] != "config-password" {
				t.Fatalf("expected connection_config password to be untouched, got %v", credentials["password"])
			}
		})
	}
}

func TestApplyWriteOnlyCredentials_InvalidJSONErrors(t *testing.T) {
	t.Parallel()

	rawConfig := sourceRawConfig(map[string]cty.Value{
		"connection_config_wo": cty.StringVal("not-json"),
	})

	err := applyWriteOnlyCredentialsFromConfig(rawConfig, map[string]interface{}{})
	if err == nil {
		t.Fatal("expected an error for non-JSON connection_config_wo")
	}
}

func TestApplyWriteOnlyCredentials_NullRawConfigIsNoOp(t *testing.T) {
	t.Parallel()

	rawType := resourceSource().CoreConfigSchema().ImpliedType()

	credentials := map[string]interface{}{"account": "acct"}
	if err := applyWriteOnlyCredentialsFromConfig(cty.NullVal(rawType), credentials); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(credentials) != 1 || credentials["account"] != "acct" {
		t.Fatalf("expected null raw config to leave credentials unchanged, got %#v", credentials)
	}
}

// sourceRawConfig builds a raw config cty.Value for the census_source schema,
// setting the given attributes and leaving the rest null. This mirrors what
// Terraform passes to the SDK so write-only values can be read from config.
func sourceRawConfig(attrs map[string]cty.Value) cty.Value {
	implied := resourceSource().CoreConfigSchema().ImpliedType()

	values := make(map[string]cty.Value, len(implied.AttributeTypes()))
	for name, attrType := range implied.AttributeTypes() {
		if v, ok := attrs[name]; ok {
			values[name] = v
			continue
		}
		values[name] = cty.NullVal(attrType)
	}

	return cty.ObjectVal(values)
}

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
	if !field.Computed {
		t.Fatal("expected warehouse_writeback_retention_in_days to be Computed so existing sources with writeback enabled outside Terraform do not show perpetual drift when the argument is omitted from configuration")
	}
}

// Regression: a source that already has Warehouse Writeback enabled outside
// Terraform (e.g. set via the Census UI before being imported) must not show
// drift when the user has not added warehouse_writeback_retention_in_days to
// their .tf. Without Computed:true on the schema this would loop forever as
// the read populates state, then plan diffs state vs zero-value config, then
// apply no-ops the API but refresh re-populates state, ad infinitum.
func TestResourceSourceRead_NoDriftWhenWritebackPresentInAPIButAbsentFromConfig(t *testing.T) {
	t.Parallel()

	const (
		workspaceID = 69962
		sourceID    = 2280673
		retention   = 30
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

	// Simulate a source already in state from a previous apply, with the
	// warehouse_writeback_retention_in_days argument NOT in the .tf.
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
		t.Fatalf("expected read to populate warehouse_writeback_retention_in_days from the API even when the argument is absent from configuration, got %d", got)
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
