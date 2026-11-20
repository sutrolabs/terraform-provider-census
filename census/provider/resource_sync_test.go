package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"testing"

	"github.com/sutrolabs/terraform-provider-census/census/client/v2"
)

// Unit tests for sync resource helper functions
// These tests do NOT require API credentials or external dependencies

// ============================================================================
// Field Mapping Tests (Expand: Terraform config -> API MappingAttributes)
// ============================================================================

func TestExpandFieldMappings_Direct(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected []v2.MappingAttributes
	}{
		{
			name: "direct mapping with primary identifier",
			input: []interface{}{
				map[string]interface{}{
					"from":                  "email",
					"to":                    "Email",
					"type":                  "direct",
					"is_primary_identifier": true,
				},
			},
			expected: []v2.MappingAttributes{
				{
					From:                v2.MappingFrom{Type: "column", Data: "email"},
					To:                  "Email",
					IsPrimaryIdentifier: true,
				},
			},
		},
		{
			name: "direct mapping without primary identifier (type defaults to direct)",
			input: []interface{}{
				map[string]interface{}{
					"from": "first_name",
					"to":   "FirstName",
				},
			},
			expected: []v2.MappingAttributes{
				{
					From: v2.MappingFrom{Type: "column", Data: "first_name"},
					To:   "FirstName",
				},
			},
		},
		{
			name: "mapping with array operation",
			input: []interface{}{
				map[string]interface{}{
					"from":        "tags",
					"to":          "Tags",
					"array_field": true,
					"operation":   "append",
				},
			},
			expected: []v2.MappingAttributes{
				{
					From:       v2.MappingFrom{Type: "column", Data: "tags"},
					To:         "Tags",
					ArrayField: true,
					Operation:  "append",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExpandFieldMappings(context.Background(), tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ExpandFieldMappings() got = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestExpandFieldMappings_Constant(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected []v2.MappingAttributes
	}{
		{
			name: "constant string value",
			input: []interface{}{
				map[string]interface{}{
					"to":       "Source",
					"type":     "constant",
					"constant": "Website",
				},
			},
			expected: []v2.MappingAttributes{
				{
					From: v2.MappingFrom{
						Type: "constant_value",
						Data: map[string]interface{}{"basic_type": "text", "value": "Website"},
					},
					To: "Source",
				},
			},
		},
		{
			name: "constant numeric value is stringified",
			input: []interface{}{
				map[string]interface{}{
					"to":       "Priority",
					"type":     "constant",
					"constant": 1,
				},
			},
			expected: []v2.MappingAttributes{
				{
					From: v2.MappingFrom{
						Type: "constant_value",
						Data: map[string]interface{}{"basic_type": "text", "value": "1"},
					},
					To: "Priority",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExpandFieldMappings(context.Background(), tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ExpandFieldMappings() got = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestExpandFieldMappings_LiquidTemplate(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{
			"to":              "FullName",
			"type":            "liquid_template",
			"liquid_template": "{{ first_name }} {{ last_name }}",
		},
	}
	expected := []v2.MappingAttributes{
		{
			From: v2.MappingFrom{Type: "liquid_template", Data: "{{ first_name }} {{ last_name }}"},
			To:   "FullName",
		},
	}

	result := ExpandFieldMappings(context.Background(), input)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("ExpandFieldMappings() got = %+v, want %+v", result, expected)
	}
}

func TestExpandFieldMappings_SyncMetadata(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{
			"to":                "Sync_Run_ID",
			"type":              "sync_metadata",
			"sync_metadata_key": "sync_run_id",
		},
	}
	expected := []v2.MappingAttributes{
		{
			From: v2.MappingFrom{Type: "sync_metadata", Data: "sync_run_id"},
			To:   "Sync_Run_ID",
		},
	}

	result := ExpandFieldMappings(context.Background(), input)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("ExpandFieldMappings() got = %+v, want %+v", result, expected)
	}
}

func TestExpandFieldMappings_SegmentMembership(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{
			"to":                  "SegmentField",
			"type":                "segment_membership",
			"segment_identify_by": "name",
		},
	}
	expected := []v2.MappingAttributes{
		{
			From: v2.MappingFrom{Type: "segment_membership", Data: map[string]interface{}{"identify_by": "name"}},
			To:   "SegmentField",
		},
	}

	result := ExpandFieldMappings(context.Background(), input)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("ExpandFieldMappings() got = %+v, want %+v", result, expected)
	}
}

func TestExpandFieldMappings_Empty(t *testing.T) {
	result := ExpandFieldMappings(context.Background(), []interface{}{})
	if len(result) != 0 {
		t.Errorf("ExpandFieldMappings(empty) should return empty slice, got %d items", len(result))
	}
}

// ============================================================================
// Field Mapping Tests (Flatten: API MappingAttributes -> Terraform config)
// ============================================================================

func TestFlattenFieldMappings_Direct(t *testing.T) {
	input := []v2.MappingAttributes{
		{
			From:                v2.MappingFrom{Type: "column", Data: "email"},
			To:                  "Email",
			IsPrimaryIdentifier: true,
		},
		{
			From: v2.MappingFrom{Type: "column", Data: "first_name"},
			To:   "FirstName",
		},
	}

	result := FlattenFieldMappings(input)
	if len(result) != 2 {
		t.Fatalf("FlattenFieldMappings() returned %d items, want 2", len(result))
	}

	first := result[0].(map[string]interface{})
	if first["from"] != "email" || first["to"] != "Email" || first["type"] != "direct" || first["is_primary_identifier"] != true {
		t.Errorf("unexpected first mapping: %+v", first)
	}
}

func TestFlattenFieldMappings_Constant(t *testing.T) {
	input := []v2.MappingAttributes{
		{
			From: v2.MappingFrom{
				Type: "constant_value",
				Data: map[string]interface{}{"value": "Website", "basic_type": "text"},
			},
			To: "Source",
		},
	}

	result := FlattenFieldMappings(input)
	m := result[0].(map[string]interface{})
	if m["type"] != "constant" || m["constant"] != "Website" || m["from"] != "" {
		t.Errorf("unexpected mapping: %+v", m)
	}
}

func TestFlattenFieldMappings_SyncMetadata(t *testing.T) {
	input := []v2.MappingAttributes{
		{From: v2.MappingFrom{Type: "sync_metadata", Data: "sync_run_id"}, To: "SyncRunID"},
	}

	result := FlattenFieldMappings(input)
	m := result[0].(map[string]interface{})
	if m["type"] != "sync_metadata" || m["sync_metadata_key"] != "sync_run_id" {
		t.Errorf("unexpected mapping: %+v", m)
	}
}

func TestFlattenFieldMappings_SegmentMembership(t *testing.T) {
	input := []v2.MappingAttributes{
		{
			From: v2.MappingFrom{Type: "segment_membership", Data: map[string]interface{}{"identify_by": "name"}},
			To:   "SegmentField",
		},
	}

	result := FlattenFieldMappings(input)
	m := result[0].(map[string]interface{})
	if m["type"] != "segment_membership" || m["segment_identify_by"] != "name" {
		t.Errorf("unexpected mapping: %+v", m)
	}
}

func TestFlattenFieldMappings_LiquidTemplate(t *testing.T) {
	input := []v2.MappingAttributes{
		{From: v2.MappingFrom{Type: "liquid_template", Data: "{{ record['date'] | upcase }}"}, To: "FormattedDate"},
	}

	result := FlattenFieldMappings(input)
	m := result[0].(map[string]interface{})
	if m["type"] != "liquid_template" || m["liquid_template"] != "{{ record['date'] | upcase }}" {
		t.Errorf("unexpected mapping: %+v", m)
	}
}

func TestFlattenFieldMappings_Operation(t *testing.T) {
	input := []v2.MappingAttributes{
		{From: v2.MappingFrom{Type: "column", Data: "tags"}, To: "Tags", ArrayField: true, Operation: "append"},
	}

	result := FlattenFieldMappings(input)
	m := result[0].(map[string]interface{})
	if m["operation"] != "append" || m["array_field"] != true {
		t.Errorf("unexpected mapping: %+v", m)
	}
}

func TestFlattenFieldMappings_SyncNullValuesDefault(t *testing.T) {
	// If the API doesn't return sync_null_values, the schema default (true) is used.
	input := []v2.MappingAttributes{
		{From: v2.MappingFrom{Type: "column", Data: "email"}, To: "Email"},
	}

	result := FlattenFieldMappings(input)
	m := result[0].(map[string]interface{})
	if m["sync_null_values"] != true {
		t.Errorf("expected sync_null_values to default to true, got %v", m["sync_null_values"])
	}
}

func TestFlattenFieldMappings_Empty(t *testing.T) {
	result := FlattenFieldMappings([]v2.MappingAttributes{})
	if len(result) != 0 {
		t.Errorf("FlattenFieldMappings(empty) should return empty slice, got %d items", len(result))
	}
}

func TestFieldMappings_RoundTrip(t *testing.T) {
	// Terraform config -> API request -> Terraform state should round-trip
	// every mapping type without loss (aside from "from" being empty for
	// non-column mapping types, same as before this stage's rewrite).
	original := []interface{}{
		map[string]interface{}{
			"from":                  "email",
			"to":                    "Email",
			"type":                  "direct",
			"is_primary_identifier": true,
		},
		map[string]interface{}{
			"to":              "FormattedDate",
			"type":            "liquid_template",
			"liquid_template": "{{ record['date'] | upcase }}",
		},
		map[string]interface{}{
			"to":                "SyncTime",
			"type":              "sync_metadata",
			"sync_metadata_key": "sync_run_id",
		},
		map[string]interface{}{
			"to":       "Source",
			"type":     "constant",
			"constant": "Website",
		},
	}

	expanded := ExpandFieldMappings(context.Background(), original)
	flattened := FlattenFieldMappings(expanded)

	for i, m := range flattened {
		got := m.(map[string]interface{})
		want := original[i].(map[string]interface{})
		if got["to"] != want["to"] || got["type"] != want["type"] {
			t.Errorf("mapping %d: got = %+v, want (at least) %+v", i, got, want)
		}
	}
}

// ============================================================================
// Field Mapping Ordering and Compound Key Filtering Tests
// ============================================================================
// These verify the logic in resourceSyncRead that sorts mappings by
// Census-assigned position and filters out Census-managed compound_key
// mappings, preventing spurious diffs and untrackable state respectively.

func TestSortFieldMappingsByPosition_Basic(t *testing.T) {
	mappings := []v2.MappingAttributes{
		{To: "LastName", Position: 2},
		{To: "Email", Position: 0},
		{To: "FirstName", Position: 1},
	}

	sort.Slice(mappings, func(i, j int) bool {
		return mappings[i].Position < mappings[j].Position
	})

	expectedOrder := []string{"Email", "FirstName", "LastName"}
	for i, mapping := range mappings {
		if mapping.To != expectedOrder[i] {
			t.Errorf("position %d: got %s, want %s", i, mapping.To, expectedOrder[i])
		}
	}
}

func TestSortFieldMappingsByPosition_AlreadyOrdered(t *testing.T) {
	mappings := []v2.MappingAttributes{
		{To: "Email", Position: 0},
		{To: "FirstName", Position: 1},
		{To: "LastName", Position: 2},
	}

	sort.Slice(mappings, func(i, j int) bool {
		return mappings[i].Position < mappings[j].Position
	})

	expectedOrder := []string{"Email", "FirstName", "LastName"}
	for i, mapping := range mappings {
		if mapping.To != expectedOrder[i] {
			t.Errorf("position %d: got %s, want %s", i, mapping.To, expectedOrder[i])
		}
	}
}

func TestSortFieldMappingsByPosition_EmptySlice(t *testing.T) {
	mappings := []v2.MappingAttributes{}
	sort.Slice(mappings, func(i, j int) bool {
		return mappings[i].Position < mappings[j].Position
	})
	if len(mappings) != 0 {
		t.Errorf("expected 0 mappings, got %d", len(mappings))
	}
}

func TestFilterCompoundKeyMappings(t *testing.T) {
	tests := []struct {
		name     string
		input    []v2.MappingAttributes
		expected int
	}{
		{
			name: "filter out a single compound_key mapping",
			input: []v2.MappingAttributes{
				{From: v2.MappingFrom{Type: "column", Data: "ID"}, To: "a", Position: 0},
				{From: v2.MappingFrom{Type: "column", Data: "ID"}, To: "b", Position: 1},
				{From: v2.MappingFrom{Type: "compound_key"}, To: "_census_tracking_id", Position: 2},
			},
			expected: 2,
		},
		{
			name: "no compound_key mappings to filter",
			input: []v2.MappingAttributes{
				{From: v2.MappingFrom{Type: "column", Data: "email"}, To: "Email", IsPrimaryIdentifier: true},
				{From: v2.MappingFrom{Type: "column", Data: "first_name"}, To: "FirstName"},
			},
			expected: 2,
		},
		{
			name:     "empty input",
			input:    []v2.MappingAttributes{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userManagedMappings := make([]v2.MappingAttributes, 0, len(tt.input))
			for _, mapping := range tt.input {
				if mapping.From.Type != "compound_key" {
					userManagedMappings = append(userManagedMappings, mapping)
				}
			}

			if len(userManagedMappings) != tt.expected {
				t.Errorf("expected %d mappings after filtering, got %d", tt.expected, len(userManagedMappings))
			}
			for _, mapping := range userManagedMappings {
				if mapping.From.Type == "compound_key" {
					t.Errorf("compound_key mapping should have been filtered: %+v", mapping)
				}
			}
		})
	}
}

// ============================================================================
// Alert Tests
// ============================================================================

func TestExpandAlerts_Basic(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected int
	}{
		{
			name: "basic alert",
			input: []interface{}{
				map[string]interface{}{
					"type":                 "FailureAlertConfiguration",
					"send_for":             "first_time",
					"should_send_recovery": true,
				},
			},
			expected: 1,
		},
		{
			name: "multiple alerts",
			input: []interface{}{
				map[string]interface{}{"type": "FailureAlertConfiguration", "send_for": "first_time"},
				map[string]interface{}{"type": "RuntimeAlertConfiguration", "send_for": "every_time"},
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExpandAlerts(context.Background(), tt.input)
			if len(result) != tt.expected {
				t.Errorf("ExpandAlerts() returned %d items, want %d", len(result), tt.expected)
			}
		})
	}
}

// TestExpandAlerts_OptionsNeverNil verifies that v2's requirement that every
// alert send options (even {}) is satisfied even when the alert type has no
// options of its own — a nil map would marshal to JSON null, which the v2
// API rejects.
func TestExpandAlerts_OptionsNeverNil(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{"type": "FailureAlertConfiguration", "send_for": "first_time"},
	}

	result := ExpandAlerts(context.Background(), input)
	if result[0].Options == nil {
		t.Error("expected Options to be a non-nil (possibly empty) map, got nil")
	}
}

func TestExpandAlerts_Empty(t *testing.T) {
	result := ExpandAlerts(context.Background(), []interface{}{})
	if result == nil {
		t.Error("ExpandAlerts(empty) should return empty slice, got nil")
	}
	if len(result) != 0 {
		t.Errorf("ExpandAlerts(empty) should return empty slice with 0 items, got %d items", len(result))
	}
}

// TestExpandAlerts_EmptyStrings tests that alerts with empty string values
// are properly skipped, preventing invalid API payloads (regression test)
func TestExpandAlerts_EmptyStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected int
		checkFn  func(*testing.T, []v2.AlertAttribute)
	}{
		{
			name: "alert with empty type is skipped",
			input: []interface{}{
				map[string]interface{}{"type": "", "send_for": "first_time", "should_send_recovery": false},
			},
			expected: 0,
		},
		{
			name: "alert with empty send_for uses default",
			input: []interface{}{
				map[string]interface{}{"type": "FailureAlertConfiguration", "send_for": "", "should_send_recovery": true},
			},
			expected: 1,
			checkFn: func(t *testing.T, alerts []v2.AlertAttribute) {
				if alerts[0].SendFor != "first_time" {
					t.Errorf("expected SendFor to default to 'first_time', got '%s'", alerts[0].SendFor)
				}
			},
		},
		{
			name: "mixed valid and invalid alerts",
			input: []interface{}{
				map[string]interface{}{"type": "", "send_for": "first_time"},
				map[string]interface{}{"type": "FailureAlertConfiguration", "send_for": "every_time"},
				map[string]interface{}{"type": "", "send_for": ""},
			},
			expected: 1,
			checkFn: func(t *testing.T, alerts []v2.AlertAttribute) {
				if alerts[0].Type != "FailureAlertConfiguration" || alerts[0].SendFor != "every_time" {
					t.Errorf("unexpected surviving alert: %+v", alerts[0])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExpandAlerts(context.Background(), tt.input)
			if len(result) != tt.expected {
				t.Errorf("ExpandAlerts() returned %d items, want %d", len(result), tt.expected)
			}
			if tt.checkFn != nil {
				tt.checkFn(t, result)
			}
		})
	}
}

func TestFlattenAlerts_Empty(t *testing.T) {
	result := FlattenAlerts(nil)
	if len(result) != 0 {
		t.Errorf("FlattenAlerts(nil) should return empty slice, got %d items", len(result))
	}
}

func TestFlattenAlerts_Basic(t *testing.T) {
	input := []v2.AlertAttribute{
		{
			ID:                 123,
			Type:               "FailureAlertConfiguration",
			SendFor:            "first_time",
			ShouldSendRecovery: true,
			Options:            map[string]interface{}{},
		},
	}

	result := FlattenAlerts(input)
	m := result[0].(map[string]interface{})
	if m["id"] != 123 || m["type"] != "FailureAlertConfiguration" || m["send_for"] != "first_time" {
		t.Errorf("unexpected alert: %+v", m)
	}
}

// ============================================================================
// Utility Helper Tests
// ============================================================================

func TestCleanEmptyStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name:     "empty map",
			input:    map[string]interface{}{},
			expected: map[string]interface{}{},
		},
		{
			name: "map with empty strings - removed",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": "",
				"key3": "value3",
			},
			expected: map[string]interface{}{
				"key1": "value1",
				"key3": "value3",
			},
		},
		{
			name: "map with zero cohort_id - removed",
			input: map[string]interface{}{
				"key1":      "value1",
				"cohort_id": 0,
				"key3":      "value3",
			},
			expected: map[string]interface{}{
				"key1": "value1",
				"key3": "value3",
			},
		},
		{
			name: "map with non-zero integers - kept",
			input: map[string]interface{}{
				"key1":  "value1",
				"count": 0,
				"key3":  "value3",
			},
			expected: map[string]interface{}{
				"key1":  "value1",
				"count": 0,
				"key3":  "value3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanEmptyStrings(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("CleanEmptyStrings() got = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestExpandDestinationAttributes_Basic(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{
			"object":               "Contact",
			"lead_union_insert_to": "Lead",
		},
	}

	result := ExpandDestinationAttributes(context.Background(), input)
	if result.Object != "Contact" || result.LeadUnionInsertTo != "Lead" {
		t.Errorf("unexpected result: %+v", result)
	}
}

// ============================================================================
// Primary Identifier Validation Tests
// ============================================================================

func TestExpandFieldMappings_NoPrimaryIdentifier(t *testing.T) {
	// The Census API handles primary identifier validation at the API level
	// (destinations like Google Sheets don't require one); expand doesn't
	// enforce it client-side.
	input := []interface{}{
		map[string]interface{}{"from": "city", "to": "City"},
		map[string]interface{}{"from": "state", "to": "State"},
	}

	result := ExpandFieldMappings(context.Background(), input)
	for i, mapping := range result {
		if mapping.IsPrimaryIdentifier {
			t.Errorf("mapping %d unexpectedly has IsPrimaryIdentifier=true", i)
		}
	}
}

func TestValidateFieldMappingsHavePrimaryIdentifier(t *testing.T) {
	tests := []struct {
		name      string
		mappings  []interface{}
		expectErr bool
	}{
		{
			name: "one mapping with primary identifier",
			mappings: []interface{}{
				map[string]interface{}{"to": "Email", "is_primary_identifier": true},
				map[string]interface{}{"to": "Name"},
			},
			expectErr: false,
		},
		{
			name: "no mapping with primary identifier",
			mappings: []interface{}{
				map[string]interface{}{"to": "Email"},
				map[string]interface{}{"to": "Name"},
			},
			expectErr: true,
		},
		{
			name:      "no mappings at all is allowed (e.g. sync_all_properties)",
			mappings:  []interface{}{},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFieldMappingsHavePrimaryIdentifier(tt.mappings)
			if tt.expectErr && err == nil {
				t.Error("expected an error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

// ============================================================================
// State Migration Tests
// ============================================================================
// These tests verify that state migration from v0 (TypeSet) to v1 (TypeList)
// works correctly for the alert field.

func TestResourceSyncStateUpgradeV0(t *testing.T) {
	v0State := map[string]interface{}{
		"id":            "12345",
		"label":         "Test Sync",
		"activation_id": "act_67890",
		"alert": []interface{}{
			map[string]interface{}{
				"id":                   754756,
				"type":                 "FailureAlertConfiguration",
				"send_for":             "first_time",
				"should_send_recovery": true,
				"options":              map[string]interface{}{},
			},
			map[string]interface{}{
				"id":                   754757,
				"type":                 "InvalidRecordPercentAlertConfiguration",
				"send_for":             "first_time",
				"should_send_recovery": true,
				"options": map[string]interface{}{
					"threshold": "75",
				},
			},
		},
	}

	upgradedState, err := ResourceSyncStateUpgradeV0(nil, v0State, nil)
	if err != nil {
		t.Errorf("state upgrade should not return an error, got: %v", err)
	}
	if !reflect.DeepEqual(v0State, upgradedState) {
		t.Errorf("state should be unchanged - TypeSet and TypeList have same JSON format")
	}

	alerts, ok := upgradedState["alert"].([]interface{})
	if !ok || len(alerts) != 2 {
		t.Fatalf("expected 2 alerts, got %#v", upgradedState["alert"])
	}
}

// TestResourceSyncStateUpgradeV1_BackfillsActivationID verifies that state
// left over from before activation_id existed gets it filled in from a
// GetSync lookup, so the Required+ForceNew field doesn't force a spurious
// replace on first apply after upgrading.
func TestResourceSyncStateUpgradeV1_BackfillsActivationID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/syncs/12345" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"success","data":{"id":12345,"activation_id":"act_real_123","created_at":"2023-01-01T00:00:00Z","updated_at":"2023-01-01T00:00:00Z"}}`))
	}))
	defer server.Close()

	rawState := map[string]interface{}{
		"id":            "12345",
		"activation_id": "",
		"label":         "Test Sync",
	}

	upgradedState, err := ResourceSyncStateUpgradeV1(context.Background(), rawState, newTestClients(t, server.URL))
	if err != nil {
		t.Fatalf("state upgrade should not return an error, got: %v", err)
	}
	if upgradedState["activation_id"] != "act_real_123" {
		t.Errorf("expected activation_id to be backfilled to %q, got %#v", "act_real_123", upgradedState["activation_id"])
	}
}

// TestResourceSyncStateUpgradeV1_NoopWhenAlreadyPresent verifies that state
// already carrying an activation_id (any sync created/read after this
// migration landed) skips the lookup entirely.
func TestResourceSyncStateUpgradeV1_NoopWhenAlreadyPresent(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		http.NotFound(w, r)
	}))
	defer server.Close()

	rawState := map[string]interface{}{
		"id":            "12345",
		"activation_id": "act_already_set",
	}

	upgradedState, err := ResourceSyncStateUpgradeV1(context.Background(), rawState, newTestClients(t, server.URL))
	if err != nil {
		t.Fatalf("state upgrade should not return an error, got: %v", err)
	}
	if called {
		t.Error("expected GetSync not to be called when activation_id is already present")
	}
	if upgradedState["activation_id"] != "act_already_set" {
		t.Errorf("activation_id should be unchanged, got %#v", upgradedState["activation_id"])
	}
}

