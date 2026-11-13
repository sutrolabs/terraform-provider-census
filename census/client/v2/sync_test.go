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

// syncResponseJSON is a GET /syncs/{sync_id} response.
const syncResponseJSON = `{
  "status": "success",
  "data": {
    "id": 4545,
    "activation_id": "example_activation_id",
    "status": "up to date",
    "created_at": "2023-09-05T19:42:42.672Z",
    "updated_at": "2023-09-06T10:00:00.000Z",
    "label": "New Signups to Leads",
    "operation": "upsert",
    "source_attributes": {"object": {"type": "dataset", "id": 123}},
    "destination_attributes": {"object": "Contact"},
    "mappings": [
      {"from": {"type": "column", "data": "email"}, "to": "Email", "is_primary_identifier": true}
    ],
    "alert_attributes": [
      {"type": "FailureAlertConfiguration", "send_for": "first_time", "should_send_recovery": true, "options": {}}
    ],
    "mode": {"type": "triggered", "triggers": {"schedule": {"frequency": "daily", "hour": 10, "minute": 30}}},
    "paused": false,
    "field_behavior": "specific_properties"
  }
}`

func TestSyncUnmarshal_FullResponse(t *testing.T) {
	var env envelope[*Sync]
	if err := json.Unmarshal([]byte(syncResponseJSON), &env); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	createdAt := time.Date(2023, 9, 5, 19, 42, 42, 672000000, time.UTC)
	updatedAt := time.Date(2023, 9, 6, 10, 0, 0, 0, time.UTC)
	want := &Sync{
		ID:           4545,
		ActivationID: "example_activation_id",
		Status:       "up to date",
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
		Label:        "New Signups to Leads",
		Operation:    "upsert",
		SourceAttributes: map[string]interface{}{
			"object": map[string]interface{}{"type": "dataset", "id": float64(123)},
		},
		DestinationAttributes: &SyncDestinationAttributes{Object: "Contact"},
		Mappings: []MappingAttributes{
			{
				From:                MappingFrom{Type: "column", Data: "email"},
				To:                  "Email",
				IsPrimaryIdentifier: true,
			},
		},
		AlertAttributes: []AlertAttribute{
			{
				Type:               "FailureAlertConfiguration",
				SendFor:            "first_time",
				ShouldSendRecovery: true,
				Options:            map[string]interface{}{},
			},
		},
		Mode: &SyncMode{
			Type: "triggered",
			Triggers: &SyncTriggers{
				Schedule: &TriggerSchedule{Frequency: "daily", Hour: 10, Minute: 30},
			},
		},
		FieldBehavior: "specific_properties",
	}
	if !reflect.DeepEqual(env.Data, want) {
		t.Fatalf("unexpected sync:\n got: %#v\nwant: %#v", env.Data, want)
	}
}

func TestCreateSync_SendsBodyAndReturnsIDs(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]interface{}

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(raw, &gotBody); err != nil {
			t.Fatalf("request body is not JSON: %v", err)
		}
		w.WriteHeader(http.StatusCreated)
		io.WriteString(w, `{"status": "created", "data": {"sync_id": 4545, "activation_id": "example_activation_id"}}`)
	})

	sync, err := CreateSync(context.Background(), c, &CreateSyncRequest{
		ActivationID:          "example_activation_id",
		Operation:             "upsert",
		SourceAttributes:      map[string]interface{}{"object": map[string]interface{}{"type": "dataset", "id": 123}},
		DestinationAttributes: &SyncDestinationAttributes{Object: "Contact"},
		Mappings: []MappingAttributes{
			{From: MappingFrom{Type: "column", Data: "email"}, To: "Email", IsPrimaryIdentifier: true},
		},
	})
	if err != nil {
		t.Fatalf("CreateSync failed: %v", err)
	}
	if sync.ID != 4545 || sync.ActivationID != "example_activation_id" {
		t.Fatalf("unexpected sync %+v", sync)
	}

	if gotMethod != http.MethodPost || gotPath != "/syncs" {
		t.Fatalf("unexpected request %s %s", gotMethod, gotPath)
	}
	want := map[string]interface{}{
		"activation_id":           "example_activation_id",
		"operation":               "upsert",
		"source_attributes":       map[string]interface{}{"object": map[string]interface{}{"type": "dataset", "id": float64(123)}},
		"destination_attributes":  map[string]interface{}{"object": "Contact"},
		"mappings": []interface{}{
			map[string]interface{}{
				"from":                  map[string]interface{}{"type": "column", "data": "email"},
				"to":                    "Email",
				"is_primary_identifier": true,
			},
		},
	}
	if !reflect.DeepEqual(gotBody, want) {
		t.Fatalf("unexpected body (optional fields must be omitted):\n got: %#v\nwant: %#v", gotBody, want)
	}
}

func TestCreateSync_MissingIDIsError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"status": "created", "data": {"activation_id": "example_activation_id"}}`)
	})

	_, err := CreateSync(context.Background(), c, &CreateSyncRequest{})
	if err == nil || !strings.Contains(err.Error(), "may have been created anyway") {
		t.Fatalf("expected an ambiguous-create error, got %v", err)
	}
}

func TestGetSync_NotFound(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	_, err := GetSync(context.Background(), c, 4545)
	if !IsNotFoundError(err) {
		t.Fatalf("expected a not-found error, got %v", err)
	}
}

func TestGetSync_EmptyBodyIsError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	_, err := GetSync(context.Background(), c, 4545)
	if err == nil || !strings.Contains(err.Error(), "unexpected") {
		t.Fatalf("expected an empty-body error, got %v", err)
	}
}

func TestUpdateSync_SendsMethodAndPath(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]interface{}

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(raw, &gotBody); err != nil {
			t.Fatalf("request body is not JSON: %v", err)
		}
		io.WriteString(w, syncResponseJSON)
	})

	sync, err := UpdateSync(context.Background(), c, 4545, &UpdateSyncRequest{Paused: true})
	if err != nil {
		t.Fatalf("UpdateSync failed: %v", err)
	}
	if sync.ID != 4545 {
		t.Fatalf("unexpected sync ID %d", sync.ID)
	}

	if gotMethod != http.MethodPatch || gotPath != "/syncs/4545" {
		t.Fatalf("unexpected request %s %s", gotMethod, gotPath)
	}
	// Paused has no omitempty (see UpdateSyncRequest) — it must always be
	// sent explicitly, even when false, to actually unpause a sync.
	if got, ok := gotBody["paused"].(bool); !ok || !got {
		t.Fatalf("expected paused=true to be sent explicitly, got %#v", gotBody["paused"])
	}
}

func TestDeleteSync(t *testing.T) {
	var gotMethod, gotPath string
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		io.WriteString(w, `{"status": "deleted"}`)
	})

	if err := DeleteSync(context.Background(), c, 4545); err != nil {
		t.Fatalf("DeleteSync failed: %v", err)
	}
	if gotMethod != http.MethodDelete || gotPath != "/syncs/4545" {
		t.Fatalf("unexpected request %s %s", gotMethod, gotPath)
	}
}
