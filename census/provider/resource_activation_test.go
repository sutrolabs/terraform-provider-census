package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/sutrolabs/terraform-provider-census/census/client/v2"
)

func newActivationTestClients(t *testing.T, handler http.HandlerFunc) *clients {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	v2Client, err := v2.NewClient(&v2.Config{PersonalAccessToken: "test-token", BaseURL: server.URL})
	if err != nil {
		t.Fatalf("failed to create v2 client: %v", err)
	}
	return &clients{V2: v2Client}
}

// activationResponse is a GET /activations/{activation_id} response.
func activationResponse(id, name string) map[string]interface{} {
	return map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"id":       id,
			"name":     name,
			"group_id": "abundant_giraffe",
			"source":   map[string]interface{}{"id": 42, "name": "Warehouse", "type": "snowflake"},
			"destination": map[string]interface{}{
				"id":                         123,
				"type":                       "salesforce",
				"created_at":                 "2023-07-22T23:42:47.239Z",
				"last_test_succeeded":        false,
				"last_tested_at":             "2023-07-22T23:45:12Z",
				"is_file_system_destination": true,
				"connection_details":         map[string]interface{}{"account_id": 7515011393, "account_name": "Manager Account", "sandbox": true},
				"last_test_results": map[string]interface{}{"steps": []interface{}{
					map[string]interface{}{"step": "Validate Credentials", "status": "error", "error": map[string]interface{}{"message": "Invalid API credentials"}},
				}},
			},
		},
	}
}

func writeActivationJSON(t *testing.T, w http.ResponseWriter, status int, payload interface{}) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		t.Fatalf("failed to encode response: %v", err)
	}
}

func TestResourceActivationSchema(t *testing.T) {
	t.Parallel()

	s := resourceActivation().Schema

	for _, key := range []string{"group_id", "type"} {
		if !s[key].Required || !s[key].ForceNew {
			t.Errorf("expected %s to be Required and ForceNew", key)
		}
	}
	if c := s["credentials"]; !c.Required || !c.Sensitive || c.ForceNew || c.Type != schema.TypeMap {
		t.Errorf("expected credentials to be a Required, Sensitive, updatable map")
	}
	if n := s["name"]; !n.Optional || !n.Computed || n.ForceNew {
		t.Errorf("expected name to be Optional + Computed and updatable")
	}
	if id := s["id"]; id.Type != schema.TypeString || !id.Computed {
		t.Errorf("expected id to be a computed string")
	}
	for _, key := range []string{"last_test_succeeded", "last_tested_at", "last_test_results", "connection_details", "is_file_system_activation"} {
		if s[key] == nil || !s[key].Computed || s[key].Optional || s[key].Required {
			t.Errorf("expected %s to be computed-only", key)
		}
	}
}

func TestProviderSchemaIsValid(t *testing.T) {
	t.Parallel()

	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("provider schema is invalid: %v", err)
	}
}

func TestResourceActivationDiff_ForceNew(t *testing.T) {
	t.Parallel()

	state := &terraform.InstanceState{
		ID: "glimmering_donkey",
		Attributes: map[string]string{
			"id":                "glimmering_donkey",
			"group_id":          "abundant_giraffe",
			"type":              "salesforce",
			"name":              "Salesforce",
			"credentials.%":     "1",
			"credentials.token": "secret",
		},
	}

	testCases := map[string]struct {
		change      map[string]interface{}
		requiresNew bool
	}{
		"group_id change replaces":   {change: map[string]interface{}{"group_id": "other_group"}, requiresNew: true},
		"type change replaces":       {change: map[string]interface{}{"type": "hubspot"}, requiresNew: true},
		"name change updates":        {change: map[string]interface{}{"name": "Renamed"}, requiresNew: false},
		"credentials change updates": {change: map[string]interface{}{"credentials": map[string]interface{}{"token": "rotated"}}, requiresNew: false},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			config := map[string]interface{}{
				"group_id":    "abundant_giraffe",
				"type":        "salesforce",
				"name":        "Salesforce",
				"credentials": map[string]interface{}{"token": "secret"},
			}
			for k, v := range tc.change {
				config[k] = v
			}

			diff, err := resourceActivation().Diff(context.Background(), state, terraform.NewResourceConfigRaw(config), nil)
			if err != nil {
				t.Fatalf("Diff failed: %v", err)
			}
			if diff == nil || diff.Empty() {
				t.Fatal("expected a non-empty diff")
			}
			if diff.RequiresNew() != tc.requiresNew {
				t.Fatalf("expected RequiresNew=%v, got %v", tc.requiresNew, diff.RequiresNew())
			}
		})
	}
}

