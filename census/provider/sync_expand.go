package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/sutrolabs/terraform-provider-census/census/client"
)

// Helper functions for expanding complex types (Terraform config/state -> API request)

func ExpandFieldMappings(ctx context.Context, mappings []interface{}) []client.FieldMapping {
	result := make([]client.FieldMapping, 0, len(mappings))
	for i, mapping := range mappings {
		// Safe type assertion for mapping
		m, ok := mapping.(map[string]interface{})
		if !ok {
			tflog.Debug(ctx, "ExpandFieldMappings: mapping is not a map", map[string]interface{}{"index": i, "type": fmt.Sprintf("%T", mapping)})
			continue // Skip invalid entries
		}

		fieldMapping := client.FieldMapping{
			Constant: m["constant"], // This is safe as interface{}
		}

		// Safe type assertions for string fields
		if from, ok := m["from"].(string); ok {
			fieldMapping.From = from
		} else {
			tflog.Debug(ctx, "ExpandFieldMappings: 'from' is not a string", map[string]interface{}{"index": i, "type": fmt.Sprintf("%T", m["from"])})
		}

		if to, ok := m["to"].(string); ok {
			fieldMapping.To = to
		} else {
			tflog.Debug(ctx, "ExpandFieldMappings: 'to' is not a string", map[string]interface{}{"index": i, "type": fmt.Sprintf("%T", m["to"])})
		}

		// Get type field (defaults to "direct" in schema)
		mappingType := "direct" // fallback default
		if typeVal, ok := m["type"].(string); ok && typeVal != "" {
			mappingType = typeVal
		}
		fieldMapping.Type = mappingType

		// Handle type-specific fields
		if syncMetadataKey, ok := m["sync_metadata_key"].(string); ok {
			fieldMapping.SyncMetadataKey = syncMetadataKey
		}

		if segmentIdentifyBy, ok := m["segment_identify_by"].(string); ok {
			fieldMapping.SegmentIdentifyBy = segmentIdentifyBy
		}

		if liquidTemplate, ok := m["liquid_template"].(string); ok {
			fieldMapping.LiquidTemplate = liquidTemplate
		}

		// Validate: if constant is present, type must be "constant"
		if fieldMapping.Constant != nil && fieldMapping.Constant != "" {
			if mappingType != "constant" {
				tflog.Warn(ctx, "ExpandFieldMappings: field_mapping has a constant value but type is not 'constant'", map[string]interface{}{"index": i, "type": mappingType})
				// Continue processing but log the error - validation should catch this
			}
		}

		if isPrimary, ok := m["is_primary_identifier"].(bool); ok {
			fieldMapping.IsPrimaryIdentifier = isPrimary
		}

		if lookupObject, ok := m["lookup_object"].(string); ok {
			fieldMapping.LookupObject = lookupObject
		}

		if lookupField, ok := m["lookup_field"].(string); ok {
			fieldMapping.LookupField = lookupField
		}

		if preserveValues, ok := m["preserve_values"].(bool); ok {
			fieldMapping.PreserveValues = preserveValues
		}

		if generateField, ok := m["generate_field"].(bool); ok {
			fieldMapping.GenerateField = generateField
		}

		if syncNullValues, ok := m["sync_null_values"]; ok {
			if val, isBool := syncNullValues.(bool); isBool {
				fieldMapping.SyncNullValues = &val
			}
		}

		if arrayField, ok := m["array_field"].(bool); ok {
			fieldMapping.ArrayField = arrayField
		}

		if fieldType, ok := m["field_type"].(string); ok {
			fieldMapping.FieldType = fieldType
		}

		if followSourceType, ok := m["follow_source_type"].(bool); ok {
			fieldMapping.FollowSourceType = followSourceType
		}

		result = append(result, fieldMapping)
	}
	return result
}

