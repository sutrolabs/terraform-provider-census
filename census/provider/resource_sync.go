package provider

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// syncSchemaMap returns the schema map for the sync resource
// The alertCollectionType parameter allows switching between TypeSet (v0) and TypeList (v1) for alerts
func syncSchemaMap(alertCollectionType schema.ValueType) map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The ID of the sync.",
		},
		"workspace_id": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "The ID of the workspace this sync belongs to.",
		},
		"label": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name/label of the sync.",
		},
		"source_attributes": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "Source-specific configuration (e.g., SQL query, table selection).",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"connection_id": {
						Type:        schema.TypeInt,
						Required:    true,
						Description: "The ID of the source connection.",
					},
					"cohort_id": {
						Type:        schema.TypeInt,
						Optional:    true,
						Description: "The ID of the cohort (for cohort sources). When specified, object.type should be 'cohort', object.id should be the cohort ID, and object.dataset_id should be the dataset ID.",
					},
					"object": {
						Type:        schema.TypeList,
						Required:    true,
						MaxItems:    1,
						Description: "Object configuration for the source.",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"type": {
									Type:        schema.TypeString,
									Required:    true,
									Description: "Type of object (table, dataset, model, etc.).",
								},
								"table_name": {
									Type:        schema.TypeString,
									Optional:    true,
									Description: "Table name (for table type).",
								},
								"table_schema": {
									Type:        schema.TypeString,
									Optional:    true,
									Description: "Table schema (for table type).",
								},
								"table_catalog": {
									Type:        schema.TypeString,
									Optional:    true,
									Description: "Table catalog (for table type).",
								},
								"id": {
									Type:        schema.TypeString,
									Optional:    true,
									Description: "Object ID (for dataset, model, segment, cohort, topic, etc.).",
								},
								"dataset_id": {
									Type:        schema.TypeString,
									Optional:    true,
									Description: "Dataset ID (for segment and cohort sources - the underlying dataset that the segment/cohort belongs to).",
								},
							},
						},
					},
				},
			},
		},
		"destination_attributes": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "Destination-specific configuration (e.g., object, connection_id).",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"connection_id": {
						Type:        schema.TypeInt,
						Required:    true,
						Description: "The ID of the destination connection.",
					},
					"object": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "The destination object name (e.g., 'Contact' for Salesforce).",
					},
					"lead_union_insert_to": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Where to insert a union object (for Salesforce connections).",
					},
				},
			},
		},
		"operation": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "How records are synced to the destination (upsert, append, mirror, etc.).",
			ValidateFunc: validation.StringInSlice([]string{
				"append", "insert", "mirror", "update", "upsert",
			}, false),
		},
		"field_mapping": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Field mappings between source and destination.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"from": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Source field name. Required for column mappings (type='direct'). Omit for constant, sync_metadata, segment_membership, and liquid_template mappings.",
					},
					"to": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Destination field name.",
					},
					"type": {
						Type:        schema.TypeString,
						Optional:    true,
						Default:     "direct",
						Description: "Mapping type: 'direct' (default), 'constant', 'sync_metadata', 'segment_membership', or 'liquid_template'.",
						ValidateFunc: validation.StringInSlice([]string{
							"direct", "constant", "sync_metadata", "segment_membership", "liquid_template",
						}, false),
					},
					"constant": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Constant value. Must also set type='constant'.",
					},
					"sync_metadata_key": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Sync metadata key (e.g., 'sync_run_id'). Must also set type='sync_metadata'.",
					},
					"segment_identify_by": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "How to identify segments (e.g., 'name'). Must also set type='segment_membership'.",
					},
					"liquid_template": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Liquid template for transforming data (e.g., '{{ record[\"field\"] | upcase }}'). Must also set type='liquid_template'.",
					},
					"is_primary_identifier": {
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     false,
						Description: "Whether this field is the primary identifier (sync key) for matching records. Exactly one field_mapping must have this set to true.",
					},
					"lookup_object": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Object to lookup for relationship mapping (e.g., 'user_list'). Used with lookup_field for foreign key lookups.",
					},
					"lookup_field": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Field to use for lookup in the lookup_object (e.g., 'id'). Used with lookup_object for foreign key lookups.",
					},
					"preserve_values": {
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     false,
						Description: "If true, preserves existing values in the destination field and prevents Census from overwriting them.",
					},
					"generate_field": {
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     false,
						Description: "If true, Census will generate/create this field in the destination.",
					},
					"sync_null_values": {
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     true,
						Description: "If true (default), null values in the source will be synced to the destination. Set to false to skip syncing null values.",
					},
					"array_field": {
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     false,
						Description: "Whether the destination field is an array type. Only applicable when generate_field is true (for user-defined fields).",
					},
					"field_type": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "The type of the destination field. Only applicable when generate_field is true (for user-defined fields). Available types depend on the destination.",
					},
					"follow_source_type": {
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     false,
						Description: "Whether the destination field type should automatically follow changes to the source column type.",
					},
				},
			},
		},
		"sync_mode": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "DEPRECATED: This field is ignored. Use 'operation' instead.",
			Deprecated:  "This field is ignored. The 'operation' field is used for sync mode instead.",
		},
		"paused": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Whether the sync is paused.",
		},
		"field_behavior": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			Description: "Specify how fields are synced. Use 'sync_all_properties' to automatically sync all properties from source to destination. " +
				"Use 'specific_properties' (default) for manual field mappings only.",
			ValidateFunc: validation.StringInSlice([]string{
				"sync_all_properties", "specific_properties",
			}, false),
		},
		"field_normalization": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			Description: "If field_behavior is 'sync_all_properties', specify how automatic field names should be normalized. " +
				"Options: 'start_case', 'lower_case', 'upper_case', 'camel_case', 'snake_case', 'match_source_names'.",
			ValidateFunc: validation.StringInSlice([]string{
				"start_case", "lower_case", "upper_case", "camel_case", "snake_case", "match_source_names",
			}, false),
		},
		"field_order": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			Description: "Specify how destination fields should be ordered. Options: 'alphabetical_column_name' (default), 'mapping_order', or 'source_column_order'. " +
				"Only works on destinations that support field ordering.",
			ValidateFunc: validation.StringInSlice([]string{
				"alphabetical_column_name", "mapping_order", "source_column_order",
			}, false),
		},
		"sync_behavior_family": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			Description: "Specifies the behavior family for the sync. Use 'activateEvents' for event-based activation syncs " +
				"(only supported for live syncs from Kafka/streaming sources). Use 'mapRecords' for record mapping syncs " +
				"(not supported for live syncs from Materialize).",
			ValidateFunc: validation.StringInSlice([]string{
				"activateEvents", "mapRecords",
			}, false),
		},
		"advanced_configuration": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "Advanced configuration options specific to the destination type as JSON. Use jsonencode() to specify values. Available options vary by destination.",
			ValidateFunc:     validation.StringIsJSON,
			DiffSuppressFunc: suppressEquivalentJSON,
		},
		"high_water_mark_attribute": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The name of the timestamp column to use for high water mark diffing strategy. When set, append syncs will use this column to identify new records instead of the default Census diff engine (using primary keys). Example: 'updated_at'.",
		},
		"historical_sync_operation": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			Description: "Specifies how the first sync should handle historical records when using append operation. " +
				"Only applicable for append syncs. Options: 'skip_current_records' (skip existing records on first sync) or " +
				"'backfill_all_records' (include all existing records on first sync).",
			ValidateFunc: validation.StringInSlice([]string{
				"skip_current_records", "backfill_all_records",
			}, false),
		},
		"mirror_strategy": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			Description: "Specifies the strategy for mirror syncs. Only applicable when operation is 'mirror'. " +
				"Options: 'sync_updates_and_deletes' (incrementally sync changes - most common), " +
				"'sync_updates_and_nulls' (update records and set nulls without deletes), " +
				"'upload_and_swap' (replace entire destination table with source snapshot).",
			ValidateFunc: validation.StringInSlice([]string{
				"sync_updates_and_deletes",
				"sync_updates_and_nulls",
				"upload_and_swap",
			}, false),
		},
		"alert": {
			Type:        alertCollectionType, // Parameterized: TypeSet for v0, TypeList for v1
			Optional:    true,
			Description: "Alert configurations for the sync. Multiple alerts of different types can be configured.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"id": {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: "The ID of the alert configuration (assigned by Census).",
					},
					"type": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Type of alert configuration.",
						ValidateFunc: validation.StringInSlice([]string{
							"FailureAlertConfiguration",
							"InvalidRecordPercentAlertConfiguration",
							"FullSyncTriggerAlertConfiguration",
							"RecordCountDeviationAlertConfiguration",
							"RuntimeAlertConfiguration",
							"StatusAlertConfiguration",
						}, false),
					},
					"send_for": {
						Type:        schema.TypeString,
						Optional:    true,
						Default:     "first_time",
						Description: "When to send alerts: 'first_time' (default) or 'every_time'.",
						ValidateFunc: validation.StringInSlice([]string{
							"first_time",
							"every_time",
						}, false),
					},
					"should_send_recovery": {
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     true,
						Description: "Whether to send a recovery notification when the alert condition is resolved.",
					},
					"options": {
						Type:        schema.TypeMap,
						Optional:    true,
						Description: "Alert-specific options (e.g., threshold for InvalidRecordPercentAlertConfiguration).",
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
				},
			},
		},
		"run_mode": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "Run mode configuration for the sync (triggered with various trigger types). If not specified, Census will use a default schedule (manually triggered). To avoid drift, explicitly specify the desired run_mode in your configuration.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"type": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Mode type: 'live' for continuous syncing or 'triggered' for event-based syncing.",
						ValidateFunc: validation.StringInSlice([]string{
							"live", "triggered",
						}, false),
					},
					"triggers": {
						Type:        schema.TypeList,
						Optional:    true,
						MaxItems:    1,
						Description: "Trigger configurations (only for 'triggered' mode).",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"schedule": {
									Type:        schema.TypeList,
									Optional:    true,
									MaxItems:    1,
									Description: "Schedule-based trigger configuration.",
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"frequency": {
												Type:        schema.TypeString,
												Required:    true,
												Description: "Sync frequency: never, continuous, quarter_hourly, hourly, daily, weekly, or expression (for cron).",
												ValidateFunc: validation.StringInSlice([]string{
													"never", "continuous", "quarter_hourly", "hourly", "daily", "weekly", "expression",
												}, false),
											},
											"day": {
												Type:        schema.TypeString,
												Optional:    true,
												Description: "Day of week (Sunday-Saturday, for weekly schedules).",
												ValidateFunc: validation.StringInSlice([]string{
													"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday",
												}, false),
											},
											"hour": {
												Type:         schema.TypeInt,
												Optional:     true,
												Description:  "Hour to run (0-24).",
												ValidateFunc: validation.IntBetween(0, 24),
											},
											"minute": {
												Type:         schema.TypeInt,
												Optional:     true,
												Description:  "Minute to run (0-59).",
												ValidateFunc: validation.IntBetween(0, 59),
											},
											"cron_expression": {
												Type:        schema.TypeString,
												Optional:    true,
												Description: "Cron expression (only valid when frequency is 'expression').",
											},
										},
									},
								},
								"dbt_cloud": {
									Type:        schema.TypeList,
									Optional:    true,
									MaxItems:    1,
									Description: "dbt Cloud job trigger configuration.",
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"project_id": {
												Type:        schema.TypeString,
												Required:    true,
												Description: "dbt Cloud project ID.",
											},
											"job_id": {
												Type:        schema.TypeString,
												Required:    true,
												Description: "dbt Cloud job ID.",
											},
										},
									},
								},
								"fivetran": {
									Type:        schema.TypeList,
									Optional:    true,
									MaxItems:    1,
									Description: "Fivetran connector trigger configuration.",
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"job_id": {
												Type:        schema.TypeString,
												Required:    true,
												Description: "Fivetran job ID.",
											},
											"job_name": {
												Type:        schema.TypeString,
												Required:    true,
												Description: "Fivetran job name.",
											},
										},
									},
								},
								"sync_sequence": {
									Type:        schema.TypeList,
									Optional:    true,
									MaxItems:    1,
									Description: "Sync dependency trigger configuration (triggers after another sync completes).",
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"sync_id": {
												Type:        schema.TypeInt,
												Required:    true,
												Description: "ID of the sync to trigger after.",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"status": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Current status of the sync.",
		},
		"created_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Timestamp when the sync was created.",
		},
		"updated_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Timestamp when the sync was last updated.",
		},
		"last_run_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Timestamp when the sync was last executed.",
			Deprecated:  "This field is unsupported and will be removed in a future version.",
		},
		"next_run_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Timestamp when the sync is scheduled to run next.",
			Deprecated:  "This field is unsupported and will be removed in a future version.",
		},
		"last_run_id": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "ID of the last sync run.",
			Deprecated:  "This field is unsupported and will be removed in a future version.",
		},
	}
}