func TestParseActivationImportID(t *testing.T) {
	t.Parallel()

	group, dest, err := parseActivationImportID("abundant_giraffe:123")
	if err != nil || group != "abundant_giraffe" || dest != 123 {
		t.Fatalf("expected (abundant_giraffe, 123, nil), got (%q, %d, %v)", group, dest, err)
	}

	for _, bad := range []string{"abundant_giraffe:", ":123", "abundant_giraffe:123:7", "abundant_giraffe:abc", "abundant_giraffe:0", "abundant_giraffe:-1"} {
		if _, _, err := parseActivationImportID(bad); err == nil {
			t.Errorf("expected %q to be rejected", bad)
		}
	}
}

func TestResourceActivationCreate_SendsRequestAndReadsBack(t *testing.T) {
	t.Parallel()

	var createBody map[string]interface{}
	apiClient := newActivationTestClients(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/activations":
			raw, _ := io.ReadAll(r.Body)
			json.Unmarshal(raw, &createBody)
			writeActivationJSON(t, w, http.StatusCreated, activationResponse("glimmering_donkey", "Salesforce"))
		case r.Method == http.MethodGet && r.URL.Path == "/activations/glimmering_donkey":
			writeActivationJSON(t, w, http.StatusOK, activationResponse("glimmering_donkey", "Salesforce"))
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	})

	d := schema.TestResourceDataRaw(t, resourceActivation().Schema, map[string]interface{}{
		"group_id":    "abundant_giraffe",
		"type":        "salesforce",
		"credentials": map[string]interface{}{"instance_url": "https://example.my.salesforce.com", "options": `{"sandbox":true}`},
	})

	if diags := resourceActivationCreate(context.Background(), d, apiClient); diags.HasError() {
		t.Fatalf("create failed: %#v", diags)
	}

	wantBody := map[string]interface{}{"activation": map[string]interface{}{
		"group_id": "abundant_giraffe",
		"type":     "salesforce",
		"credentials": map[string]interface{}{
			"instance_url": "https://example.my.salesforce.com",
			"options":      map[string]interface{}{"sandbox": true},
		},
	}}
	if !reflect.DeepEqual(createBody, wantBody) {
		t.Fatalf("unexpected create body:\n got: %#v\nwant: %#v", createBody, wantBody)
	}

	if d.Id() != "glimmering_donkey" {
		t.Fatalf("expected the string activation ID to be used as-is, got %q", d.Id())
	}
	assertActivationFlattened(t, d)
}

func assertActivationFlattened(t *testing.T, d *schema.ResourceData) {
	t.Helper()

	checks := map[string]interface{}{
		"name":                      "Salesforce",
		"group_id":                  "abundant_giraffe",
		"type":                      "salesforce",
		"last_test_succeeded":       false,
		"last_tested_at":            "2023-07-22T23:45:12Z",
		"is_file_system_activation": true,
		"connection_details": map[string]interface{}{
			"account_id":   "7515011393",
			"account_name": "Manager Account",
			"sandbox":      "true",
		},
		"last_test_results": []interface{}{map[string]interface{}{
			"step":          "Validate Credentials",
			"status":        "error",
			"error_message": "Invalid API credentials",
		}},
	}
	for key, want := range checks {
		if got := d.Get(key); !reflect.DeepEqual(got, want) {
			t.Errorf("unexpected %s: got %#v, want %#v", key, got, want)
		}
	}
}

