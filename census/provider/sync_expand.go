package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/sutrolabs/terraform-provider-census/census/client/v2"
)

// Helper functions for expanding complex types (Terraform config/state -> API request)

// ExpandFieldMappings converts the Terraform field_mapping blocks directly
// into the v2 API's MappingAttributes wire shape. The two are bridged here,
// in one step, rather than through an intermediate Terraform-shaped struct —
// v2 only speaks one mapping shape, unlike v1 (which had a legacy
// FieldMappings shape alongside Mappings), so there's nothing left to bridge
// between.
func ExpandFieldMappings(ctx context.Context, mappings []interface{}) []v2.MappingAttributes {
	result := make([]v2.MappingAttributes, 0, len(mappings))
	for i, mapping := range mappings {
		m, ok := mapping.(map[string]interface{})
		if !ok {
			tflog.Debug(ctx, "ExpandFieldMappings: mapping is not a map", map[string]interface{}{"index": i, "type": fmt.Sprintf("%T", mapping)})
			continue
		}

		from, _ := m["from"].(string)
		to, _ := m["to"].(string)
		constant := m["constant"]
		syncMetadataKey, _ := m["sync_metadata_key"].(string)
		segmentIdentifyBy, _ := m["segment_identify_by"].(string)
		liquidTemplate, _ := m["liquid_template"].(string)

		mappingType := "direct"
		if typeVal, ok := m["type"].(string); ok && typeVal != "" {
			mappingType = typeVal
		}

		var mappingFrom v2.MappingFrom
		switch mappingType {
		case "constant":
			mappingFrom = v2.MappingFrom{
				Type: "constant_value",
				Data: map[string]interface{}{
					"basic_type": "text",
					"value":      fmt.Sprintf("%v", constant),
				},
			}
		case "sync_metadata":
			mappingFrom = v2.MappingFrom{Type: "sync_metadata", Data: syncMetadataKey}
		case "segment_membership":
			mappingFrom = v2.MappingFrom{
				Type: "segment_membership",
				Data: map[string]interface{}{"identify_by": segmentIdentifyBy},
			}
		case "liquid_template":
			mappingFrom = v2.MappingFrom{Type: "liquid_template", Data: liquidTemplate}
		default:
			mappingFrom = v2.MappingFrom{Type: "column", Data: from}
		}

		attr := v2.MappingAttributes{
			From: mappingFrom,
			To:   to,
		}

		if isPrimary, ok := m["is_primary_identifier"].(bool); ok {
			attr.IsPrimaryIdentifier = isPrimary
		}
		if operation, ok := m["operation"].(string); ok {
			attr.Operation = operation
		}
		if lookupObject, ok := m["lookup_object"].(string); ok {
			attr.LookupObject = lookupObject
		}
		if lookupField, ok := m["lookup_field"].(string); ok {
			attr.LookupField = lookupField
		}
		if preserveValues, ok := m["preserve_values"].(bool); ok {
			attr.PreserveValues = preserveValues
		}
		if generateField, ok := m["generate_field"].(bool); ok {
			attr.GenerateField = generateField
		}
		if syncNullValues, ok := m["sync_null_values"].(bool); ok {
			attr.SyncNullValues = &syncNullValues
		}
		if arrayField, ok := m["array_field"].(bool); ok {
			attr.ArrayField = arrayField
		}
		if fieldType, ok := m["field_type"].(string); ok {
			attr.FieldType = fieldType
		}
		if followSourceType, ok := m["follow_source_type"].(bool); ok {
			attr.FollowSourceType = followSourceType
		}

		result = append(result, attr)
	}
	return result
}

