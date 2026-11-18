package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasetResponse() map[string]interface{} {
	return map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"id":                  157,
			"name":                "Users",
			"type":                "sql",
			"resource_identifier": "dataset:users",
			"description":         nil,
			"created_at":          "2023-07-22T23:42:47.239Z",
			"updated_at":          "2023-07-23T01:02:03.000Z",
			"cached_record_count": 809,
			"query":               "select id, email from users",
			"source_id":           123,
			"group_id":            "wandering_aimlessly",
			"columns": []interface{}{
				map[string]interface{}{"name": "id", "data_type": "integer", "can_be_upsert_key": true},
			},
		},
	}
}

func TestResourceDatasetCreate(t *testing.T) {
	t.Parallel()

	var calls []string
	var createBody map[string]interface{}
	apiClient := newActivationTestClients(t, func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/datasets":
			raw, _ := io.ReadAll(r.Body)
			json.Unmarshal(raw, &createBody)
			writeActivationJSON(t, w, http.StatusCreated, datasetResponse())
		case r.Method == http.MethodGet && r.URL.Path == "/datasets/157":
			writeActivationJSON(t, w, http.StatusOK, datasetResponse())
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	})

	d := schema.TestResourceDataRaw(t, resourceDataset().Schema, map[string]interface{}{
		"name":     "Users",
		"query":    "select id, email from users",
		"group_id": "wandering_aimlessly",
	})

	if diags := resourceDatasetCreate(context.Background(), d, apiClient); diags.HasError() {
		t.Fatalf("create failed: %#v", diags)
	}

	wantCalls := []string{"POST /datasets", "GET /datasets/157"}
	if !reflect.DeepEqual(calls, wantCalls) {
		t.Fatalf("unexpected calls:\n got: %v\nwant: %v", calls, wantCalls)
	}
	wantBody := map[string]interface{}{
		"name":     "Users",
		"type":     "sql",
		"query":    "select id, email from users",
		"group_id": "wandering_aimlessly",
	}
	if !reflect.DeepEqual(createBody, wantBody) {
		t.Fatalf("unexpected create body:\n got: %#v\nwant: %#v", createBody, wantBody)
	}

	want := map[string]string{
		"id":                  "157",
		"group_id":            "wandering_aimlessly",
		"resource_identifier": "dataset:users",
		"cached_record_count": "809",
		"created_at":          "2023-07-22T23:42:47Z",
		"updated_at":          "2023-07-23T01:02:03Z",
		"columns.#":           "1",
		"columns.0.name":      "id",
		"columns.0.data_type": "integer",
		"metadata_ready":      "true",
	}
	state := d.State().Attributes
	for k, v := range want {
		if state[k] != v {
			t.Errorf("%s: got %q, want %q", k, state[k], v)
		}
	}
}

func TestResourceDatasetDelete_NotFoundIsSuccess(t *testing.T) {
	t.Parallel()

	apiClient := newActivationTestClients(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/datasets/157" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNotFound)
	})

	d := schema.TestResourceDataRaw(t, resourceDataset().Schema, map[string]interface{}{})
	d.SetId("157")

	if diags := resourceDatasetDelete(context.Background(), d, apiClient); diags.HasError() {
		t.Fatalf("expected deleting a missing dataset to succeed, got %#v", diags)
	}
}