func TestResourceActivationRead_NotFoundClearsID(t *testing.T) {
	t.Parallel()

	apiClient := newActivationTestClients(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	d := schema.TestResourceDataRaw(t, resourceActivation().Schema, map[string]interface{}{})
	d.SetId("glimmering_donkey")

	if diags := resourceActivationRead(context.Background(), d, apiClient); diags.HasError() {
		t.Fatalf("expected a 404 to be handled, got %#v", diags)
	}
	if d.Id() != "" {
		t.Fatalf("expected ID to be cleared, got %q", d.Id())
	}
}

func TestResourceActivationRead_ServerErrorKeepsID(t *testing.T) {
	t.Parallel()

	apiClient := newActivationTestClients(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusGatewayTimeout)
	})

	d := schema.TestResourceDataRaw(t, resourceActivation().Schema, map[string]interface{}{})
	d.SetId("glimmering_donkey")

	if diags := resourceActivationRead(context.Background(), d, apiClient); !diags.HasError() {
		t.Fatal("expected an error")
	}
	if d.Id() != "glimmering_donkey" {
		t.Fatalf("expected ID to be preserved, got %q", d.Id())
	}
}

func TestResourceActivationRead_PreservesCredentials(t *testing.T) {
	t.Parallel()

	apiClient := newActivationTestClients(t, func(w http.ResponseWriter, r *http.Request) {
		writeActivationJSON(t, w, http.StatusOK, activationResponse("glimmering_donkey", "Salesforce"))
	})

	d := schema.TestResourceDataRaw(t, resourceActivation().Schema, map[string]interface{}{
		"credentials": map[string]interface{}{"token": "secret"},
	})
	d.SetId("glimmering_donkey")

	if diags := resourceActivationRead(context.Background(), d, apiClient); diags.HasError() {
		t.Fatalf("read failed: %#v", diags)
	}
	if got := d.Get("credentials").(map[string]interface{}); got["token"] != "secret" {
		t.Fatalf("expected credentials to be left as configured, got %#v", got)
	}
}

// activationUpdateData builds the ResourceData Terraform passes to Update:
// prior state plus the planned changes from config.
func activationUpdateData(t *testing.T, config map[string]interface{}) *schema.ResourceData {
	t.Helper()

	r := resourceActivation()
	state := &terraform.InstanceState{
		ID: "glimmering_donkey",
		Attributes: map[string]string{
			"id":                "glimmering_donkey",
			"group_id":          "abundant_giraffe",
			"type":              "salesforce",
			"name":              "Salesforce",
			"credentials.%":     "1",
			"credentials.token": "secret",
		},
	}
	diff, err := r.Diff(context.Background(), state, terraform.NewResourceConfigRaw(config), nil)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	d, err := schema.InternalMap(r.Schema).Data(state, diff)
	if err != nil {
		t.Fatalf("failed to build ResourceData: %v", err)
	}
	return d
}

func TestResourceActivationUpdate_RequestBody(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		config map[string]interface{}
		want   map[string]interface{}
	}{
		// type is never sent (the API rejects it on update) and neither is
		// group_id (the API ignores it); unchanged credentials aren't re-sent.
		"rename only": {
			config: map[string]interface{}{"group_id": "abundant_giraffe", "type": "salesforce", "name": "Renamed", "credentials": map[string]interface{}{"token": "secret"}},
			want:   map[string]interface{}{"name": "Renamed"},
		},
		"credentials change": {
			config: map[string]interface{}{"group_id": "abundant_giraffe", "type": "salesforce", "name": "Salesforce", "credentials": map[string]interface{}{"token": "rotated"}},
			want:   map[string]interface{}{"name": "Salesforce", "credentials": map[string]interface{}{"token": "rotated"}},
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var body map[string]interface{}
			apiClient := newActivationTestClients(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodPatch {
					raw, _ := io.ReadAll(r.Body)
					json.Unmarshal(raw, &body)
				}
				writeActivationJSON(t, w, http.StatusOK, activationResponse("glimmering_donkey", "Salesforce"))
			})

			if diags := resourceActivationUpdate(context.Background(), activationUpdateData(t, tc.config), apiClient); diags.HasError() {
				t.Fatalf("update failed: %#v", diags)
			}
			want := map[string]interface{}{"activation": tc.want}
			if !reflect.DeepEqual(body, want) {
				t.Fatalf("unexpected PATCH body:\n got: %#v\nwant: %#v", body, want)
			}
		})
	}
}

