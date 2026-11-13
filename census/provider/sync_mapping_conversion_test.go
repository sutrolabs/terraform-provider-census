package provider

import (
	"reflect"
	"testing"

	"github.com/sutrolabs/terraform-provider-census/census/client"
)

// Unit tests for symbols declared in sync_mapping_conversion.go

// Helper function for tests
func boolPtr(b bool) *bool {
	return &b
}

// ============================================================================
// Mapping Conversion Tests (FieldMapping → MappingAttributes)
// ============================================================================

func TestConvertFieldMappingsToMappingAttributes_Constant(t *testing.T) {
	tests := []struct {
		name     string
		input    []client.FieldMapping
		expected []client.MappingAttributes
	}{
		{
			name: "single constant string mapping",
			input: []client.FieldMapping{
				{
					To:       "Source",
					Type:     "constant",
					Constant: "Website",
				},
			},
			expected: []client.MappingAttributes{
				{
					From: client.MappingFrom{
						Type: "constant_value",
						Data: map[string]interface{}{
							"basic_type": "text",
							"value":      "Website",
						},
					},
					To:                  "Source",
					IsPrimaryIdentifier: false,
				},
			},
		},
		{
			name: "multiple constant mappings with different to fields",
			input: []client.FieldMapping{
				{
					To:       "AssistantName",
					Type:     "constant",
					Constant: "HERE is my constant value",
				},
				{
					To:       "Title",
					Type:     "constant",
					Constant: "HERE is my constant value 2",
				},
			},
			expected: []client.MappingAttributes{
				{
					From: client.MappingFrom{
						Type: "constant_value",
						Data: map[string]interface{}{
							"basic_type": "text",
							"value":      "HERE is my constant value",
						},
					},
					To:                  "AssistantName",
					IsPrimaryIdentifier: false,
				},
				{
					From: client.MappingFrom{
						Type: "constant_value",
						Data: map[string]interface{}{
							"basic_type": "text",
							"value":      "HERE is my constant value 2",
						},
					},
					To:                  "Title",
					IsPrimaryIdentifier: false,
				},
			},
		},
		{
			name: "constant numeric value",
			input: []client.FieldMapping{
				{
					To:       "Priority",
					Type:     "constant",
					Constant: 1,
				},
			},
			expected: []client.MappingAttributes{
				{
					From: client.MappingFrom{
						Type: "constant_value",
						Data: map[string]interface{}{
							"basic_type": "text",
							"value":      "1",
						},
					},
					To:                  "Priority",
					IsPrimaryIdentifier: false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertFieldMappingsToMappingAttributes(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("convertFieldMappingsToMappingAttributes() got = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestConvertFieldMappingsToMappingAttributes_Direct(t *testing.T) {
	tests := []struct {
		name     string
		input    []client.FieldMapping
		expected []client.MappingAttributes
	}{
		{
			name: "direct mapping",
			input: []client.FieldMapping{
				{
					From: "user_id",
					To:   "UserID",
					Type: "direct",
				},
			},
			expected: []client.MappingAttributes{
				{
					From: client.MappingFrom{
						Type: "column",
						Data: "user_id", // Census API expects just the column name as string
					},
					To:                  "UserID",
					IsPrimaryIdentifier: false,
				},
			},
		},
		{
			name: "direct mapping with primary identifier",
			input: []client.FieldMapping{
				{
					From:                "user_id",
					To:                  "UserID",
					Type:                "direct",
					IsPrimaryIdentifier: true,
				},
			},
			expected: []client.MappingAttributes{
				{
					From: client.MappingFrom{
						Type: "column",
						Data: "user_id", // Census API expects just the column name as string
					},
					To:                  "UserID",
					IsPrimaryIdentifier: true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertFieldMappingsToMappingAttributes(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("convertFieldMappingsToMappingAttributes() got = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestConvertFieldMappingsToMappingAttributes_Mixed(t *testing.T) {
	// Test a realistic scenario with multiple mapping types
	input := []client.FieldMapping{
		{
			From:                "user_id",
			To:                  "UserID",
			Type:                "direct",
			IsPrimaryIdentifier: true,
		},
		{
			From: "email",
			To:   "Email",
			Type: "direct",
		},
		{
			To:       "Source",
			Type:     "constant",
			Constant: "Website",
		},
		{
			To:              "SyncTime",
			Type:            "sync_metadata",
			SyncMetadataKey: "sync_run_id",
		},
	}

	expected := []client.MappingAttributes{
		{
			From: client.MappingFrom{
				Type: "column",
				Data: "user_id", // Census API expects string for column
			},
			To:                  "UserID",
			IsPrimaryIdentifier: true,
		},
		{
			From: client.MappingFrom{
				Type: "column",
				Data: "email", // Census API expects string for column
			},
			To:                  "Email",
			IsPrimaryIdentifier: false,
		},
		{
			From: client.MappingFrom{
				Type: "constant_value",
				Data: map[string]interface{}{
					"basic_type": "text",
					"value":      "Website",
				},
			},
			To:                  "Source",
			IsPrimaryIdentifier: false,
		},
		{
			From: client.MappingFrom{
				Type: "sync_metadata",
				Data: "sync_run_id", // Census API expects string for sync_metadata
			},
			To:                  "SyncTime",
			IsPrimaryIdentifier: false,
		},
	}

	result := ConvertFieldMappingsToMappingAttributes(input)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("convertFieldMappingsToMappingAttributes() got = %+v, want %+v", result, expected)
	}
}

func TestConvertFieldMappingsToMappingAttributes_LiquidTemplate(t *testing.T) {
	// Test liquid template format
	input := []client.FieldMapping{
		{
			To:             "FormattedDate",
			Type:           "liquid_template",
			LiquidTemplate: "{{ record['date'] | upcase }}",
		},
	}

	expected := []client.MappingAttributes{
		{
			From: client.MappingFrom{
				Type: "liquid_template",
				Data: "{{ record['date'] | upcase }}", // Census API expects string directly, not wrapped in hash
			},
			To:                  "FormattedDate",
			IsPrimaryIdentifier: false,
		},
	}

	result := ConvertFieldMappingsToMappingAttributes(input)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("convertFieldMappingsToMappingAttributes() got = %+v, want %+v", result, expected)
	}
}

// ============================================================================
// Reverse Mapping Conversion Tests (MappingAttributes → FieldMapping)
// ============================================================================

func TestConvertMappingAttributesToFieldMappings_LiquidTemplate(t *testing.T) {
	// Test that we correctly extract liquid template from API response
	// The Census API returns liquid templates as: {"liquid_template": "..."}
	input := []client.MappingAttributes{
		{
			From: client.MappingFrom{
				Type: "liquid_template",
				Data: map[string]interface{}{
					"liquid_template": "{{ record['date'] | upcase }}",
				},
			},
			To:                  "FormattedDate",
			IsPrimaryIdentifier: false,
			SyncNullValues:      boolPtr(true),
		},
	}

	expected := []client.FieldMapping{
		{
			From:                "",
			To:                  "FormattedDate",
			Type:                "liquid_template",
			LiquidTemplate:      "{{ record['date'] | upcase }}",
			IsPrimaryIdentifier: false,
			SyncNullValues:      boolPtr(true),
		},
	}

	result := ConvertMappingAttributesToFieldMappings(input)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("ConvertMappingAttributesToFieldMappings() got = %+v, want %+v", result, expected)
	}
}

func TestConvertMappingAttributesToFieldMappings_Constant(t *testing.T) {
	// Test constant value extraction from API
	input := []client.MappingAttributes{
		{
			From: client.MappingFrom{
				Type: "constant_value",
				Data: map[string]interface{}{
					"value":      "Website",
					"basic_type": "text",
				},
			},
			To:                  "Source",
			IsPrimaryIdentifier: false,
			SyncNullValues:      boolPtr(true),
		},
	}

	expected := []client.FieldMapping{
		{
			From:                "",
			To:                  "Source",
			Type:                "constant",
			Constant:            "Website",
			IsPrimaryIdentifier: false,
			SyncNullValues:      boolPtr(true),
		},
	}

	result := ConvertMappingAttributesToFieldMappings(input)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("ConvertMappingAttributesToFieldMappings() got = %+v, want %+v", result, expected)
	}
}

func TestConvertMappingAttributesToFieldMappings_SyncMetadata(t *testing.T) {
	// Test sync_metadata extraction from API (returned as string)
	input := []client.MappingAttributes{
		{
			From: client.MappingFrom{
				Type: "sync_metadata",
				Data: "sync_run_id",
			},
			To:                  "SyncRunID",
			IsPrimaryIdentifier: false,
			SyncNullValues:      boolPtr(true),
		},
	}

	expected := []client.FieldMapping{
		{
			From:                "",
			To:                  "SyncRunID",
			Type:                "sync_metadata",
			SyncMetadataKey:     "sync_run_id",
			IsPrimaryIdentifier: false,
			SyncNullValues:      boolPtr(true),
		},
	}

	result := ConvertMappingAttributesToFieldMappings(input)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("ConvertMappingAttributesToFieldMappings() got = %+v, want %+v", result, expected)
	}
}

func TestConvertMappingAttributesToFieldMappings_SegmentMembership(t *testing.T) {
	// Test segment_membership extraction from API
	input := []client.MappingAttributes{
		{
			From: client.MappingFrom{
				Type: "segment_membership",
				Data: map[string]interface{}{
					"identify_by": "name",
				},
			},
			To:                  "SegmentField",
			IsPrimaryIdentifier: false,
			SyncNullValues:      boolPtr(true),
		},
	}

	expected := []client.FieldMapping{
		{
			From:                "",
			To:                  "SegmentField",
			Type:                "segment_membership",
			SegmentIdentifyBy:   "name",
			IsPrimaryIdentifier: false,
			SyncNullValues:      boolPtr(true),
		},
	}

	result := ConvertMappingAttributesToFieldMappings(input)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("ConvertMappingAttributesToFieldMappings() got = %+v, want %+v", result, expected)
	}
}

func TestConvertMappingAttributesToFieldMappings_RoundTrip(t *testing.T) {
	// Test full round-trip: FieldMapping → MappingAttributes → FieldMapping
	original := []client.FieldMapping{
		{
			From:                "email",
			To:                  "Email",
			Type:                "direct",
			IsPrimaryIdentifier: true,
			SyncNullValues:      boolPtr(true),
		},
		{
			To:             "FormattedDate",
			Type:           "liquid_template",
			LiquidTemplate: "{{ record['date'] | upcase }}",
			SyncNullValues: boolPtr(true),
		},
		{
			To:              "SyncTime",
			Type:            "sync_metadata",
			SyncMetadataKey: "sync_run_id",
			SyncNullValues:  boolPtr(true),
		},
		{
			To:             "Source",
			Type:           "constant",
			Constant:       "Website",
			SyncNullValues: boolPtr(true),
		},
	}

	// Convert to MappingAttributes (what we send to API)
	mappingAttrs := ConvertFieldMappingsToMappingAttributes(original)

	// Convert back to FieldMapping (what we read from API)
	result := ConvertMappingAttributesToFieldMappings(mappingAttrs)

	// Should match original (with "from" field empty for non-direct mappings)
	expected := []client.FieldMapping{
		{
			From:                "email",
			To:                  "Email",
			Type:                "direct",
			IsPrimaryIdentifier: true,
			SyncNullValues:      boolPtr(true),
		},
		{
			From:           "",
			To:             "FormattedDate",
			Type:           "liquid_template",
			LiquidTemplate: "{{ record['date'] | upcase }}",
			SyncNullValues: boolPtr(true),
		},
		{
			From:            "",
			To:              "SyncTime",
			Type:            "sync_metadata",
			SyncMetadataKey: "sync_run_id",
			SyncNullValues:  boolPtr(true),
		},
		{
			From:           "",
			To:             "Source",
			Type:           "constant",
			Constant:       "Website",
			SyncNullValues: boolPtr(true),
		},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Round-trip conversion failed.\nGot:      %+v\nExpected: %+v", result, expected)
	}
}

// ============================================================================
// Compound Key Mapping Tests
// ============================================================================
// These tests verify that Census-managed compound_key mappings are properly
// handled by the API-direction conversion (filtering them out is a separate
// concern tested alongside resourceSyncRead — see sync_crud_test.go).

func TestConvertMappingAttributesToFieldMappings_CompoundKey(t *testing.T) {
	// Test that compound_key type from API is properly converted
	// Census auto-generates these for destinations like Google Ads with compound keys
	input := []client.MappingAttributes{
		{
			From: client.MappingFrom{
				Type: "compound_key",
				Data: []interface{}{
					map[string]interface{}{
						"type": "column",
						"data": "ID",
					},
					map[string]interface{}{
						"type": "column",
						"data": "ID",
					},
					map[string]interface{}{
						"type": "column",
						"data": "ID",
					},
				},
			},
			To:                  "click_conversion._census_tracking_id",
			IsPrimaryIdentifier: true,
			SyncNullValues:      boolPtr(true),
		},
	}

	expected := []client.FieldMapping{
		{
			From:                "[map[data:ID type:column] map[data:ID type:column] map[data:ID type:column]]",
			To:                  "click_conversion._census_tracking_id",
			Type:                "compound_key",
			IsPrimaryIdentifier: true,
			SyncNullValues:      boolPtr(true),
		},
	}

	result := ConvertMappingAttributesToFieldMappings(input)
	if len(result) != len(expected) {
		t.Errorf("Expected %d mappings, got %d", len(expected), len(result))
	}
	if result[0].Type != "compound_key" {
		t.Errorf("Expected Type='compound_key', got '%s'", result[0].Type)
	}
	if result[0].To != expected[0].To {
		t.Errorf("Expected To='%s', got '%s'", expected[0].To, result[0].To)
	}
	if result[0].IsPrimaryIdentifier != expected[0].IsPrimaryIdentifier {
		t.Errorf("Expected IsPrimaryIdentifier=%v, got %v", expected[0].IsPrimaryIdentifier, result[0].IsPrimaryIdentifier)
	}
}

func TestConvertMappingAttributesToFieldMappings_WithCompoundKey(t *testing.T) {
	// Test conversion of mixed mappings including a Census-managed compound_key
	// This simulates the real API response for a Google Ads sync
	input := []client.MappingAttributes{
		{
			From: client.MappingFrom{
				Type: "column",
				Data: "ID",
			},
			To:                  "click_conversion.conversion_date_time",
			IsPrimaryIdentifier: false,
			Position:            0,
		},
		{
			From: client.MappingFrom{
				Type: "column",
				Data: "ID",
			},
			To:                  "click_conversion.gclid",
			IsPrimaryIdentifier: false,
			Position:            1,
		},
		{
			From: client.MappingFrom{
				Type: "column",
				Data: "ID",
			},
			To:                  "click_conversion.conversion_action",
			IsPrimaryIdentifier: false,
			Position:            2,
		},
		{
			From: client.MappingFrom{
				Type: "compound_key",
				Data: []interface{}{
					map[string]interface{}{
						"type": "column",
						"data": "ID",
					},
					map[string]interface{}{
						"type": "column",
						"data": "ID",
					},
					map[string]interface{}{
						"type": "column",
						"data": "ID",
					},
				},
			},
			To:                  "click_conversion._census_tracking_id",
			IsPrimaryIdentifier: true,
			Position:            3,
		},
	}

	result := ConvertMappingAttributesToFieldMappings(input)

	// Verify we have all 4 mappings
	if len(result) != 4 {
		t.Errorf("Expected 4 mappings, got %d", len(result))
	}

	// Verify the compound_key mapping is at the end
	compoundKeyMapping := result[3]
	if compoundKeyMapping.Type != "compound_key" {
		t.Errorf("Expected last mapping Type='compound_key', got '%s'", compoundKeyMapping.Type)
	}
	if compoundKeyMapping.To != "click_conversion._census_tracking_id" {
		t.Errorf("Expected To='click_conversion._census_tracking_id', got '%s'", compoundKeyMapping.To)
	}
	if !compoundKeyMapping.IsPrimaryIdentifier {
		t.Errorf("Expected compound_key mapping to be primary identifier")
	}

	// Verify the user-defined mappings are correct
	if result[0].Type != "direct" {
		t.Errorf("Expected mapping 0 Type='direct', got '%s'", result[0].Type)
	}
	if result[0].From != "ID" {
		t.Errorf("Expected mapping 0 From='ID', got '%s'", result[0].From)
	}
}
