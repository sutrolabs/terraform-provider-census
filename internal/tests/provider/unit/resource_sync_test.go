package unit_test

import (
	"reflect"
	"testing"

	"github.com/sutrolabs/terraform-provider-census/internal/client"
	"github.com/sutrolabs/terraform-provider-census/internal/provider"
)

// Unit tests for sync resource helper functions
// These tests do NOT require API credentials or external dependencies

// ============================================================================
// Field Mapping Tests
// ============================================================================

func TestExpandFieldMappings_Direct(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected []client.FieldMapping
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
			expected: []client.FieldMapping{
				{
					From:                "email",
					To:                  "Email",
					Type:                "direct",
					IsPrimaryIdentifier: true,
				},
			},
		},
		{
			name: "direct mapping without primary identifier",
			input: []interface{}{
				map[string]interface{}{
					"from": "first_name",
					"to":   "FirstName",
				},
			},
			expected: []client.FieldMapping{
				{
					From: "first_name",
					To:   "FirstName",
					Type: "direct", // Default
				},
			},
		},
		{
			name: "multiple direct mappings",
			input: []interface{}{
				map[string]interface{}{
					"from":                  "email",
					"to":                    "Email",
					"is_primary_identifier": true,
				},
				map[string]interface{}{
					"from": "first_name",
					"to":   "FirstName",
				},
				map[string]interface{}{
					"from": "last_name",
					"to":   "LastName",
				},
			},
			expected: []client.FieldMapping{
				{
					From:                "email",
					To:                  "Email",
					Type:                "direct",
					IsPrimaryIdentifier: true,
				},
				{
					From: "first_name",
					To:   "FirstName",
					Type: "direct",
				},
				{
					From: "last_name",
					To:   "LastName",
					Type: "direct",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.ExpandFieldMappings(tt.input)
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
		expected []client.FieldMapping
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
			expected: []client.FieldMapping{
				{
					To:       "Source",
					Type:     "constant",
					Constant: "Website",
				},
			},
		},
		{
			name: "constant numeric value",
			input: []interface{}{
				map[string]interface{}{
					"to":       "Priority",
					"type":     "constant",
					"constant": 1,
				},
			},
			expected: []client.FieldMapping{
				{
					To:       "Priority",
					Type:     "constant",
					Constant: 1,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.ExpandFieldMappings(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ExpandFieldMappings() got = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestExpandFieldMappings_LiquidTemplate(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected []client.FieldMapping
	}{
		{
			name: "liquid template mapping",
			input: []interface{}{
				map[string]interface{}{
					"to":              "FullName",
					"type":            "liquid_template",
					"liquid_template": "{{ first_name }} {{ last_name }}",
				},
			},
			expected: []client.FieldMapping{
				{
					To:             "FullName",
					Type:           "liquid_template",
					LiquidTemplate: "{{ first_name }} {{ last_name }}",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.ExpandFieldMappings(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ExpandFieldMappings() got = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestExpandFieldMappings_SyncMetadata(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected []client.FieldMapping
	}{
		{
			name: "sync metadata mapping",
			input: []interface{}{
				map[string]interface{}{
					"to":                "Sync_Run_ID",
					"type":              "sync_metadata",
					"sync_metadata_key": "sync_run_id",
				},
			},
			expected: []client.FieldMapping{
				{
					To:              "Sync_Run_ID",
					Type:            "sync_metadata",
					SyncMetadataKey: "sync_run_id",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.ExpandFieldMappings(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ExpandFieldMappings() got = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestExpandFieldMappings_Empty(t *testing.T) {
	result := provider.ExpandFieldMappings([]interface{}{})
	if len(result) != 0 {
		t.Errorf("ExpandFieldMappings(empty) should return empty slice, got %d items", len(result))
	}
}

func TestFlattenFieldMappings_Direct(t *testing.T) {
	tests := []struct {
		name     string
		input    []client.FieldMapping
		expected int // We'll check length and specific fields
	}{
		{
			name: "direct mapping with primary identifier",
			input: []client.FieldMapping{
				{
					From:                "email",
					To:                  "Email",
					Type:                "direct",
					IsPrimaryIdentifier: true,
				},
			},
			expected: 1,
		},
		{
			name: "multiple direct mappings",
			input: []client.FieldMapping{
				{
					From:                "email",
					To:                  "Email",
					Type:                "direct",
					IsPrimaryIdentifier: true,
				},
				{
					From: "first_name",
					To:   "FirstName",
					Type: "direct",
				},
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.FlattenFieldMappings(tt.input)
			if len(result) != tt.expected {
				t.Errorf("FlattenFieldMappings() returned %d items, want %d", len(result), tt.expected)
			}

			// Check first mapping
			if len(result) > 0 {
				firstMapping := result[0].(map[string]interface{})
				if firstMapping["from"] != tt.input[0].From {
					t.Errorf("FlattenFieldMappings() first mapping from = %v, want %v", firstMapping["from"], tt.input[0].From)
				}
				if firstMapping["to"] != tt.input[0].To {
					t.Errorf("FlattenFieldMappings() first mapping to = %v, want %v", firstMapping["to"], tt.input[0].To)
				}
			}
		})
	}
}

func TestFlattenFieldMappings_Empty(t *testing.T) {
	result := provider.FlattenFieldMappings([]client.FieldMapping{})
	if len(result) != 0 {
		t.Errorf("FlattenFieldMappings(empty) should return empty slice, got %d items", len(result))
	}
}

// ============================================================================
// Alert Tests
// ============================================================================

func TestExpandAlerts_Basic(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected int // Check length since AlertAttribute has Options map
	}{
		{
			name: "basic alert",
			input: []interface{}{
				map[string]interface{}{
					"type":                  "email",
					"send_for":              "failure",
					"should_send_recovery":  true,
					"emails":                []interface{}{"admin@example.com"},
				},
			},
			expected: 1,
		},
		{
			name: "multiple alerts",
			input: []interface{}{
				map[string]interface{}{
					"type":                 "email",
					"send_for":             "failure",
					"should_send_recovery": true,
					"emails":               []interface{}{"failure@example.com"},
				},
				map[string]interface{}{
					"type":                 "email",
					"send_for":             "success",
					"should_send_recovery": false,
					"emails":               []interface{}{"success@example.com"},
				},
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.ExpandAlerts(tt.input)
			if len(result) != tt.expected {
				t.Errorf("ExpandAlerts() returned %d items, want %d", len(result), tt.expected)
			}
		})
	}
}

func TestExpandAlerts_Empty(t *testing.T) {
	result := provider.ExpandAlerts([]interface{}{})
	if result != nil {
		t.Errorf("ExpandAlerts(empty) should return nil, got %d items", len(result))
	}
}

// ============================================================================
// Schedule Tests
// ============================================================================

func TestExpandSyncSchedule_Hourly(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		wantFreq string
	}{
		{
			name: "hourly schedule with minute",
			input: []interface{}{
				map[string]interface{}{
					"frequency": "hourly",
					"minute":    30,
				},
			},
			wantFreq: "hourly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.ExpandSyncSchedule(tt.input)
			if result == nil {
				t.Errorf("ExpandSyncSchedule() returned nil")
				return
			}
			if result.Frequency != tt.wantFreq {
				t.Errorf("ExpandSyncSchedule() frequency = %v, want %v", result.Frequency, tt.wantFreq)
			}
		})
	}
}

func TestExpandSyncSchedule_Daily(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		wantFreq string
	}{
		{
			name: "daily schedule at 9am",
			input: []interface{}{
				map[string]interface{}{
					"frequency": "daily",
					"hour":      9,
					"minute":    0,
				},
			},
			wantFreq: "daily",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.ExpandSyncSchedule(tt.input)
			if result == nil {
				t.Errorf("ExpandSyncSchedule() returned nil")
				return
			}
			if result.Frequency != tt.wantFreq {
				t.Errorf("ExpandSyncSchedule() frequency = %v, want %v", result.Frequency, tt.wantFreq)
			}
		})
	}
}

func TestExpandSyncSchedule_Nil(t *testing.T) {
	result := provider.ExpandSyncSchedule([]interface{}{})
	if result != nil {
		t.Errorf("ExpandSyncSchedule(empty) should return nil, got %+v", result)
	}
}

// ============================================================================
// Utility Helper Tests
// ============================================================================

func TestExpandStringList(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected []string
	}{
		{
			name:     "empty list",
			input:    []interface{}{},
			expected: []string{},
		},
		{
			name:     "single item",
			input:    []interface{}{"item1"},
			expected: []string{"item1"},
		},
		{
			name:     "multiple items",
			input:    []interface{}{"item1", "item2", "item3"},
			expected: []string{"item1", "item2", "item3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.ExpandStringList(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ExpandStringList() got = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestExpandStringMap(t *testing.T) {
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
			name: "simple map",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			expected: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name: "map with empty string values - kept as-is",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": "",
				"key3": "value3",
			},
			expected: map[string]interface{}{
				"key1": "value1",
				"key2": "",
				"key3": "value3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.ExpandStringMap(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ExpandStringMap() got = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

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
			result := provider.CleanEmptyStrings(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("CleanEmptyStrings() got = %+v, want %+v", result, tt.expected)
			}
		})
	}
}
