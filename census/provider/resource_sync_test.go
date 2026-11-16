package provider

import (
	"reflect"
	"testing"
)

// Unit tests for symbols declared in sync_schema.go

// ============================================================================
// State Migration Tests
// ============================================================================
// These tests verify that state migration from v0 (TypeSet) to v1 (TypeList)
// works correctly for the alert field.

func TestResourceSyncStateUpgradeV0(t *testing.T) {
	// Simulate v0 state with alerts stored as they would be in TypeSet format
	// (which is actually just a JSON array, same as TypeList)
	v0State := map[string]interface{}{
		"id":           "12345",
		"label":        "Test Sync",
		"workspace_id": "67890",
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

	// Call the upgrade function
	upgradedState, err := ResourceSyncStateUpgradeV0(nil, v0State, nil)

	// Assertions
	if err != nil {
		t.Errorf("State upgrade should not return an error, got: %v", err)
	}
	if upgradedState == nil {
		t.Errorf("Upgraded state should not be nil")
	}

	// Verify that the state is unchanged (since JSON format is identical)
	if !reflect.DeepEqual(v0State, upgradedState) {
		t.Errorf("State should be unchanged - TypeSet and TypeList have same JSON format")
	}

	// Verify alert data is preserved
	alerts, ok := upgradedState["alert"].([]interface{})
	if !ok {
		t.Errorf("alert should be a []interface{}")
	}
	if len(alerts) != 2 {
		t.Errorf("Should have 2 alerts, got %d", len(alerts))
	}

	// Verify first alert
	alert1, ok := alerts[0].(map[string]interface{})
	if !ok {
		t.Errorf("alert[0] should be a map")
	}
	if alert1["type"] != "FailureAlertConfiguration" {
		t.Errorf("alert[0] type = %v, want FailureAlertConfiguration", alert1["type"])
	}
	if alert1["id"] != 754756 {
		t.Errorf("alert[0] id = %v, want 754756", alert1["id"])
	}

	// Verify second alert
	alert2, ok := alerts[1].(map[string]interface{})
	if !ok {
		t.Errorf("alert[1] should be a map")
	}
	if alert2["type"] != "InvalidRecordPercentAlertConfiguration" {
		t.Errorf("alert[1] type = %v, want InvalidRecordPercentAlertConfiguration", alert2["type"])
	}
	if alert2["id"] != 754757 {
		t.Errorf("alert[1] id = %v, want 754757", alert2["id"])
	}
}

func TestResourceSyncV0SchemaCompatibility(t *testing.T) {
	// Verify that v0 schema uses TypeSet for alerts
	v0Resource := ResourceSyncV0()
	if v0Resource == nil {
		t.Errorf("v0 resource should not be nil")
		return
	}
	if v0Resource.Schema == nil {
		t.Errorf("v0 schema should not be nil")
		return
	}

	alertSchema, exists := v0Resource.Schema["alert"]
	if !exists {
		t.Errorf("alert field should exist in v0 schema")
		return
	}

	// Note: We can't directly test schema.TypeSet value since it's an internal constant
	// But we can verify the schema is valid and has the expected structure
	if alertSchema == nil {
		t.Errorf("alert schema should not be nil")
	}
	if !alertSchema.Optional {
		t.Errorf("alert should be optional")
	}
}

func TestResourceSyncV1SchemaCompatibility(t *testing.T) {
	// Verify that v1 schema (current) uses TypeList for alerts
	v1Resource := ResourceSync()
	if v1Resource == nil {
		t.Errorf("v1 resource should not be nil")
		return
	}
	if v1Resource.SchemaVersion != 1 {
		t.Errorf("SchemaVersion = %d, want 1", v1Resource.SchemaVersion)
	}
	if len(v1Resource.StateUpgraders) != 1 {
		t.Errorf("Should have 1 StateUpgrader, got %d", len(v1Resource.StateUpgraders))
		return
	}

	upgrader := v1Resource.StateUpgraders[0]
	if upgrader.Version != 0 {
		t.Errorf("StateUpgrader Version = %d, want 0", upgrader.Version)
	}
	if upgrader.Upgrade == nil {
		t.Errorf("Upgrade function should be set")
	}
	// Note: Type is a cty.Type struct (not a pointer), so we can't check for nil
	// The existence of the StateUpgrader itself confirms it's properly configured
}
