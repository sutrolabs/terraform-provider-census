package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/sutrolabs/terraform-provider-census/census/client/v2"
)

// Helper functions for flattening complex types (API response -> Terraform state)

// FlattenFieldMappings converts the API's MappingAttributes directly into
// the Terraform field_mapping blocks' flat from/to/type shape — the reverse
// of ExpandFieldMappings, and likewise a single step rather than two (see
// that function's comment for why there's no intermediate struct anymore).
func FlattenFieldMappings(mappings []v2.MappingAttributes) []interface{} {
	result := make([]interface{}, len(mappings))
	for i, ma := range mappings {
		var mappingType string
		var constant interface{}
		var from string
		var syncMetadataKey string
		var segmentIdentifyBy string
		var liquidTemplate string

		if ma.From.Data == nil {
			mappingType = "direct"
		} else {
			switch ma.From.Type {
			case "constant_value":
				mappingType = "constant"
				// Census API returns: {"value": "...", "basic_type": "text"}
				if dataMap, ok := ma.From.Data.(map[string]interface{}); ok {
					if value, ok := dataMap["value"]; ok {
						constant = value
					} else {
						constant = ma.From.Data
					}
				} else {
					constant = ma.From.Data
				}

			case "sync_metadata":
				mappingType = "sync_metadata"
				if dataStr, ok := ma.From.Data.(string); ok {
					syncMetadataKey = dataStr
				} else {
					syncMetadataKey = fmt.Sprintf("%v", ma.From.Data)
				}

			case "segment_membership":
				mappingType = "segment_membership"
				if dataMap, ok := ma.From.Data.(map[string]interface{}); ok {
					if identifyBy, ok := dataMap["identify_by"].(string); ok {
						segmentIdentifyBy = identifyBy
					}
				}

			case "liquid_template":
				mappingType = "liquid_template"
				if dataStr, ok := ma.From.Data.(string); ok {
					liquidTemplate = dataStr
				} else if dataMap, ok := ma.From.Data.(map[string]interface{}); ok {
					if template, ok := dataMap["liquid_template"].(string); ok {
						liquidTemplate = template
					}
				}

			case "compound_key":
				// Census-managed compound key mappings (auto-generated for certain
				// destinations) — callers are expected to filter these out before
				// flattening; kept here for completeness if one slips through.
				mappingType = "compound_key"
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

		mappingMap := map[string]interface{}{
			"from":                  from,
			"to":                    ma.To,
			"type":                  mappingType,
			"constant":              constant,
			"sync_metadata_key":     syncMetadataKey,
			"segment_identify_by":   segmentIdentifyBy,
			"liquid_template":       liquidTemplate,
			"is_primary_identifier": ma.IsPrimaryIdentifier,
			"operation":             ma.Operation,
			"lookup_object":         ma.LookupObject,
			"lookup_field":          ma.LookupField,
			"preserve_values":       ma.PreserveValues,
			"generate_field":        ma.GenerateField,
			"array_field":           ma.ArrayField,
			"field_type":            ma.FieldType,
			"follow_source_type":    ma.FollowSourceType,
		}

		// Always include sync_null_values explicitly to match API response
		if ma.SyncNullValues != nil {
			mappingMap["sync_null_values"] = *ma.SyncNullValues
		} else {
			// If API didn't return it, use schema default
			mappingMap["sync_null_values"] = true
		}

		result[i] = mappingMap
	}
	return result
}

func FlattenAlerts(alerts []v2.AlertAttribute) []interface{} {
	if len(alerts) == 0 {
		return []interface{}{}
	}

	result := make([]interface{}, 0, len(alerts))
	for _, alert := range alerts {
		// Convert options to string map for Terraform
		options := make(map[string]interface{})
		for key, value := range alert.Options {
			switch v := value.(type) {
			case int:
				options[key] = strconv.Itoa(v)
			case float64:
				options[key] = strconv.Itoa(int(v))
			case string:
				options[key] = v
			default:
				options[key] = fmt.Sprintf("%v", v)
			}
		}

		result = append(result, map[string]interface{}{
			"id":                   alert.ID,
			"type":                 alert.Type,
			"send_for":             alert.SendFor,
			"should_send_recovery": alert.ShouldSendRecovery,
			"options":              options,
		})
	}
	return result
}

// FlattenRunMode converts the API's SyncMode struct to Terraform run_mode config
func FlattenRunMode(mode *v2.SyncMode) []interface{} {
	if mode == nil {
		return []interface{}{}
	}

	modeMap := map[string]interface{}{
		"type": mode.Type,
	}

	// Flatten triggers if present
	if mode.Triggers != nil {
		triggersMap := map[string]interface{}{}

		// Flatten schedule trigger
		if mode.Triggers.Schedule != nil {
			scheduleMap := map[string]interface{}{
				"frequency": mode.Triggers.Schedule.Frequency,
			}

			if mode.Triggers.Schedule.CronExpression != "" {
				scheduleMap["cron_expression"] = mode.Triggers.Schedule.CronExpression
			} else {
				if mode.Triggers.Schedule.Day != "" {
					scheduleMap["day"] = mode.Triggers.Schedule.Day
				}
				if mode.Triggers.Schedule.Hour != 0 {
					scheduleMap["hour"] = mode.Triggers.Schedule.Hour
				}
				if mode.Triggers.Schedule.Minute != 0 {
					scheduleMap["minute"] = mode.Triggers.Schedule.Minute
				}
			}

			triggersMap["schedule"] = []interface{}{scheduleMap}
		}

		// Flatten dbt_cloud trigger
		if mode.Triggers.DbtCloud != nil {
			dbtMap := map[string]interface{}{
				"project_id": mode.Triggers.DbtCloud.ProjectID,
				"job_id":     mode.Triggers.DbtCloud.JobID,
			}
			triggersMap["dbt_cloud"] = []interface{}{dbtMap}
		}

		// Flatten fivetran trigger
		if mode.Triggers.Fivetran != nil {
			fivetranMap := map[string]interface{}{
				"job_id":   mode.Triggers.Fivetran.JobID,
				"job_name": mode.Triggers.Fivetran.JobName,
			}
			triggersMap["fivetran"] = []interface{}{fivetranMap}
		}

		// Flatten sync_sequence trigger
		if mode.Triggers.SyncSequence != nil {
			seqMap := map[string]interface{}{
				"sync_id": mode.Triggers.SyncSequence.SyncID,
			}
			triggersMap["sync_sequence"] = []interface{}{seqMap}
		}

		modeMap["triggers"] = []interface{}{triggersMap}
	}

	return []interface{}{modeMap}
}

func FlattenAdvancedConfiguration(ctx context.Context, m map[string]interface{}) string {
	if m == nil || len(m) == 0 {
		return ""
	}
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		tflog.Debug(ctx, "Failed to marshal advanced_configuration", map[string]interface{}{"error": err.Error()})
		return ""
	}
	return string(jsonBytes)
}

// FlattenSourceAttributes converts API source_attributes map to Terraform list structure
func FlattenSourceAttributes(attrs map[string]interface{}) []map[string]interface{} {
	if attrs == nil {
		return nil
	}

	result := make(map[string]interface{})

	// Set cohort_id if present (for cohort sources)
	if cohortId, ok := attrs["cohort_id"]; ok {
		switch v := cohortId.(type) {
		case float64:
			result["cohort_id"] = int(v)
		case int:
			result["cohort_id"] = v
		default:
			result["cohort_id"] = cohortId
		}
	}

	// Handle object structure with smart translation
	if objectData, ok := attrs["object"]; ok {
		objectMap, isMap := objectData.(map[string]interface{})
		if isMap {
			convertedObject := make(map[string]interface{})

			// Get the object type from API response
			objectType := ""
			if t, ok := objectMap["type"]; ok {
				objectType = convertToString(t)
			}

			// Check for special source types that need reverse translation
			filterSegmentId, hasFilterSegment := attrs["filter_segment_id"]
			cohortId, hasCohort := attrs["cohort_id"]

			if hasFilterSegment && objectType == "filter_segment_source" {
				// API returns filter_segment_source - translate to segment for user
				convertedObject["type"] = "segment"
				// Use filter_segment_id as the object id
				switch v := filterSegmentId.(type) {
				case float64:
					convertedObject["id"] = fmt.Sprintf("%.0f", v)
				case int:
					convertedObject["id"] = fmt.Sprintf("%d", v)
				default:
					convertedObject["id"] = convertToString(filterSegmentId)
				}
				// Include dataset_id from the object
				if datasetId, ok := objectMap["dataset_id"]; ok {
					convertedObject["dataset_id"] = convertToString(datasetId)
				}
			} else if hasCohort && objectType == "cohort_source" {
				// API returns cohort_source - translate to cohort for user
				convertedObject["type"] = "cohort"
				// Use cohort_id as the object id
				switch v := cohortId.(type) {
				case float64:
					convertedObject["id"] = fmt.Sprintf("%.0f", v)
				case int:
					convertedObject["id"] = fmt.Sprintf("%d", v)
				default:
					convertedObject["id"] = convertToString(cohortId)
				}
				// Try to get dataset_id from the object
				if datasetId, ok := objectMap["dataset_id"]; ok {
					convertedObject["dataset_id"] = convertToString(datasetId)
				}
			} else if objectType == "business_object_source" {
				// Translate business_object_source -> dataset for Terraform
				convertedObject["type"] = "dataset"
				// For datasets, use dataset_id instead of id
				if datasetId, ok := objectMap["dataset_id"]; ok {
					convertedObject["id"] = convertToString(datasetId)
				}
			} else if objectType == "table" {
				// For table sources, include table identifiers but NOT id
				convertedObject["type"] = "table"
				if v, ok := objectMap["table_name"]; ok {
					convertedObject["table_name"] = convertToString(v)
				}
				if v, ok := objectMap["table_schema"]; ok {
					convertedObject["table_schema"] = convertToString(v)
				}
				if v, ok := objectMap["table_catalog"]; ok {
					convertedObject["table_catalog"] = convertToString(v)
				}
			} else {
				// For other types (model, topic, dataset), copy type and id if present
				if objectType != "" {
					convertedObject["type"] = objectType
				}
				if id, ok := objectMap["id"]; ok {
					convertedObject["id"] = convertToString(id)
				}
			}

			objectList := []map[string]interface{}{convertedObject}
			result["object"] = objectList
		}
	}

	return []map[string]interface{}{result}
}

// convertToString converts various types to string for Terraform compatibility
func convertToString(value interface{}) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case int:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case float64:
		// Check if it's actually an integer
		if v == float64(int64(v)) {
			return fmt.Sprintf("%.0f", v)
		}
		return fmt.Sprintf("%g", v)
	case bool:
		return fmt.Sprintf("%t", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// FlattenDestinationAttributes converts the API's typed destination_attributes
// into the Terraform list structure.
func FlattenDestinationAttributes(attrs *v2.SyncDestinationAttributes) []map[string]interface{} {
	if attrs == nil {
		return nil
	}

	result := map[string]interface{}{
		"object": attrs.Object,
	}
	if attrs.LeadUnionInsertTo != "" {
		result["lead_union_insert_to"] = attrs.LeadUnionInsertTo
	}

	return []map[string]interface{}{result}
}