func TestResourceActivationDiff_AfterImportShowsCredentialsUpdate(t *testing.T) {
	t.Parallel()

	// State as left by import + read: everything but credentials.
	state := &terraform.InstanceState{
		ID: "glimmering_donkey",
		Attributes: map[string]string{
			"id":       "glimmering_donkey",
			"group_id": "abundant_giraffe",
			"type":     "salesforce",
			"name":     "Salesforce",
		},
	}
	config := terraform.NewResourceConfigRaw(map[string]interface{}{
		"group_id":    "abundant_giraffe",
		"type":        "salesforce",
		"credentials": map[string]interface{}{"token": "secret"},
	})

	diff, err := resourceActivation().Diff(context.Background(), state, config, nil)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if diff == nil || diff.Attributes["credentials.token"] == nil {
		t.Fatalf("expected an in-place credentials update, got %#v", diff)
	}
	if diff.RequiresNew() {
		t.Fatal("expected the credentials update not to replace the activation")
	}
}

func TestExpandActivationCredentials(t *testing.T) {
	t.Parallel()

	got := expandActivationCredentials(map[string]interface{}{
		"api_key": "123456",
		"flag":    "true",
		"options": `{"sandbox":true}`,
		"list":    `["a","b"]`,
	})
	want := map[string]interface{}{
		"api_key": "123456",
		"flag":    "true",
		"options": map[string]interface{}{"sandbox": true},
		"list":    []interface{}{"a", "b"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestResourceActivationDelete_NotFoundIsSuccess(t *testing.T) {
	t.Parallel()

	apiClient := newActivationTestClients(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("unexpected request %s", r.Method)
		}
		w.WriteHeader(http.StatusNotFound)
	})

	d := schema.TestResourceDataRaw(t, resourceActivation().Schema, map[string]interface{}{})
	d.SetId("glimmering_donkey")

	if diags := resourceActivationDelete(context.Background(), d, apiClient); diags.HasError() {
		t.Fatalf("expected deleting a missing activation to succeed, got %#v", diags)
	}
}

// legacyPairServer lists one activation for any filter and serves it by ID.
func legacyPairServer(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/activations":
			writeActivationJSON(t, w, http.StatusOK, map[string]interface{}{
				"status": "success",
				"data": []interface{}{
					activationResponse("glimmering_donkey", "Salesforce")["data"],
				},
				"pagination": map[string]interface{}{"page": 1, "next_page": nil},
			})
		case "/activations/glimmering_donkey":
			writeActivationJSON(t, w, http.StatusOK, activationResponse("glimmering_donkey", "Salesforce"))
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.String())
			http.NotFound(w, r)
		}
	}
}

func TestResourceActivationImport(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		importID string
		wantID   string
		wantErr  string
	}{
		"activation ID passes through": {importID: "glimmering_donkey", wantID: "glimmering_donkey"},
		"legacy pair resolves":         {importID: "abundant_giraffe:123", wantID: "glimmering_donkey"},
		"malformed pair":               {importID: "abundant_giraffe:abc", wantErr: "destination_id must be a positive integer"},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			apiClient := newActivationTestClients(t, legacyPairServer(t))
			d := schema.TestResourceDataRaw(t, resourceActivation().Schema, map[string]interface{}{})
			d.SetId(tc.importID)

			result, err := resourceActivationImport(context.Background(), d, apiClient)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("expected error containing %q, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("import failed: %v", err)
			}
			if len(result) != 1 || result[0].Id() != tc.wantID {
				t.Fatalf("expected ID %q, got %#v", tc.wantID, result)
			}
		})
	}
}

func TestDataSourceActivationRead(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]interface{}{
		"by ID":          {"id": "glimmering_donkey"},
		"by legacy pair": {"group_id": "abundant_giraffe", "destination_id": 123},
	}

	for name, config := range testCases {
		config := config
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			apiClient := newActivationTestClients(t, legacyPairServer(t))
			d := schema.TestResourceDataRaw(t, dataSourceActivation().Schema, config)

			if diags := dataSourceActivationRead(context.Background(), d, apiClient); diags.HasError() {
				t.Fatalf("read failed: %#v", diags)
			}
			if d.Id() != "glimmering_donkey" {
				t.Fatalf("unexpected ID %q", d.Id())
			}
			assertActivationFlattened(t, d)
		})
	}
}

func TestFlattenConnectionDetails(t *testing.T) {
	t.Parallel()

	got := flattenConnectionDetails(map[string]interface{}{
		"string": "value",
		"number": float64(3),
		"nested": map[string]interface{}{"a": "b"},
		"null":   nil,
	})
	want := map[string]interface{}{"string": "value", "number": "3", "nested": `{"a":"b"}`}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}
