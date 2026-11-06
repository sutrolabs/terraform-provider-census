package v2

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"
)

// sourceResponseJSON is a fetchSource response built from the examples in
// apidocs/v2/components/sources.yaml#/Source, including fields the provider
// doesn't model (last_test_*, connection_details) to confirm they're ignored
// rather than breaking the decode.
const sourceResponseJSON = `{
  "status": "success",
  "data": {
    "id": 4,
    "group_id": "wandering_aimlessly",
    "name": "Acme Inc. Snowflake",
    "type": "snowflake",
    "sync_engine": "advanced",
    "created_at": "2023-07-22T23:42:47.239Z",
    "last_test_results": {"steps": [{"step": "Verify connectivity", "status": "success", "error": {"message": null}}]},
    "last_test_succeeded": true,
    "last_tested_at": null,
    "connection_details": {"account": "xxxxxxx.us-east-1", "user": "DEV", "warehouse": "TEST", "use_keypair": false},
    "warehouse_writeback_retention_in_days": 7,
    "region": "AWS-US-EAST-1",
    "storage_backend": {"type": "s3", "bucket_name": "my-census-bucket", "role_arn": "arn:aws:iam::123456789012:role/census-role", "external_id": "census-external-id"},
    "notification_emails": ["notifications@mycompany.slack.com", "alerts@mycompany.com"]
  }
}`

func TestSourceUnmarshal_SpecExample(t *testing.T) {
	var env envelope[*Source]
	if err := json.Unmarshal([]byte(sourceResponseJSON), &env); err != nil {
		t.Fatalf("failed to unmarshal spec example: %v", err)
	}
	source := env.Data

	retention := 7
	want := &Source{
		ID:                                4,
		Name:                              "Acme Inc. Snowflake",
		Type:                              "snowflake",
		GroupID:                           "wandering_aimlessly",
		SyncEngine:                        "advanced",
		WarehouseWritebackRetentionInDays: &retention,
		Region:                            "AWS-US-EAST-1",
		NotificationEmails:                []string{"notifications@mycompany.slack.com", "alerts@mycompany.com"},
		StorageBackend: map[string]interface{}{
			"type":        "s3",
			"bucket_name": "my-census-bucket",
			"role_arn":    "arn:aws:iam::123456789012:role/census-role",
			"external_id": "census-external-id",
		},
		CreatedAt: time.Date(2023, 7, 22, 23, 42, 47, 239000000, time.UTC),
	}
	if !reflect.DeepEqual(source, want) {
		t.Fatalf("unexpected source:\n got: %#v\nwant: %#v", source, want)
	}
}

func TestSourceUnmarshal_StorageBackendTypes(t *testing.T) {
	// One example per storage backend from apidocs/v2/components/sources.yaml
	// (S3StorageBackend, GCSStorageBackend, AzureStorageBackend). The response
	// types storage_backend loosely (additionalProperties), so each shape must
	// survive as-is.
	testCases := map[string]string{
		"s3":    `{"type":"s3","bucket_name":"my-census-bucket","role_arn":"arn:aws:iam::123456789012:role/census-role","external_id":"census-external-id"}`,
		"gcs":   `{"type":"gcs","bucket_name":"my-census-bucket","project_id":"my-gcp-project","service_account_email":"census@my-project.iam.gserviceaccount.com"}`,
		"azure": `{"type":"azure","storage_account_name":"mycensusstorageaccount","storage_container_name":"census-container"}`,
	}

	for backendType, backend := range testCases {
		t.Run(backendType, func(t *testing.T) {
			var source Source
			if err := json.Unmarshal([]byte(`{"id":4,"storage_backend":`+backend+`}`), &source); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			var want map[string]interface{}
			if err := json.Unmarshal([]byte(backend), &want); err != nil {
				t.Fatalf("bad fixture: %v", err)
			}
			if !reflect.DeepEqual(source.StorageBackend, want) {
				t.Fatalf("unexpected storage_backend:\n got: %#v\nwant: %#v", source.StorageBackend, want)
			}
		})
	}
}

func TestSourceUnmarshal_NullOptionalFields(t *testing.T) {
	var source Source
	err := json.Unmarshal([]byte(`{"id":4,"group_id":"g","warehouse_writeback_retention_in_days":null,"storage_backend":null,"notification_emails":null}`), &source)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if source.WarehouseWritebackRetentionInDays != nil || source.StorageBackend != nil || source.NotificationEmails != nil {
		t.Fatalf("expected null optional fields to stay nil, got %#v", source)
	}
}

func TestSourceConnectionMarshal_OmitsUnsetOptionalFields(t *testing.T) {
	body, err := json.Marshal(CreateSourceRequest{Connection: SourceConnection{
		Name:        "Acme",
		Type:        "snowflake",
		Credentials: map[string]interface{}{"account": "acct"},
	}})
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var got map[string]map[string]interface{}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	connection, ok := got["connection"]
	if !ok || len(got) != 1 {
		t.Fatalf("expected the body to be wrapped in a single \"connection\" key, got %s", body)
	}
	for _, field := range []string{"sync_engine", "warehouse_writeback_retention_in_days", "region", "notification_emails", "storage_backend"} {
		if _, present := connection[field]; present {
			t.Errorf("expected unset %s to be omitted, got %s", field, body)
		}
	}
}

