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

// datasetResponseJSON is a fetchDataset response shaped like the SqlResponse
// schema in apidocs/v2/components/datasets.yaml, including a column field
// the provider doesn't model (can_be_upsert_key) to confirm it's ignored.
const datasetResponseJSON = `{
  "status": "success",
  "data": {
    "id": 157,
    "name": "Users",
    "type": "sql",
    "resource_identifier": "dataset:users",
    "description": "This dataset contains golden user records",
    "created_at": "2023-07-22T23:42:47.239Z",
    "updated_at": "2023-07-23T01:02:03.000Z",
    "cached_record_count": 809,
    "query": "SELECT id, name, email FROM users WHERE active = true",
    "source_id": 123,
    "group_id": "wandering_aimlessly",
    "columns": [
      {"name": "id", "data_type": "integer", "can_be_upsert_key": true},
      {"name": "email", "data_type": "character varying (256)", "can_be_upsert_key": false}
    ]
  }
}`

func TestDatasetUnmarshal_SpecExample(t *testing.T) {
	var env envelope[*Dataset]
	if err := json.Unmarshal([]byte(datasetResponseJSON), &env); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	description := "This dataset contains golden user records"
	count := 809
	want := &Dataset{
		ID:                 157,
		Name:               "Users",
		Type:               "sql",
		Description:        &description,
		Query:              "SELECT id, name, email FROM users WHERE active = true",
		GroupID:            "wandering_aimlessly",
		ResourceIdentifier: "dataset:users",
		CachedRecordCount:  &count,
		Columns: []DatasetColumn{
			{Name: "id", DataType: "integer"},
			{Name: "email", DataType: "character varying (256)"},
		},
		CreatedAt: time.Date(2023, 7, 22, 23, 42, 47, 239000000, time.UTC),
		UpdatedAt: time.Date(2023, 7, 23, 1, 2, 3, 0, time.UTC),
	}
	if !reflect.DeepEqual(env.Data, want) {
		t.Fatalf("unexpected dataset:\n got: %#v\nwant: %#v", env.Data, want)
	}
}

func TestDatasetUnmarshal_NullOptionalFields(t *testing.T) {
	var dataset Dataset
	if err := json.Unmarshal([]byte(`{"id":157,"description":null,"cached_record_count":null,"columns":[]}`), &dataset); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if dataset.Description != nil || dataset.CachedRecordCount != nil {
		t.Fatalf("expected null description and cached_record_count to stay nil, got %#v", dataset)
	}
}

func TestCreateDataset_SendsFlatBodyWithGroupID(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]interface{}

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(raw, &gotBody); err != nil {
			t.Fatalf("request body is not JSON: %v", err)
		}
		w.WriteHeader(http.StatusCreated)
		io.WriteString(w, datasetResponseJSON)
	})

	dataset, err := CreateDataset(context.Background(), c, &CreateDatasetRequest{
		Name:    "Users",
		Type:    "sql",
		Query:   "select 1",
		GroupID: "wandering_aimlessly",
	})
	if err != nil {
		t.Fatalf("CreateDataset failed: %v", err)
	}
	if dataset.ID != 157 {
		t.Fatalf("unexpected dataset ID %d", dataset.ID)
	}

	if gotMethod != http.MethodPost || gotPath != "/datasets" {
		t.Fatalf("unexpected request %s %s", gotMethod, gotPath)
	}
	want := map[string]interface{}{
		"name":     "Users",
		"type":     "sql",
		"query":    "select 1",
		"group_id": "wandering_aimlessly",
	}
	if !reflect.DeepEqual(gotBody, want) {
		t.Fatalf("unexpected body (an unset description must be omitted):\n got: %#v\nwant: %#v", gotBody, want)
	}
}

