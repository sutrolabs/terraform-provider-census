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

// activationResponseJSON is a GET /activations/{activation_id} response.
const activationResponseJSON = `{
  "status": "success",
  "data": {
    "id": "glimmering_donkey",
    "name": "Google Sheets",
    "group_id": "abundant_giraffe",
    "source": {"id": 42, "name": "My Snowflake Warehouse", "type": "snowflake"},
    "destination": {
      "id": 123,
      "type": "salesforce",
      "created_at": "2023-07-22T23:42:47.239Z",
      "last_test_succeeded": true,
      "last_tested_at": "2023-07-22T23:45:12.123Z",
      "is_file_system_destination": false,
      "connection_details": {"account_id": 7515011393, "account_name": "Manager Account Test"},
      "last_test_results": {"steps": [
        {"step": "Validate Credentials", "status": "success", "error": null},
        {"step": "Check Permissions", "status": "error", "error": {"message": "Invalid API credentials"}}
      ]}
    }
  }
}`

func TestActivationUnmarshal_FullResponse(t *testing.T) {
	var env envelope[*Activation]
	if err := json.Unmarshal([]byte(activationResponseJSON), &env); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	succeeded := true
	createdAt := time.Date(2023, 7, 22, 23, 42, 47, 239000000, time.UTC)
	testedAt := time.Date(2023, 7, 22, 23, 45, 12, 123000000, time.UTC)
	message := "Invalid API credentials"
	want := &Activation{
		ID:      "glimmering_donkey",
		Name:    "Google Sheets",
		GroupID: "abundant_giraffe",
		Source:  &ActivationSource{ID: 42, Name: "My Snowflake Warehouse", Type: "snowflake"},
		Destination: &ActivationDestination{
			ID:                123,
			Type:              "salesforce",
			CreatedAt:         &createdAt,
			LastTestSucceeded: &succeeded,
			LastTestedAt:      &testedAt,
			ConnectionDetails: map[string]interface{}{"account_id": float64(7515011393), "account_name": "Manager Account Test"},
			LastTestResults: &ActivationTestResults{Steps: []ActivationTestStep{
				{Step: "Validate Credentials", Status: "success"},
				{Step: "Check Permissions", Status: "error", Error: &ActivationTestStepError{Message: &message}},
			}},
		},
	}
	if !reflect.DeepEqual(env.Data, want) {
		t.Fatalf("unexpected activation:\n got: %#v\nwant: %#v", env.Data, want)
	}
}

func TestActivationUnmarshal_NullTestFields(t *testing.T) {
	body := `{"id": "a", "destination": {"last_test_succeeded": null, "last_tested_at": null, "last_test_results": {"steps": null}}}`

	var activation Activation
	if err := json.Unmarshal([]byte(body), &activation); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	dest := activation.Destination
	if dest.LastTestSucceeded != nil || dest.LastTestedAt != nil || dest.LastTestResults == nil || dest.LastTestResults.Steps != nil {
		t.Fatalf("expected null test fields to stay nil, got %#v", dest)
	}
}

func TestCreateActivation_SendsWrappedBody(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]interface{}

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(raw, &gotBody); err != nil {
			t.Fatalf("request body is not JSON: %v", err)
		}
		w.WriteHeader(http.StatusCreated)
		io.WriteString(w, activationResponseJSON)
	})

	activation, err := CreateActivation(context.Background(), c, &ActivationRequest{Activation: ActivationAttributes{
		GroupID:     "abundant_giraffe",
		Type:        "salesforce",
		Credentials: map[string]interface{}{"api_key": "secret"},
	}})
	if err != nil {
		t.Fatalf("CreateActivation failed: %v", err)
	}
	if activation.ID != "glimmering_donkey" {
		t.Fatalf("unexpected ID %q", activation.ID)
	}

	if gotMethod != http.MethodPost || gotPath != "/activations" {
		t.Fatalf("unexpected request %s %s", gotMethod, gotPath)
	}
	want := map[string]interface{}{"activation": map[string]interface{}{
		"group_id":    "abundant_giraffe",
		"type":        "salesforce",
		"credentials": map[string]interface{}{"api_key": "secret"},
	}}
	if !reflect.DeepEqual(gotBody, want) {
		t.Fatalf("unexpected body (an unset name must be omitted):\n got: %#v\nwant: %#v", gotBody, want)
	}
}

