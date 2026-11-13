package provider

import (
	"fmt"

	"github.com/sutrolabs/terraform-provider-census/census/client"
)

// This file bridges the two parallel field-mapping representations that
// currently exist for historical reasons: client.FieldMapping (Terraform's
// shape) and client.MappingAttributes (the Census API's shape). Stage 5 of
// the v2 migration collapses these into one type and deletes this file
// outright, rather than extracting pieces of it — kept isolated here, apart
// from sync_expand.go/sync_flatten.go, so that future deletion is a clean
// single-file removal.

// ConvertFieldMappingsToMappingAttributes converts Terraform FieldMapping structs to Census API MappingAttributes format
func ConvertFieldMappingsToMappingAttributes(fieldMappings []client.FieldMapping) []client.MappingAttributes {
	result := make([]client.MappingAttributes, len(fieldMappings))

	for i, fm := range fieldMappings {
		// Convert based on type
		var mappingFrom client.MappingFrom

		switch fm.Type {
		case "constant":
			// Format constant values as required by Census API
			constantData := map[string]interface{}{
				"basic_type": "text", // Default to text type for string constants
				"value":      fmt.Sprintf("%v", fm.Constant),
			}
			mappingFrom = client.MappingFrom{
				Type: "constant_value",
				Data: constantData,
			}

		case "sync_metadata":
			// Sync metadata mapping (e.g., sync_run_id)
			mappingFrom = client.MappingFrom{
				Type: "sync_metadata",
				Data: fm.SyncMetadataKey, // e.g., "sync_run_id"
			}

		case "segment_membership":
			// Segment membership mapping
			segmentData := map[string]interface{}{
				"identify_by": fm.SegmentIdentifyBy,
			}
			mappingFrom = client.MappingFrom{
				Type: "segment_membership",
				Data: segmentData,
			}

		case "liquid_template":
			// Liquid template transformation - Census API expects the template string directly
			mappingFrom = client.MappingFrom{
				Type: "liquid_template",
				Data: fm.LiquidTemplate,
			}

		default:
			// Default to column mapping (direct)
			mappingFrom = client.MappingFrom{
				Type: "column",
				Data: fm.From,
			}
		}

		result[i] = client.MappingAttributes{
			From:                mappingFrom,
			To:                  fm.To,
			IsPrimaryIdentifier: fm.IsPrimaryIdentifier,
			LookupObject:        fm.LookupObject,
			LookupField:         fm.LookupField,
			PreserveValues:      fm.PreserveValues,
			GenerateField:       fm.GenerateField,
			SyncNullValues:      fm.SyncNullValues,
			ArrayField:          fm.ArrayField,
			FieldType:           fm.FieldType,
			FollowSourceType:    fm.FollowSourceType,
		}
	}

	return result
}

// ConvertMappingAttributesToFieldMappings converts Census API MappingAttributes back to Terraform FieldMapping
func ConvertMappingAttributesToFieldMappings(mappings []client.MappingAttributes) []client.FieldMapping {
	if mappings == nil {
		return []client.FieldMapping{}
	}

	result := make([]client.FieldMapping, len(mappings))

	for i, ma := range mappings {
		var mappingType string
		var constant interface{}
		var from string
		var syncMetadataKey string
		var segmentIdentifyBy string
		var liquidTemplate string

		// Convert based on mapping from type - add nil checks
		if ma.From.Data == nil {
			// Handle nil data gracefully
			from = ""
			mappingType = "direct"
		} else {
			switch ma.From.Type {
			case "constant_value":
				mappingType = "constant"
				// Extract value from the data map structure
				// Census API returns: {"value": "...", "basic_type": "text"}
				if dataMap, ok := ma.From.Data.(map[string]interface{}); ok {
					if value, ok := dataMap["value"]; ok {
						constant = value // Extract just the value
					} else {
						constant = ma.From.Data // Fallback to full data if no value field
					}
				} else {
					constant = ma.From.Data // Fallback if not a map
				}
				from = "" // Empty from field for constants

			case "sync_metadata":
				mappingType = "sync_metadata"
				// Census API returns: "sync_run_id" as string
				if dataStr, ok := ma.From.Data.(string); ok {
					syncMetadataKey = dataStr
				} else {
					syncMetadataKey = fmt.Sprintf("%v", ma.From.Data)
				}
				from = "" // Empty from field for sync_metadata

			case "segment_membership":
				mappingType = "segment_membership"
				// Census API returns: {"identify_by": "name"}
				if dataMap, ok := ma.From.Data.(map[string]interface{}); ok {
					if identifyBy, ok := dataMap["identify_by"].(string); ok {
						segmentIdentifyBy = identifyBy
					}
				}
				from = "" // Empty from field for segment_membership

			case "liquid_template":
				mappingType = "liquid_template"
				// Handle both formats:
				// - API returns: {"liquid_template": "..."} (hash)
				// - We send: "..." (string)
				if dataStr, ok := ma.From.Data.(string); ok {
					liquidTemplate = dataStr
				} else if dataMap, ok := ma.From.Data.(map[string]interface{}); ok {
					if template, ok := dataMap["liquid_template"].(string); ok {
						liquidTemplate = template
					}
				}
				from = "" // Empty from field for liquid_template

			case "compound_key":
				// Census-managed compound key mappings (auto-generated for certain destinations)
				// These should be filtered out during Read to avoid drift
				mappingType = "compound_key"
				// Store the from data for completeness, though this will be filtered
				if dataStr, ok := ma.From.Data.(string); ok {
					from = dataStr
				} else {
					from = fmt.Sprintf("%v", ma.From.Data)
				}

			default: // "column"
				mappingType = "direct"
				if dataStr, ok := ma.From.Data.(string); ok {
					from = dataStr
				} else {
					from = fmt.Sprintf("%v", ma.From.Data)
				}
			}
		}

		result[i] = client.FieldMapping{
			From:                from,
			To:                  ma.To,
			Position:            ma.Position,
			Type:                mappingType,
			Constant:            constant,
			SyncMetadataKey:     syncMetadataKey,
			SegmentIdentifyBy:   segmentIdentifyBy,
			LiquidTemplate:      liquidTemplate,
			IsPrimaryIdentifier: ma.IsPrimaryIdentifier,
			LookupObject:        ma.LookupObject,
			LookupField:         ma.LookupField,
			PreserveValues:      ma.PreserveValues,
			GenerateField:       ma.GenerateField,
			SyncNullValues:      ma.SyncNullValues,
		}
	}

	return result
}