func ExpandAlerts(ctx context.Context, alerts []interface{}) []client.AlertAttribute {
	// Return empty slice instead of nil to allow deleting all alerts
	// When nil is used, the omitempty tag causes the field to be omitted entirely
	// from the JSON, which the API interprets as "don't change"
	result := make([]client.AlertAttribute, 0, len(alerts))

	if len(alerts) == 0 {
		return result
	}
	for i, alert := range alerts {
		m, ok := alert.(map[string]interface{})
		if !ok {
			tflog.Debug(ctx, "ExpandAlerts: alert is not a map", map[string]interface{}{"index": i, "type": fmt.Sprintf("%T", alert)})
			continue
		}

		alertAttr := client.AlertAttribute{
			Options: make(map[string]interface{}),
		}

		// Preserve alert ID if it exists (required for updates)
		if id, ok := m["id"].(int); ok {
			alertAttr.ID = id
		}

		if alertType, ok := m["type"].(string); ok && alertType != "" {
			alertAttr.Type = alertType
		}

		if sendFor, ok := m["send_for"].(string); ok && sendFor != "" {
			alertAttr.SendFor = sendFor
		} else {
			alertAttr.SendFor = "first_time" // default
		}

		if shouldSendRecovery, ok := m["should_send_recovery"].(bool); ok {
			alertAttr.ShouldSendRecovery = shouldSendRecovery
		} else {
			alertAttr.ShouldSendRecovery = true // default
		}

		// Handle options - convert string values to appropriate types
		if options, ok := m["options"].(map[string]interface{}); ok {
			for key, value := range options {
				// Try to convert string values to integers for threshold fields
				if strVal, ok := value.(string); ok {
					if key == "threshold" {
						if intVal, err := strconv.Atoi(strVal); err == nil {
							alertAttr.Options[key] = intVal
							continue
						}
					}
				}
				alertAttr.Options[key] = value
			}
		}

		// Only append if we have a valid alert type (required field)
		// This prevents sending invalid empty alerts to the API
		if alertAttr.Type != "" {
			result = append(result, alertAttr)
		}
	}
	return result
}

// ExpandRunMode converts Terraform run_mode config to API SyncMode struct
func ExpandRunMode(ctx context.Context, runModes []interface{}) *client.SyncMode {
	if len(runModes) == 0 || runModes[0] == nil {
		return nil
	}

	runModeMap, ok := runModes[0].(map[string]interface{})
	if !ok {
		tflog.Debug(ctx, "ExpandRunMode: runModes[0] is not a map", map[string]interface{}{"type": fmt.Sprintf("%T", runModes[0])})
		return nil
	}

	mode := &client.SyncMode{}

	// Extract type (required)
	if modeType, ok := runModeMap["type"].(string); ok {
		mode.Type = modeType
	}

	// Extract triggers (optional, only for triggered mode)
	if triggersList, ok := runModeMap["triggers"].([]interface{}); ok && len(triggersList) > 0 {
		if triggersMap, ok := triggersList[0].(map[string]interface{}); ok {
			triggers := &client.SyncTriggers{}

			// Extract schedule trigger
			if scheduleList, ok := triggersMap["schedule"].([]interface{}); ok && len(scheduleList) > 0 {
				if scheduleMap, ok := scheduleList[0].(map[string]interface{}); ok {
					schedule := &client.TriggerSchedule{}

					if freq, ok := scheduleMap["frequency"].(string); ok {
						schedule.Frequency = freq
					}
					if day, ok := scheduleMap["day"].(string); ok {
						schedule.Day = day
					}
					if hour, ok := scheduleMap["hour"].(int); ok {
						schedule.Hour = hour
					}
					if minute, ok := scheduleMap["minute"].(int); ok {
						schedule.Minute = minute
					}
					if cron, ok := scheduleMap["cron_expression"].(string); ok {
						schedule.CronExpression = cron
					}

					triggers.Schedule = schedule
				}
			}

			// Extract dbt_cloud trigger
			if dbtList, ok := triggersMap["dbt_cloud"].([]interface{}); ok && len(dbtList) > 0 {
				if dbtMap, ok := dbtList[0].(map[string]interface{}); ok {
					dbt := &client.DbtCloudTrigger{}

					if projectId, ok := dbtMap["project_id"].(string); ok {
						dbt.ProjectId = projectId
					}
					if jobId, ok := dbtMap["job_id"].(string); ok {
						dbt.JobId = jobId
					}

					triggers.DbtCloud = dbt
				}
			}

			// Extract fivetran trigger
			if fivetranList, ok := triggersMap["fivetran"].([]interface{}); ok && len(fivetranList) > 0 {
				if fivetranMap, ok := fivetranList[0].(map[string]interface{}); ok {
					fivetran := &client.FivetranTrigger{}

					if jobId, ok := fivetranMap["job_id"].(string); ok {
						fivetran.JobId = jobId
					}
					if jobName, ok := fivetranMap["job_name"].(string); ok {
						fivetran.JobName = jobName
					}

					triggers.Fivetran = fivetran
				}
			}

			// Extract sync_sequence trigger
			if seqList, ok := triggersMap["sync_sequence"].([]interface{}); ok && len(seqList) > 0 {
				if seqMap, ok := seqList[0].(map[string]interface{}); ok {
					seq := &client.SyncSequenceTrigger{}

					if syncId, ok := seqMap["sync_id"].(int); ok {
						seq.SyncId = syncId
					}

					triggers.SyncSequence = seq
				}
			}

			mode.Triggers = triggers
		}
	}

	return mode
}

func ExpandStringMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return nil
	}
	result := make(map[string]interface{})
	for k, v := range m {
		result[k] = v
	}
	return result
}

// CleanEmptyStrings removes empty string and zero values from a map
// to avoid sending invalid data to the Census API
func CleanEmptyStrings(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return nil
	}
	result := make(map[string]interface{})
	for k, v := range m {
		// Skip empty strings
		if str, ok := v.(string); ok && str == "" {
			continue
		}
		// Skip zero integers for cohort_id (default value from Terraform)
		if num, ok := v.(int); ok && num == 0 && k == "cohort_id" {
			continue
		}
		result[k] = v
	}
	return result
}

func ExpandAdvancedConfiguration(ctx context.Context, jsonStr string) map[string]interface{} {
	if jsonStr == "" {
		return nil
	}
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		tflog.Debug(ctx, "Failed to unmarshal advanced_configuration", map[string]interface{}{"error": err.Error()})
		return nil
	}
	return result
}

// ExpandSourceAttributes converts list-based source_attributes from Terraform to map format for API
func ExpandSourceAttributes(ctx context.Context, sourceAttrs []interface{}) map[string]interface{} {
	if len(sourceAttrs) == 0 {
		return nil
	}

	// Safe type assertion for source attributes
	attrInterface := sourceAttrs[0]
	attr, ok := attrInterface.(map[string]interface{})
	if !ok {
		tflog.Debug(ctx, "ExpandSourceAttributes: sourceAttrs[0] is not a map", map[string]interface{}{"type": fmt.Sprintf("%T", attrInterface)})
		return nil
	}

	result := make(map[string]interface{})

	// Copy basic attributes
	if v, ok := attr["connection_id"]; ok && v != "" {
		result["connection_id"] = v
	}

	// Copy top-level cohort_id if present (for cohort sources)
	if v, ok := attr["cohort_id"]; ok && v != "" {
		result["cohort_id"] = v
	}

	// Handle nested object - it can be either a list of maps (from Terraform state) or a direct map
	var objectMap map[string]interface{}
	if objData, ok := attr["object"]; ok {
		switch v := objData.(type) {
		case []interface{}:
			// Object stored as list in Terraform state
			if len(v) > 0 {
				if obj, ok := v[0].(map[string]interface{}); ok {
					objectMap = obj
				} else {
					tflog.Debug(ctx, "ExpandSourceAttributes: object list element is not a map", map[string]interface{}{"type": fmt.Sprintf("%T", v[0])})
					return result // Return partial result instead of nil
				}
			}
		case map[string]interface{}:
			// Object is directly a map (direct config)
			objectMap = v
		default:
			tflog.Debug(ctx, "ExpandSourceAttributes: object is an unexpected type", map[string]interface{}{"type": fmt.Sprintf("%T", objData)})
		}
	}

	// Smart translation for segment and cohort sources
	if objectMap != nil {
		objectType := ""
		if t, ok := objectMap["type"].(string); ok {
			objectType = t
		}

		switch objectType {
		case "segment":
			// User provides: type="segment", id=<segment_id>, dataset_id=<dataset_id>
			// API needs: object.type="dataset", object.id=<dataset_id>, filter_segment_id=<segment_id>
			translatedObject := make(map[string]interface{})
			translatedObject["type"] = "dataset"
			if datasetId, ok := objectMap["dataset_id"]; ok && datasetId != "" {
				translatedObject["id"] = datasetId
			}
			result["object"] = translatedObject
			if segmentId, ok := objectMap["id"]; ok && segmentId != "" {
				result["filter_segment_id"] = segmentId
			}
			tflog.Debug(ctx, "ExpandSourceAttributes: translated segment source", map[string]interface{}{"object": fmt.Sprintf("%+v", translatedObject), "filter_segment_id": result["filter_segment_id"]})

		case "cohort":
			// User provides: type="cohort", id=<cohort_id>, dataset_id=<dataset_id>
			// API needs: object.type="dataset", object.id=<dataset_id>, cohort_id=<cohort_id>
			translatedObject := make(map[string]interface{})
			translatedObject["type"] = "dataset"
			if datasetId, ok := objectMap["dataset_id"]; ok && datasetId != "" {
				translatedObject["id"] = datasetId
			}
			result["object"] = translatedObject
			if cohortId, ok := objectMap["id"]; ok && cohortId != "" {
				result["cohort_id"] = cohortId
			}
			tflog.Debug(ctx, "ExpandSourceAttributes: translated cohort source", map[string]interface{}{"object": fmt.Sprintf("%+v", translatedObject), "cohort_id": result["cohort_id"]})

		default:
			// For all other types (model, topic, dataset, table), pass through as-is
			// But clean empty strings to avoid API errors
			result["object"] = CleanEmptyStrings(objectMap)
			tflog.Debug(ctx, "ExpandSourceAttributes: pass-through object (cleaned)", map[string]interface{}{"object_type": objectType, "object": fmt.Sprintf("%+v", result["object"])})
		}
	}

	// Remove cohort_id if it's 0 (default value from Terraform, not actually set)
	if cohortId, ok := result["cohort_id"].(int); ok && cohortId == 0 {
		delete(result, "cohort_id")
	}

	return result
}

// ExpandDestinationAttributes converts list-based destination_attributes from Terraform to map format for API
func ExpandDestinationAttributes(ctx context.Context, destAttrs []interface{}) map[string]interface{} {
	if len(destAttrs) == 0 {
		return nil
	}

	// Safe type assertion for destination attributes
	attrInterface := destAttrs[0]
	attr, ok := attrInterface.(map[string]interface{})
	if !ok {
		tflog.Debug(ctx, "ExpandDestinationAttributes: destAttrs[0] is not a map", map[string]interface{}{"type": fmt.Sprintf("%T", attrInterface)})
		return nil
	}

	result := make(map[string]interface{})

	// Copy attributes
	if v, ok := attr["connection_id"]; ok && v != "" {
		result["connection_id"] = v
	}
	if v, ok := attr["object"]; ok && v != "" {
		result["object"] = v
	}
	if v, ok := attr["lead_union_insert_to"]; ok && v != "" {
		result["lead_union_insert_to"] = v
	}

	return result
}
