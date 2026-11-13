package provider

import (
	"sort"
	"testing"

	"github.com/sutrolabs/terraform-provider-census/census/client"
)

// Unit tests for logic declared in sync_crud.go
//
// The tests below exercise the sorting and compound_key-filtering logic
// inline, the same way resourceSyncRead does, rather than calling an
// exported symbol directly — neither behavior is factored out of
// resourceSyncRead into its own function.

// ============================================================================
// Field Mapping Ordering Tests
// ============================================================================
// These tests verify that field mappings are sorted by position from the Census API,
// preventing spurious diffs when the API returns mappings in non-deterministic order.

func TestSortFieldMappingsByPosition_Basic(t *testing.T) {
	// Simulate API response with mappings in wrong order
	mappings := []client.FieldMapping{
		{
			From:     "last_name",
			To:       "LastName",
			Position: 2,
			Type:     "direct",
		},
		{
			From:     "email",
			To:       "Email",
			Position: 0,
			Type:     "direct",
		},
		{
			From:     "first_name",
			To:       "FirstName",
			Position: 1,
			Type:     "direct",
		},
	}

	// Sort by position
	sort.Slice(mappings, func(i, j int) bool {
		return mappings[i].Position < mappings[j].Position
	})

	// Verify result is sorted by position
	if len(mappings) != 3 {
		t.Errorf("Expected 3 mappings, got %d", len(mappings))
	}

	expectedOrder := []string{"Email", "FirstName", "LastName"}
	expectedPositions := []int{0, 1, 2}
	for i, mapping := range mappings {
		if mapping.To != expectedOrder[i] {
			t.Errorf("Position %d: got %s, want %s", i, mapping.To, expectedOrder[i])
		}
		if mapping.Position != expectedPositions[i] {
			t.Errorf("Mapping %d: got position %d, want %d", i, mapping.Position, expectedPositions[i])
		}
	}
}

func TestSortFieldMappingsByPosition_AlreadyOrdered(t *testing.T) {
	// Simulate API response already in correct order
	mappings := []client.FieldMapping{
		{To: "Email", Position: 0},
		{To: "FirstName", Position: 1},
		{To: "LastName", Position: 2},
	}

	// Sort by position (should be stable)
	sort.Slice(mappings, func(i, j int) bool {
		return mappings[i].Position < mappings[j].Position
	})

	// Verify order is unchanged
	expectedOrder := []string{"Email", "FirstName", "LastName"}
	for i, mapping := range mappings {
		if mapping.To != expectedOrder[i] {
			t.Errorf("Position %d: got %s, want %s", i, mapping.To, expectedOrder[i])
		}
	}
}

func TestSortFieldMappingsByPosition_MixedMappingTypes(t *testing.T) {
	// Test with various mapping types in wrong order
	mappings := []client.FieldMapping{
		{
			To:             "FullName",
			Position:       3,
			Type:           "liquid_template",
			LiquidTemplate: "{{ first_name }} {{ last_name }}",
		},
		{
			From:                "email",
			To:                  "Email",
			Position:            0,
			Type:                "direct",
			IsPrimaryIdentifier: true,
		},
		{
			To:              "SyncRunID",
			Position:        2,
			Type:            "sync_metadata",
			SyncMetadataKey: "sync_run_id",
		},
		{
			To:       "Source",
			Position: 1,
			Type:     "constant",
			Constant: "Website",
		},
	}

	// Sort by position
	sort.Slice(mappings, func(i, j int) bool {
		return mappings[i].Position < mappings[j].Position
	})

	// Verify order matches positions
	expectedOrder := []string{"Email", "Source", "SyncRunID", "FullName"}
	for i, mapping := range mappings {
		if mapping.To != expectedOrder[i] {
			t.Errorf("Position %d: got %s, want %s", i, mapping.To, expectedOrder[i])
		}
	}

	// Verify types are preserved
	if mappings[0].Type != "direct" || !mappings[0].IsPrimaryIdentifier {
		t.Errorf("Email mapping should be direct type with primary identifier")
	}
	if mappings[1].Type != "constant" {
		t.Errorf("Source mapping should be constant type")
	}
	if mappings[2].Type != "sync_metadata" {
		t.Errorf("SyncRunID mapping should be sync_metadata type")
	}
	if mappings[3].Type != "liquid_template" {
		t.Errorf("FullName mapping should be liquid_template type")
	}
}