func TestCreateDataset_MissingIDIsError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		io.WriteString(w, `{"status":"created","data":null}`)
	})

	_, err := CreateDataset(context.Background(), c, &CreateDatasetRequest{})
	if err == nil || !strings.Contains(err.Error(), "may have been created anyway") {
		t.Fatalf("expected an ambiguous-create error, got %v", err)
	}
}

func TestUpdateDataset_SendsOnlySetFields(t *testing.T) {
	empty := ""
	query := "select 2"

	testCases := map[string]struct {
		req  UpdateDatasetRequest
		want map[string]interface{}
	}{
		"query only": {
			req:  UpdateDatasetRequest{Query: &query},
			want: map[string]interface{}{"query": "select 2"},
		},
		"cleared description is sent": {
			req:  UpdateDatasetRequest{Description: &empty},
			want: map[string]interface{}{"description": ""},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			var gotMethod, gotPath string
			var gotBody map[string]interface{}

			c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				gotMethod, gotPath = r.Method, r.URL.Path
				raw, _ := io.ReadAll(r.Body)
				if err := json.Unmarshal(raw, &gotBody); err != nil {
					t.Fatalf("request body is not JSON: %v", err)
				}
				io.WriteString(w, datasetResponseJSON)
			})

			if _, err := UpdateDataset(context.Background(), c, 157, &tc.req); err != nil {
				t.Fatalf("UpdateDataset failed: %v", err)
			}
			if gotMethod != http.MethodPatch || gotPath != "/datasets/157" {
				t.Fatalf("unexpected request %s %s", gotMethod, gotPath)
			}
			if !reflect.DeepEqual(gotBody, tc.want) {
				t.Fatalf("unexpected body:\n got: %#v\nwant: %#v", gotBody, tc.want)
			}
		})
	}
}

func TestGetDataset_NotFound(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	_, err := GetDataset(context.Background(), c, 157)
	if !IsNotFoundError(err) {
		t.Fatalf("expected a wrapped 404 to still satisfy IsNotFoundError, got %v", err)
	}
}

func TestDeleteDataset(t *testing.T) {
	var gotMethod, gotPath string
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		io.WriteString(w, `{"status":"deleted"}`)
	})

	if err := DeleteDataset(context.Background(), c, 157); err != nil {
		t.Fatalf("DeleteDataset failed: %v", err)
	}
	if gotMethod != http.MethodDelete || gotPath != "/datasets/157" {
		t.Fatalf("unexpected request %s %s", gotMethod, gotPath)
	}
}

func TestRefreshDatasetColumns_BareResponses(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/datasets/157/refresh_columns":
			w.WriteHeader(http.StatusAccepted)
			io.WriteString(w, `{"refresh_key":1647978948}`)
		case r.Method == http.MethodGet && r.URL.Path == "/datasets/157/refresh_columns_status":
			if got := r.URL.Query().Get("refresh_key"); got != "1647978948" {
				t.Errorf("unexpected refresh_key %q", got)
			}
			io.WriteString(w, `{"status":"completed"}`)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	})
	ctx := context.Background()

	key, err := RefreshDatasetColumns(ctx, c, 157)
	if err != nil {
		t.Fatalf("RefreshDatasetColumns failed: %v", err)
	}
	if key != 1647978948 {
		t.Fatalf("unexpected refresh key %d", key)
	}

	status, err := GetDatasetRefreshStatus(ctx, c, 157, key)
	if err != nil {
		t.Fatalf("GetDatasetRefreshStatus failed: %v", err)
	}
	if status != "completed" {
		t.Fatalf("unexpected status %q", status)
	}
}

func TestGetDatasetRefreshStatus_FailedRefreshIsError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"status":"error","message":"Failed to refresh source columns: relation \"users\" does not exist"}`)
	})

	_, err := GetDatasetRefreshStatus(context.Background(), c, 157, 1)
	if err == nil || !strings.Contains(err.Error(), `relation "users" does not exist`) {
		t.Fatalf("expected the refresh failure reason in the error, got %v", err)
	}
}
