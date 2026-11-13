package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/sutrolabs/terraform-provider-census/census/client"
)

// Helper functions for flattening complex types (API response -> Terraform state)

func FlattenFieldMappings(mappings []client.FieldMapping) []interface{} {
	result := make([]interface{}, len(mappings))
	for i, mapping := range mappings {
		mappingMap := map[string]interface{}{
			"from":                  mapping.From,
			"to":                    mapping.To,
			"type":                  mapping.Type,
			"constant":              mapping.Constant,
			"sync_metadata_key":     mapping.SyncMetadataKey,
			"segment_identify_by":   mapping.SegmentIdentifyBy,
			"liquid_template":       mapping.LiquidTemplate,
			"is_primary_identifier": mapping.IsPrimaryIdentifier,
			"lookup_object":         mapping.LookupObject,
			"lookup_field":          mapping.LookupField,
			"preserve_values":       mapping.PreserveValues,
			"generate_field":        mapping.GenerateField,
			"array_field":           mapping.ArrayField,
			"field_type":            mapping.FieldType,
			"follow_source_type":    mapping.FollowSourceType,
		}

		// Always include sync_null_values explicitly to match API response
		if mapping.SyncNullValues != nil {
			mappingMap["sync_null_values"] = *mapping.SyncNullValues
		} else {
			// If API didn't return it, use schema default
			mappingMap["sync_null_values"] = true
		}

		result[i] = mappingMap
	}
	return result
}

func FlattenAlerts(alerts []client.AlertAttribute) []interface{} {
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

// FlattenRunMode converts API SyncMode struct to Terraform run_mode config
func FlattenRunMode(mode *client.SyncMode) []interface{} {
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
				"project_id": mode.Triggers.DbtCloud.ProjectId,
				"job_id":     mode.Triggers.DbtCloud.JobId,
			}
			triggersMap["dbt_cloud"] = []interface{}{dbtMap}
		}

		// Flatten fivetran trigger
		if mode.Triggers.Fivetran != nil {
			fivetranMap := map[string]interface{}{
				"job_id":   mode.Triggers.Fivetran.JobId,
				"job_name": mode.Triggers.Fivetran.JobName,
			}
			triggersMap["fivetran"] = []interface{}{fivetranMap}
		}

		// Flatten sync_sequence trigger
		if mode.Triggers.SyncSequence != nil {
			seqMap := map[string]interface{}{
				"sync_id": mode.Triggers.SyncSequence.SyncId,
			}
			triggersMap["sync_sequence"] = []interface{}{seqMap}
		}

		modeMap["triggers"] = []interface{}{triggersMap}
	}

	return []interface{}{modeMap}
}

func FlattenStringMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return nil
	}
	result := make(map[string]interface{})
	for k, v := range m {
		// Convert numeric types to strings for Terraform TypeMap with TypeString elements
		switch val := v.(type) {
		case float64:
			result[k] = strconv.FormatFloat(val, 'f', -1, 64)
		case int:
			result[k] = strconv.Itoa(val)
		case int64:
			result[k] = strconv.FormatInt(val, 10)
		case map[string]interface{}:
			// For nested maps (like source_attributes.object), serialize to JSON
			jsonBytes, err := json.Marshal(val)
			if err != nil {
				result[k] = fmt.Sprintf("%v", val) // Fallback to string representation
			} else {
				result[k] = string(jsonBytes)
			}
		default:
			result[k] = v
		}
	}
	return result
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

	// Set connection_id as integer (schema expects TypeInt)
	if connectionId, ok := attrs["connection_id"]; ok {
		// Convert float64 to int if needed, otherwise use as-is
		switch v := connectionId.(type) {
		case float64:
			result["connection_id"] = int(v)
		case int:
			result["connection_id"] = v
		default:
			result["connection_id"] = connectionId
		}
	}

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

// FlattenDestinationAttributes converts API destination_attributes map to Terraform list structure
func FlattenDestinationAttributes(attrs map[string]interface{}) []map[string]interface{} {
	if attrs == nil {
		return nil
	}

	result := make(map[string]interface{})

	// Set connection_id as integer (schema expects TypeInt)
	if connectionId, ok := attrs["connection_id"]; ok {
		// Convert float64 to int if needed, otherwise use as-is
		switch v := connectionId.(type) {
		case float64:
			result["connection_id"] = int(v)
		case int:
			result["connection_id"] = v
		default:
			result["connection_id"] = connectionId
		}
	}

	// Set object as string
	if object, ok := attrs["object"]; ok {
		result["object"] = convertToString(object)
	}

	// Set lead_union_insert_to as string if present
	if leadUnionInsertTo, ok := attrs["lead_union_insert_to"]; ok {
		result["lead_union_insert_to"] = convertToString(leadUnionInsertTo)
	}

	return []map[string]interface{}{result}
}