func TestCreateActivation_MissingIDIsError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"status": "created", "data": {"name": "x"}}`)
	})

	_, err := CreateActivation(context.Background(), c, &ActivationRequest{})
	if err == nil || !strings.Contains(err.Error(), "may have been created anyway") {
		t.Fatalf("expected an ambiguous-create error, got %v", err)
	}
}

func TestGetActivation_EscapesID(t *testing.T) {
	var gotPath string
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		io.WriteString(w, activationResponseJSON)
	})

	if _, err := GetActivation(context.Background(), c, "a/b"); err != nil {
		t.Fatalf("GetActivation failed: %v", err)
	}
	if gotPath != "/activations/a%2Fb" {
		t.Fatalf("expected the ID to be path-escaped, got %q", gotPath)
	}
}

func TestGetActivation_NotFound(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	_, err := GetActivation(context.Background(), c, "missing")
	if !IsNotFoundError(err) {
		t.Fatalf("expected a not-found error, got %v", err)
	}
}

func TestDeleteActivation(t *testing.T) {
	var gotMethod, gotPath string
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		io.WriteString(w, `{"message": "Activation with ID glimmering_donkey successfully deleted"}`)
	})

	if err := DeleteActivation(context.Background(), c, "glimmering_donkey"); err != nil {
		t.Fatalf("DeleteActivation failed: %v", err)
	}
	if gotMethod != http.MethodDelete || gotPath != "/activations/glimmering_donkey" {
		t.Fatalf("unexpected request %s %s", gotMethod, gotPath)
	}
}

// activationListServer serves a paginated GET /activations for
// legacy_destination_id 123, two activations per page.
func activationListServer(t *testing.T, activations []map[string]interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/activations" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.String())
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if got := r.URL.Query().Get("legacy_destination_id"); got != "123" {
			t.Errorf("expected legacy_destination_id=123, got %q", got)
		}
		if got := r.URL.Query().Get("per_page"); got != "100" {
			t.Errorf("expected per_page=100, got %q", got)
		}
		page := 1
		if r.URL.Query().Get("page") == "2" {
			page = 2
		}
		start, end := (page-1)*2, page*2
		if end > len(activations) {
			end = len(activations)
		}
		if start > end {
			start = end
		}
		var next interface{}
		if end < len(activations) {
			next = page + 1
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":     "success",
			"data":       activations[start:end],
			"pagination": map[string]interface{}{"page": page, "next_page": next},
		})
	}
}

func groupActivation(id, groupID string) map[string]interface{} {
	return map[string]interface{}{"id": id, "group_id": groupID}
}

func TestFindActivationForGroupAndDestination_MatchesGroupAcrossPages(t *testing.T) {
	c := newTestClient(t, activationListServer(t, []map[string]interface{}{
		groupActivation("other_group_one", "other_group"),
		groupActivation("other_group_two", "other_group"),
		groupActivation("match", "abundant_giraffe"),
	}))

	activation, err := FindActivationForGroupAndDestination(context.Background(), c, "abundant_giraffe", 123)
	if err != nil {
		t.Fatalf("FindActivationForGroupAndDestination failed: %v", err)
	}
	if activation.ID != "match" {
		t.Fatalf("expected the activation on page 2 with the matching group, got %q", activation.ID)
	}
}

func TestFindActivationForGroupAndDestination_NoMatch(t *testing.T) {
	c := newTestClient(t, activationListServer(t, []map[string]interface{}{
		groupActivation("one", "other_group"),
	}))

	_, err := FindActivationForGroupAndDestination(context.Background(), c, "abundant_giraffe", 123)
	if err == nil || !strings.Contains(err.Error(), "no activation found for group abundant_giraffe and destination 123") {
		t.Fatalf("expected a no-match error, got %v", err)
	}
}

func TestFindActivationForGroupAndDestination_MoreThanOneMatch(t *testing.T) {
	c := newTestClient(t, activationListServer(t, []map[string]interface{}{
		groupActivation("one", "abundant_giraffe"),
		groupActivation("two", "abundant_giraffe"),
	}))

	_, err := FindActivationForGroupAndDestination(context.Background(), c, "abundant_giraffe", 123)
	if err == nil || !strings.Contains(err.Error(), "more than one activation") {
		t.Fatalf("expected a more-than-one error, got %v", err)
	}
}