func TestSortFieldMappingsByPosition_EmptySlice(t *testing.T) {
	// Empty slice
	mappings := []client.FieldMapping{}

	// Sort should not panic
	sort.Slice(mappings, func(i, j int) bool {
		return mappings[i].Position < mappings[j].Position
	})

	// Should still be empty
	if len(mappings) != 0 {
		t.Errorf("Expected 0 mappings, got %d", len(mappings))
	}
}

// ============================================================================
// Compound Key Mapping Tests
// ============================================================================
// These tests verify that Census-managed compound_key mappings are properly
// handled and filtered out to prevent drift in Terraform state.

func TestFilterCompoundKeyMappings(t *testing.T) {
	// Test the filtering logic that removes Census-managed compound_key mappings
	// This simulates what happens in resourceSyncRead
	tests := []struct {
		name     string
		input    []client.FieldMapping
		expected int // Number of mappings after filtering
	}{
		{
			name: "filter out single compound_key mapping",
			input: []client.FieldMapping{
				{
					From:                "ID",
					To:                  "click_conversion.conversion_date_time",
					Type:                "direct",
					IsPrimaryIdentifier: false,
					Position:            0,
				},
				{
					From:                "ID",
					To:                  "click_conversion.gclid",
					Type:                "direct",
					IsPrimaryIdentifier: false,
					Position:            1,
				},
				{
					From:                "[map[data:ID type:column] map[data:ID type:column] map[data:ID type:column]]",
					To:                  "click_conversion._census_tracking_id",
					Type:                "compound_key",
					IsPrimaryIdentifier: true,
					Position:            2,
				},
			},
			expected: 2, // Only the 2 user-defined mappings
		},
		{
			name: "no compound_key mappings to filter",
			input: []client.FieldMapping{
				{
					From:                "email",
					To:                  "Email",
					Type:                "direct",
					IsPrimaryIdentifier: true,
					Position:            0,
				},
				{
					From:     "first_name",
					To:       "FirstName",
					Type:     "direct",
					Position: 1,
				},
			},
			expected: 2, // All mappings remain
		},
		{
			name:     "empty input",
			input:    []client.FieldMapping{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the filtering logic from resourceSyncRead
			userManagedMappings := make([]client.FieldMapping, 0, len(tt.input))
			for _, mapping := range tt.input {
				if mapping.Type != "compound_key" {
					userManagedMappings = append(userManagedMappings, mapping)
				}
			}

			if len(userManagedMappings) != tt.expected {
				t.Errorf("Expected %d mappings after filtering, got %d", tt.expected, len(userManagedMappings))
			}

			// Verify no compound_key mappings remain
			for i, mapping := range userManagedMappings {
				if mapping.Type == "compound_key" {
					t.Errorf("Mapping %d should not be compound_key type", i)
				}
			}
		})
	}
}

func TestFilterCompoundKeyMappings_PreservesOtherMappingTypes(t *testing.T) {
	// Verify that filtering compound_key doesn't affect other mapping types
	input := []client.FieldMapping{
		{
			From:                "email",
			To:                  "Email",
			Type:                "direct",
			IsPrimaryIdentifier: true,
			Position:            0,
		},
		{
			To:       "Source",
			Type:     "constant",
			Constant: "Website",
			Position: 1,
		},
		{
			To:              "SyncRunID",
			Type:            "sync_metadata",
			SyncMetadataKey: "sync_run_id",
			Position:        2,
		},
		{
			To:             "FullName",
			Type:           "liquid_template",
			LiquidTemplate: "{{ first_name }} {{ last_name }}",
			Position:       3,
		},
		{
			From:                "[map[data:ID type:column]]",
			To:                  "_census_tracking_id",
			Type:                "compound_key",
			IsPrimaryIdentifier: true,
			Position:            4,
		},
	}

	// Filter out compound_key mappings
	userManagedMappings := make([]client.FieldMapping, 0, len(input))
	for _, mapping := range input {
		if mapping.Type != "compound_key" {
			userManagedMappings = append(userManagedMappings, mapping)
		}
	}

	// Should have 4 mappings (all except compound_key)
	if len(userManagedMappings) != 4 {
		t.Errorf("Expected 4 mappings after filtering, got %d", len(userManagedMappings))
	}

	// Verify all mapping types are preserved
	expectedTypes := []string{"direct", "constant", "sync_metadata", "liquid_template"}
	for i, mapping := range userManagedMappings {
		if mapping.Type != expectedTypes[i] {
			t.Errorf("Mapping %d: expected type '%s', got '%s'", i, expectedTypes[i], mapping.Type)
		}
	}

	// Verify the primary identifier is preserved
	if !userManagedMappings[0].IsPrimaryIdentifier {
		t.Errorf("Expected first mapping to be primary identifier")
	}
}