func TestSourceConnectionMarshal_IncludesSetOptionalFields(t *testing.T) {
	retention := 7
	region := "AWS-US-EAST-1"
	body, err := json.Marshal(UpdateSourceRequest{Connection: SourceConnection{
		Name:                              "Acme",
		WarehouseWritebackRetentionInDays: &retention,
		Region:                            &region,
		NotificationEmails:                []string{"alerts@mycompany.com"},
		StorageBackend:                    map[string]interface{}{"type": "gcs", "bucket_name": "b", "service_account_key": "{}"},
		Credentials:                       map[string]interface{}{"account": "acct"},
	}})
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	want := `{"connection":{"name":"Acme","warehouse_writeback_retention_in_days":7,"region":"AWS-US-EAST-1","notification_emails":["alerts@mycompany.com"],"storage_backend":{"bucket_name":"b","service_account_key":"{}","type":"gcs"},"credentials":{"account":"acct"}}}`
	if string(body) != want {
		t.Fatalf("unexpected request body:\n got: %s\nwant: %s", body, want)
	}
}

func TestSourceCRUD_PathsAndMethods(t *testing.T) {
	var calls []string

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/sources":
			body, _ := io.ReadAll(r.Body)
			if !strings.HasPrefix(string(body), `{"connection":{`) {
				t.Errorf("expected create body wrapped in connection, got %s", body)
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(sourceResponseJSON))
		case r.Method == http.MethodGet && r.URL.Path == "/sources/4",
			r.Method == http.MethodPatch && r.URL.Path == "/sources/4":
			_, _ = w.Write([]byte(sourceResponseJSON))
		case r.Method == http.MethodDelete && r.URL.Path == "/sources/4":
			_, _ = w.Write([]byte(`{"status":"deleted"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/sources/4/refresh_tables":
			// Bare, non-enveloped body per the spec.
			_, _ = w.Write([]byte(`{"refresh_key":1647978948}`))
		default:
			http.NotFound(w, r)
		}
	})
	ctx := context.Background()

	created, err := CreateSource(ctx, c, &CreateSourceRequest{Connection: SourceConnection{Name: "Acme", Type: "snowflake"}})
	if err != nil {
		t.Fatalf("CreateSource failed: %v", err)
	}
	if created.ID != 4 || created.GroupID != "wandering_aimlessly" {
		t.Fatalf("unexpected created source: %#v", created)
	}
	if _, err := GetSource(ctx, c, 4); err != nil {
		t.Fatalf("GetSource failed: %v", err)
	}
	if _, err := UpdateSource(ctx, c, 4, &UpdateSourceRequest{Connection: SourceConnection{Name: "Acme"}}); err != nil {
		t.Fatalf("UpdateSource failed: %v", err)
	}
	if err := RefreshSourceTables(ctx, c, 4); err != nil {
		t.Fatalf("RefreshSourceTables failed: %v", err)
	}
	if err := DeleteSource(ctx, c, 4); err != nil {
		t.Fatalf("DeleteSource failed: %v", err)
	}

	want := []string{"POST /sources", "GET /sources/4", "PATCH /sources/4", "POST /sources/4/refresh_tables", "DELETE /sources/4"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("unexpected calls:\n got: %v\nwant: %v", calls, want)
	}
}

func TestGetSource_NotFoundStaysDetectable(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	_, err := GetSource(context.Background(), c, 4)
	if !IsNotFoundError(err) {
		t.Fatalf("expected a wrapped 404 to still satisfy IsNotFoundError, got %v", err)
	}
}

func TestCreateSource_NullDataIsError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"status":"created","data":null}`))
	})

	_, err := CreateSource(context.Background(), c, &CreateSourceRequest{})
	if err == nil || !strings.Contains(err.Error(), "may have been created anyway") {
		t.Fatalf("expected an ambiguous-create error, got %v", err)
	}
}

func TestFieldRulesUnmarshal(t *testing.T) {
	testCases := map[string]FieldRules{
		`"required"`:                             {"required"},
		`["required", "required:notForEditing"]`: {"required", "required:notForEditing"},
		`""`:                                     nil,
		`null`:                                   nil,
	}

	for input, want := range testCases {
		var got struct {
			Rules FieldRules `json:"rules"`
		}
		if err := json.Unmarshal([]byte(`{"rules":`+input+`}`), &got); err != nil {
			t.Fatalf("failed to unmarshal %s: %v", input, err)
		}
		if !reflect.DeepEqual(got.Rules, want) {
			t.Errorf("rules %s: got %#v, want %#v", input, got.Rules, want)
		}
	}
}

func TestGetSourceType_UnknownType(t *testing.T) {
	var gotPath string
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		http.NotFound(w, r)
	})

	_, err := GetSourceType(context.Background(), c, "not_a_source")
	if err == nil || err.Error() != "unknown source type: not_a_source" {
		t.Fatalf("expected unknown source type error, got %v", err)
	}
	if gotPath != "/source_types/not_a_source" {
		t.Fatalf("expected fetch by service name, got path %q", gotPath)
	}
}