// TestResourceSyncStateUpgradeV1_NotFoundTolerated verifies that a sync
// deleted out from under Terraform doesn't hard-fail the upgrade — the
// normal read path clears it from state on the next refresh.
func TestResourceSyncStateUpgradeV1_NotFoundTolerated(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	}))
	defer server.Close()

	rawState := map[string]interface{}{
		"id":            "12345",
		"activation_id": "",
	}

	upgradedState, err := ResourceSyncStateUpgradeV1(context.Background(), rawState, newTestClients(t, server.URL))
	if err != nil {
		t.Errorf("a 404 should not fail the state upgrade, got: %v", err)
	}
	if upgradedState["activation_id"] != "" {
		t.Errorf("activation_id should stay empty when the sync is gone, got %#v", upgradedState["activation_id"])
	}
}

// TestResourceSyncStateUpgradeV1_PropagatesOtherErrors verifies that
// non-404 failures (auth, network, 5xx) fail the upgrade loudly instead of
// silently leaving activation_id empty, which would just defer the forced
// replace to an unpredictable later point.
func TestResourceSyncStateUpgradeV1_PropagatesOtherErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"internal error"}`))
	}))
	defer server.Close()

	rawState := map[string]interface{}{
		"id":            "12345",
		"activation_id": "",
	}

	if _, err := ResourceSyncStateUpgradeV1(context.Background(), rawState, newTestClients(t, server.URL)); err == nil {
		t.Error("expected a non-404 API error to fail the state upgrade")
	}
}

// TestResourceSyncStateUpgradeV1_MissingID verifies the upgrader doesn't
// panic or call the API when id is absent or unparseable.
func TestResourceSyncStateUpgradeV1_MissingID(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		http.NotFound(w, r)
	}))
	defer server.Close()

	rawState := map[string]interface{}{
		"activation_id": "",
	}

	upgradedState, err := ResourceSyncStateUpgradeV1(context.Background(), rawState, newTestClients(t, server.URL))
	if err != nil {
		t.Fatalf("state upgrade should not return an error, got: %v", err)
	}
	if called {
		t.Error("expected GetSync not to be called when id is missing")
	}
	if upgradedState["activation_id"] != "" {
		t.Errorf("activation_id should be unchanged, got %#v", upgradedState["activation_id"])
	}
}

func TestResourceSyncV0SchemaCompatibility(t *testing.T) {
	v0Resource := ResourceSyncV0()
	alertSchema, exists := v0Resource.Schema["alert"]
	if !exists {
		t.Fatal("alert field should exist in v0 schema")
	}
	if !alertSchema.Optional {
		t.Error("alert should be optional")
	}
}

func TestResourceSyncV1SchemaCompatibility(t *testing.T) {
	v1Resource := ResourceSync()
	if v1Resource.SchemaVersion != 2 {
		t.Errorf("SchemaVersion = %d, want 2", v1Resource.SchemaVersion)
	}
	if len(v1Resource.StateUpgraders) != 2 {
		t.Fatalf("should have 2 StateUpgraders, got %d", len(v1Resource.StateUpgraders))
	}

	upgraderV0 := v1Resource.StateUpgraders[0]
	if upgraderV0.Version != 0 {
		t.Errorf("StateUpgraders[0] Version = %d, want 0", upgraderV0.Version)
	}
	if upgraderV0.Upgrade == nil {
		t.Error("StateUpgraders[0] Upgrade function should be set")
	}

	upgraderV1 := v1Resource.StateUpgraders[1]
	if upgraderV1.Version != 1 {
		t.Errorf("StateUpgraders[1] Version = %d, want 1", upgraderV1.Version)
	}
	if upgraderV1.Upgrade == nil {
		t.Error("StateUpgraders[1] Upgrade function should be set")
	}

	if _, ok := v1Resource.Schema["activation_id"]; !ok {
		t.Error("expected activation_id in schema")
	}
	if _, ok := v1Resource.Schema["workspace_id"]; ok {
		t.Error("workspace_id should no longer be in the v2 schema")
	}
}