// ResourceSync returns the resource schema for a Census sync
func ResourceSync() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Census data sync between a source and destination.",

		CreateContext: resourceSyncCreate,
		ReadContext:   resourceSyncRead,
		UpdateContext: resourceSyncUpdate,
		DeleteContext: resourceSyncDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceSyncImport,
		},

		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    ResourceSyncV0().CoreConfigSchema().ImpliedType(),
				Upgrade: ResourceSyncStateUpgradeV0,
				Version: 0,
			},
		},

		Schema: syncSchemaMap(schema.TypeList),
	}
}

// ResourceSyncV0 returns the v0 schema (with TypeSet for alerts)
func ResourceSyncV0() *schema.Resource {
	return &schema.Resource{
		Schema: syncSchemaMap(schema.TypeSet),
	}
}

// ResourceSyncStateUpgradeV0 upgrades the state from v0 to v1
// Migrates alert field from TypeSet to TypeList
func ResourceSyncStateUpgradeV0(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	// TypeSet and TypeList store data identically in JSON state (as arrays)
	// No data transformation needed - just return state unchanged
	// Terraform will re-interpret it with the new schema
	return rawState, nil
}

// suppressEquivalentJSON suppresses diffs for JSON strings that are semantically equivalent
func suppressEquivalentJSON(k, old, new string, d *schema.ResourceData) bool {
	if old == "" && new == "" {
		return true
	}
	if old == "" || new == "" {
		return false
	}

	var oldJSON, newJSON interface{}
	if err := json.Unmarshal([]byte(old), &oldJSON); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(new), &newJSON); err != nil {
		return false
	}

	return reflect.DeepEqual(oldJSON, newJSON)
}