func ExpandAlerts(ctx context.Context, alerts []interface{}) []v2.AlertAttribute {
	// Return empty slice instead of nil to allow deleting all alerts, and so
	// that the (required, not omitempty) Options field below is never left
	// nil on an element — a nil map marshals to JSON null, not {}, and v2
	// requires options on every alert even when there are none to set.
	result := make([]v2.AlertAttribute, 0, len(alerts))

	for i, alert := range alerts {
		m, ok := alert.(map[string]interface{})
		if !ok {
			tflog.Debug(ctx, "ExpandAlerts: alert is not a map", map[string]interface{}{"index": i, "type": fmt.Sprintf("%T", alert)})
			continue
		}

		alertAttr := v2.AlertAttribute{
			Options: make(map[string]interface{}),
		}

		if id, ok := m["id"].(int); ok {
			alertAttr.ID = id
		}

		if alertType, ok := m["type"].(string); ok && alertType != "" {
			alertAttr.Type = alertType
		}

		if sendFor, ok := m["send_for"].(string); ok && sendFor != "" {
			alertAttr.SendFor = sendFor
		} else {
			alertAttr.SendFor = "first_time"
		}

		if shouldSendRecovery, ok := m["should_send_recovery"].(bool); ok {
			alertAttr.ShouldSendRecovery = shouldSendRecovery
		} else {
			alertAttr.ShouldSendRecovery = true
		}

		if options, ok := m["options"].(map[string]interface{}); ok {
			for key, value := range options {
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

		// Only append if we have a valid alert type (required field) — this
		// prevents sending invalid empty alerts to the API.
		if alertAttr.Type != "" {
			result = append(result, alertAttr)
		}
	}
	return result
}

// ExpandRunMode converts Terraform run_mode config to the API's SyncMode
// struct. Only "triggered" mode is schema-valid now — "live" has been fully
// sunset by Fivetran — but this doesn't special-case the type value at all;
// narrowing that is the schema's job (see run_mode.type's ValidateFunc).
func ExpandRunMode(ctx context.Context, runModes []interface{}) *v2.SyncMode {
	if len(runModes) == 0 || runModes[0] == nil {
		return nil
	}

	runModeMap, ok := runModes[0].(map[string]interface{})
	if !ok {
		tflog.Debug(ctx, "ExpandRunMode: runModes[0] is not a map", map[string]interface{}{"type": fmt.Sprintf("%T", runModes[0])})
		return nil
	}

	mode := &v2.SyncMode{}

	if modeType, ok := runModeMap["type"].(string); ok {
		mode.Type = modeType
	}

	if triggersList, ok := runModeMap["triggers"].([]interface{}); ok && len(triggersList) > 0 {
		if triggersMap, ok := triggersList[0].(map[string]interface{}); ok {
			triggers := &v2.SyncTriggers{}

			if scheduleList, ok := triggersMap["schedule"].([]interface{}); ok && len(scheduleList) > 0 {
				if scheduleMap, ok := scheduleList[0].(map[string]interface{}); ok {
					schedule := &v2.TriggerSchedule{}

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

			if dbtList, ok := triggersMap["dbt_cloud"].([]interface{}); ok && len(dbtList) > 0 {
				if dbtMap, ok := dbtList[0].(map[string]interface{}); ok {
					dbt := &v2.DbtCloudTrigger{}

					if projectID, ok := dbtMap["project_id"].(string); ok {
						dbt.ProjectID = projectID
					}
					if jobID, ok := dbtMap["job_id"].(string); ok {
						dbt.JobID = jobID
					}

					triggers.DbtCloud = dbt
				}
			}

			if fivetranList, ok := triggersMap["fivetran"].([]interface{}); ok && len(fivetranList) > 0 {
				if fivetranMap, ok := fivetranList[0].(map[string]interface{}); ok {
					fivetran := &v2.FivetranTrigger{}

					if jobID, ok := fivetranMap["job_id"].(string); ok {
						fivetran.JobID = jobID
					}
					if jobName, ok := fivetranMap["job_name"].(string); ok {
						fivetran.JobName = jobName
					}

					triggers.Fivetran = fivetran
				}
			}

			if seqList, ok := triggersMap["sync_sequence"].([]interface{}); ok && len(seqList) > 0 {
				if seqMap, ok := seqList[0].(map[string]interface{}); ok {
					seq := &v2.SyncSequenceTrigger{}

					if syncID, ok := seqMap["sync_id"].(int); ok {
						seq.SyncID = syncID
					}

					triggers.SyncSequence = seq
				}
			}

			mode.Triggers = triggers
		}
	}

	return mode
}

// CleanEmptyStrings removes empty string and zero values from a map
// to avoid sending invalid data to the Activations API
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

// ExpandSourceAttributes converts list-based source_attributes from
// Terraform to the map format the API expects. The source is now determined
// by the sync's activation_id, not by a connection_id — that field has been
// removed from the schema entirely.
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

// ExpandDestinationAttributes converts the destination_attributes block to
// the API's typed shape. The destination is now determined by the sync's
// activation_id, not by a connection_id — that field has been removed from
// the schema entirely.
func ExpandDestinationAttributes(ctx context.Context, destAttrs []interface{}) *v2.SyncDestinationAttributes {
	if len(destAttrs) == 0 {
		return nil
	}

	attrInterface := destAttrs[0]
	attr, ok := attrInterface.(map[string]interface{})
	if !ok {
		tflog.Debug(ctx, "ExpandDestinationAttributes: destAttrs[0] is not a map", map[string]interface{}{"type": fmt.Sprintf("%T", attrInterface)})
		return nil
	}

	result := &v2.SyncDestinationAttributes{}
	if v, ok := attr["object"].(string); ok {
		result.Object = v
	}
	if v, ok := attr["lead_union_insert_to"].(string); ok {
		result.LeadUnionInsertTo = v
	}

	return result
}
