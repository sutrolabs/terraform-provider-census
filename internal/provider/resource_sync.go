package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"reflect"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/sutrolabs/terraform-provider-census/internal/client"
)

func resourceSync() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Census data sync between a source and destination.",

		CreateContext: resourceSyncCreate,
		ReadContext:   resourceSyncRead,
		UpdateContext: resourceSyncUpdate,
		DeleteContext: resourceSyncDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceSyncImport,
		},

		Schema: map[string]*schema.Schema{
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
							Description: "Mapping type: 'direct' (default), 'hash', 'constant', 'sync_metadata', 'segment_membership', or 'liquid_template'.",
							ValidateFunc: validation.StringInSlice([]string{
								"direct", "hash", "constant", "sync_metadata", "segment_membership", "liquid_template",
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
				Description: "Specify how destination fields should be ordered. Options: 'alphabetical_column_name' (default) or 'mapping_order'. " +
					"Only works on destinations that support field ordering.",
				ValidateFunc: validation.StringInSlice([]string{
					"alphabetical_column_name", "mapping_order",
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
				Type:        schema.TypeSet,
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
				Set: alertHash,
			},
			"run_mode": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Run mode configuration for the sync (live vs triggered with various trigger types).",
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
			// Computed fields
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
			},
			"next_run_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp when the sync is scheduled to run next.",
			},
			"last_run_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "ID of the last sync run.",
			},
		},
	}
}

// alertHash creates a hash for an alert to use in a TypeSet
func alertHash(v interface{}) int {
	m := v.(map[string]interface{})

	// Create a unique string representation based on type and key options
	alertType := ""
	if val, ok := m["type"].(string); ok {
		alertType = val
	}

	sendFor := "first_time" // default
	if val, ok := m["send_for"].(string); ok && val != "" {
		sendFor = val
	}

	shouldSendRecovery := "true" // default
	if val, ok := m["should_send_recovery"].(bool); ok {
		shouldSendRecovery = fmt.Sprintf("%t", val)
	}

	// Include options in hash for uniqueness
	hashStr := fmt.Sprintf("%s:%s:%s", alertType, sendFor, shouldSendRecovery)

	// Add options to hash if present
	if options, ok := m["options"].(map[string]interface{}); ok && len(options) > 0 {
		hashStr = fmt.Sprintf("%s:%v", hashStr, options)
	}

	h := fnv.New32a()
	h.Write([]byte(hashStr))
	return int(h.Sum32())
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

func resourceSyncCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*client.Client)

	workspaceId := d.Get("workspace_id").(string)

	// Get the workspace API key dynamically using the personal access token
	workspaceIdInt, err := strconv.Atoi(workspaceId)
	if err != nil {
		return diag.Errorf("invalid workspace ID: %s", workspaceId)
	}

	workspaceToken, err := apiClient.GetWorkspaceAPIKey(ctx, workspaceIdInt)
	if err != nil {
		return diag.Errorf("failed to get workspace API key for workspace %d: %v", workspaceIdInt, err)
	}
	if workspaceToken == "" {
		return diag.Errorf("workspace API key is empty for workspace %d", workspaceIdInt)
	}

	destinationAttributes := expandDestinationAttributes(d.Get("destination_attributes").([]interface{}))
	fieldMappings := expandFieldMappings(d.Get("field_mapping").([]interface{}))

	// Validate exactly one primary identifier
	if err := validatePrimaryIdentifier(fieldMappings); err != nil {
		return diag.FromErr(err)
	}

	// Get operation from top-level field (per OpenAPI spec)
	operation := d.Get("operation").(string)

	// Convert FieldMappings to MappingAttributes for API compliance
	mappings := convertFieldMappingsToMappingAttributes(fieldMappings)

	// Handle run_mode
	var mode *client.SyncMode
	runModeRaw := d.Get("run_mode").([]interface{})

	if len(runModeRaw) > 0 {
		mode = expandRunMode(runModeRaw)
	}

	req := &client.CreateSyncRequest{
		// Required fields per OpenAPI spec
		Operation:             operation,
		SourceAttributes:      expandSourceAttributes(d.Get("source_attributes").([]interface{})),
		DestinationAttributes: destinationAttributes,
		Mappings:              mappings,

		// Optional fields
		Label: d.Get("label").(string),

		// Mode - live vs triggered with trigger configurations
		Mode: mode,

		Paused: d.Get("paused").(bool),

		// Field configuration
		FieldBehavior:      d.Get("field_behavior").(string),
		FieldNormalization: d.Get("field_normalization").(string),
		FieldOrder:         d.Get("field_order").(string),

		// Sync behavior family
		SyncBehaviorFamily: d.Get("sync_behavior_family").(string),

		// Advanced configuration
		AdvancedConfiguration: expandAdvancedConfiguration(d.Get("advanced_configuration").(string)),

		// High water mark attribute
		HighWaterMarkAttribute: d.Get("high_water_mark_attribute").(string),

		// Historical sync operation
		HistoricalSyncOperation: d.Get("historical_sync_operation").(string),

		// Mirror strategy
		MirrorStrategy: d.Get("mirror_strategy").(string),

		// Alert configuration
		AlertAttributes: expandAlerts(d.Get("alert").(*schema.Set).List()),
	}

	fmt.Printf("[DEBUG] Creating sync with request: %+v\n", req)
	sync, err := apiClient.CreateSyncWithToken(ctx, req, workspaceToken)
	if err != nil {
		return diag.FromErr(err)
	}

	fmt.Printf("[DEBUG] Sync created successfully with ID: %d\n", sync.ID)
	d.SetId(strconv.Itoa(sync.ID))
	fmt.Printf("[DEBUG] Resource ID set to: %s\n", d.Id())

	// Explicitly set workspace_id from our input since API doesn't return it
	d.Set("workspace_id", workspaceId)

	return resourceSyncRead(ctx, d, meta)
}

func resourceSyncRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*client.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("invalid sync ID: %s", d.Id())
	}

	// DEBUG: Log entry point
	fmt.Printf("[DEBUG] Starting resourceSyncRead for sync ID: %d\n", id)

	// Get workspace token dynamically if we have workspace_id
	workspaceId := d.Get("workspace_id").(string)
	fmt.Printf("[DEBUG] Got workspace_id from state: %s\n", workspaceId)

	var sync *client.Sync
	if workspaceId != "" {
		workspaceIdInt, err := strconv.Atoi(workspaceId)
		if err != nil {
			return diag.Errorf("invalid workspace ID: %s", workspaceId)
		}

		fmt.Printf("[DEBUG] Getting workspace token for workspace %d\n", workspaceIdInt)
		workspaceToken, err := apiClient.GetWorkspaceAPIKey(ctx, workspaceIdInt)
		if err != nil {
			fmt.Printf("[DEBUG] Failed to get workspace API key: %v\n", err)
			return diag.FromErr(err)
		}

		fmt.Printf("[DEBUG] Successfully got workspace token, calling GetSyncWithToken for sync %d\n", id)
		sync, err = apiClient.GetSyncWithToken(ctx, id, workspaceToken)
		fmt.Printf("[DEBUG] GetSyncWithToken completed - sync: %v, err: %v\n", sync != nil, err)
	} else {
		return diag.Errorf(`workspace_id is required but missing from resource state.

To fix this, add the missing workspace_id to terraform state:
  terraform state rm census_sync.example
  terraform import census_sync.example workspace_id:sync_id`)
	}

	if err != nil {
		fmt.Printf("[DEBUG] Error from GetSyncWithToken: %v\n", err)
		// Check if sync was not found
		if IsNotFoundError(err) {
			fmt.Printf("[DEBUG] Sync not found, clearing resource ID\n")
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	// Check if sync is nil (API returned successfully but with nil data)
	if sync == nil {
		fmt.Printf("[DEBUG] Sync is nil, clearing resource ID\n")
		d.SetId("")
		return nil
	}

	fmt.Printf("[DEBUG] Sync found successfully, setting resource attributes\n")

	// Only update workspace_id if API returned it, otherwise preserve what's in state
	if sync.WorkspaceID != "" {
		d.Set("workspace_id", sync.WorkspaceID)
	}

	d.Set("label", sync.Label)
	d.Set("status", sync.Status)
	d.Set("paused", sync.Paused)

	// Set operation field from API response
	if sync.Operation != "" {
		d.Set("operation", sync.Operation)
	}

	// Set field configuration only if explicitly configured by user
	// This prevents drift when API returns defaults but user hasn't specified them
	if _, ok := d.GetOk("field_behavior"); ok && sync.FieldBehavior != "" {
		d.Set("field_behavior", sync.FieldBehavior)
	}
	if _, ok := d.GetOk("field_normalization"); ok && sync.FieldNormalization != "" {
		d.Set("field_normalization", sync.FieldNormalization)
	}
	if _, ok := d.GetOk("field_order"); ok && sync.FieldOrder != "" {
		d.Set("field_order", sync.FieldOrder)
	}
	if _, ok := d.GetOk("sync_behavior_family"); ok && sync.SyncBehaviorFamily != "" {
		d.Set("sync_behavior_family", sync.SyncBehaviorFamily)
	}

	// Set advanced configuration if present
	if sync.AdvancedConfiguration != nil && len(sync.AdvancedConfiguration) > 0 {
		if err := d.Set("advanced_configuration", flattenAdvancedConfiguration(sync.AdvancedConfiguration)); err != nil {
			fmt.Printf("[DEBUG] Failed to set advanced_configuration: %v\n", err)
			return diag.Errorf("failed to set advanced_configuration: %v", err)
		}
	}

	// Set high water mark attribute if present
	if sync.HighWaterMarkAttribute != "" {
		d.Set("high_water_mark_attribute", sync.HighWaterMarkAttribute)
	}

	// Set historical sync operation if present
	if sync.HistoricalSyncOperation != "" {
		d.Set("historical_sync_operation", sync.HistoricalSyncOperation)
	}

	// Set mirror strategy if present
	if sync.MirrorStrategy != "" {
		d.Set("mirror_strategy", sync.MirrorStrategy)
	}

	// Set alert attributes if present
	if len(sync.AlertAttributes) > 0 {
		if err := d.Set("alert", flattenAlerts(sync.AlertAttributes)); err != nil {
			fmt.Printf("[DEBUG] Failed to set alert: %v\n", err)
			return diag.Errorf("failed to set alert: %v", err)
		}
	}

	// Set time fields only if they are not zero values
	if !sync.CreatedAt.IsZero() {
		d.Set("created_at", sync.CreatedAt.Format("2006-01-02T15:04:05Z"))
	}
	if !sync.UpdatedAt.IsZero() {
		d.Set("updated_at", sync.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	}

	if sync.LastRunAt != nil {
		d.Set("last_run_at", sync.LastRunAt.Format("2006-01-02T15:04:05Z"))
	}
	if sync.NextRunAt != nil {
		d.Set("next_run_at", sync.NextRunAt.Format("2006-01-02T15:04:05Z"))
	}
	if sync.LastRunID != nil {
		d.Set("last_run_id", *sync.LastRunID)
	}

	// Handle run_mode from API response
	if sync.Mode != nil {
		fmt.Printf("[DEBUG] Setting run_mode from API response\n")
		if err := d.Set("run_mode", flattenRunMode(sync.Mode)); err != nil {
			fmt.Printf("[DEBUG] Failed to set run_mode: %v\n", err)
			return diag.Errorf("failed to set run_mode: %v", err)
		}
	} else if sync.ScheduleFrequency != "" {
		// Handle very old syncs that pre-date Mode API field (created before Census added Mode support)
		fmt.Printf("[DEBUG] Legacy sync detected - has flat schedule fields but no Mode\n")
		return diag.Errorf("This sync was created with an older version of the Census API that pre-dates run_mode support. Please recreate it using run_mode configuration or contact Census support to migrate it.")
	}

	// Set complex attributes with nil checks
	fmt.Printf("[DEBUG] Setting source_attributes\n")
	if err := d.Set("source_attributes", flattenSourceAttributes(sync.SourceAttributes)); err != nil {
		fmt.Printf("[DEBUG] Failed to set source_attributes: %v\n", err)
		return diag.Errorf("failed to set source_attributes: %v", err)
	}

	fmt.Printf("[DEBUG] Setting destination_attributes\n")
	if err := d.Set("destination_attributes", flattenDestinationAttributes(sync.DestinationAttributes)); err != nil {
		fmt.Printf("[DEBUG] Failed to set destination_attributes: %v\n", err)
		return diag.Errorf("failed to set destination_attributes: %v", err)
	}

	// Convert API Mappings back to Terraform FieldMappings with defensive handling
	fmt.Printf("[DEBUG] Converting field mappings - Mappings: %v, FieldMappings: %v\n", sync.Mappings != nil, sync.FieldMappings != nil)
	var fieldMappings []client.FieldMapping
	if sync.Mappings != nil && len(sync.Mappings) > 0 {
		fmt.Printf("[DEBUG] Using sync.Mappings (count: %d)\n", len(sync.Mappings))
		fieldMappings = convertMappingAttributesToFieldMappings(sync.Mappings)
	} else if sync.FieldMappings != nil {
		fmt.Printf("[DEBUG] Using sync.FieldMappings (count: %d)\n", len(sync.FieldMappings))
		fieldMappings = sync.FieldMappings // Fallback to legacy field
	} else {
		fmt.Printf("[DEBUG] Using empty field mappings\n")
		fieldMappings = []client.FieldMapping{} // Empty slice as fallback
	}

	fmt.Printf("[DEBUG] Setting field_mapping\n")
	if err := d.Set("field_mapping", flattenFieldMappings(fieldMappings)); err != nil {
		fmt.Printf("[DEBUG] Failed to set field_mapping: %v\n", err)
		return diag.Errorf("failed to set field_mapping: %v", err)
	}

	// Handle sync_key with nil check
	fmt.Printf("[DEBUG] Setting sync_key (nil: %v)\n", sync.SyncKey == nil)
	if sync.SyncKey != nil {
		if err := d.Set("sync_key", sync.SyncKey); err != nil {
			fmt.Printf("[DEBUG] Failed to set sync_key: %v\n", err)
			return diag.Errorf("failed to set sync_key: %v", err)
		}
	}

	// Schedule is already set above from flat API response fields

	fmt.Printf("[DEBUG] resourceSyncRead completed successfully\n")

	return nil
}

func resourceSyncUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fmt.Printf("[DEBUG] === Starting resourceSyncUpdate ===\n")

	apiClient := meta.(*client.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		fmt.Printf("[DEBUG] Error parsing sync ID: %v\n", err)
		return diag.Errorf("invalid sync ID: %s", d.Id())
	}
	fmt.Printf("[DEBUG] Updating sync with ID: %d\n", id)

	// Safe type assertion for workspace_id
	workspaceIdInterface := d.Get("workspace_id")
	workspaceId, ok := workspaceIdInterface.(string)
	if !ok {
		fmt.Printf("[DEBUG] workspace_id is not a string, type: %T, value: %+v\n", workspaceIdInterface, workspaceIdInterface)
		return diag.Errorf("workspace_id is not a valid string: %v", workspaceIdInterface)
	}

	workspaceIdInt, err := strconv.Atoi(workspaceId)
	if err != nil {
		fmt.Printf("[DEBUG] Error parsing workspace ID: %v\n", err)
		return diag.Errorf("invalid workspace ID: %s", workspaceId)
	}

	workspaceToken, err := apiClient.GetWorkspaceAPIKey(ctx, workspaceIdInt)
	if err != nil {
		fmt.Printf("[DEBUG] Error getting workspace token: %v\n", err)
		return diag.FromErr(err)
	}

	fmt.Printf("[DEBUG] Building update request...\n")

	// Handle run_mode
	var mode *client.SyncMode
	runModeRaw := d.Get("run_mode").([]interface{})

	if len(runModeRaw) > 0 {
		mode = expandRunMode(runModeRaw)
		fmt.Printf("[DEBUG] Using run_mode: %+v\n", mode)
	}

	// Safe type assertions for all fields
	labelInterface := d.Get("label")
	label, ok := labelInterface.(string)
	if !ok {
		fmt.Printf("[DEBUG] label is not a string, type: %T, value: %+v\n", labelInterface, labelInterface)
		return diag.Errorf("label is not a valid string: %v", labelInterface)
	}

	sourceAttrsInterface := d.Get("source_attributes")
	var sourceAttrs map[string]interface{}

	// Handle both map and list formats
	switch v := sourceAttrsInterface.(type) {
	case map[string]interface{}:
		sourceAttrs = v
	case []interface{}:
		if len(v) > 0 {
			if m, ok := v[0].(map[string]interface{}); ok {
				sourceAttrs = m
			} else {
				fmt.Printf("[DEBUG] source_attributes list element is not a map, type: %T, value: %+v\n", v[0], v[0])
				return diag.Errorf("source_attributes list element is not a valid map: %v", v[0])
			}
		} else {
			sourceAttrs = make(map[string]interface{})
		}
	default:
		fmt.Printf("[DEBUG] source_attributes is not a map or list, type: %T, value: %+v\n", sourceAttrsInterface, sourceAttrsInterface)
		return diag.Errorf("source_attributes is not a valid map or list: %v", sourceAttrsInterface)
	}

	// Additional processing to extract object field from list format
	if sourceAttrs != nil {
		if objData, exists := sourceAttrs["object"]; exists {
			switch v := objData.(type) {
			case []interface{}:
				// Object stored as list in Terraform state - extract first element
				if len(v) > 0 {
					if obj, ok := v[0].(map[string]interface{}); ok {
						fmt.Printf("[DEBUG] Extracted object from list in UPDATE: %+v\n", obj)
						sourceAttrs["object"] = obj
					} else {
						fmt.Printf("[DEBUG] object list element is not a map, type: %T, value: %+v\n", v[0], v[0])
					}
				}
			case map[string]interface{}:
				// Object is already a direct map - no change needed
				fmt.Printf("[DEBUG] object is already a map in UPDATE: %+v\n", v)
			default:
				fmt.Printf("[DEBUG] object is unexpected type in UPDATE: %T, value: %+v\n", v, v)
			}
		}
	}

	destAttrsInterface := d.Get("destination_attributes")
	var destAttrs map[string]interface{}

	// Handle both map and list formats
	switch v := destAttrsInterface.(type) {
	case map[string]interface{}:
		destAttrs = v
	case []interface{}:
		if len(v) > 0 {
			if m, ok := v[0].(map[string]interface{}); ok {
				destAttrs = m
			} else {
				fmt.Printf("[DEBUG] destination_attributes list element is not a map, type: %T, value: %+v\n", v[0], v[0])
				return diag.Errorf("destination_attributes list element is not a valid map: %v", v[0])
			}
		} else {
			destAttrs = make(map[string]interface{})
		}
	default:
		fmt.Printf("[DEBUG] destination_attributes is not a map or list, type: %T, value: %+v\n", destAttrsInterface, destAttrsInterface)
		return diag.Errorf("destination_attributes is not a valid map or list: %v", destAttrsInterface)
	}

	fieldMappingsInterface := d.Get("field_mapping")
	fieldMappings, ok := fieldMappingsInterface.([]interface{})
	if !ok {
		fmt.Printf("[DEBUG] field_mapping is not a []interface{}, type: %T, value: %+v\n", fieldMappingsInterface, fieldMappingsInterface)
		return diag.Errorf("field_mapping is not a valid list: %v", fieldMappingsInterface)
	}

	pausedInterface := d.Get("paused")
	paused, ok := pausedInterface.(bool)
	if !ok {
		fmt.Printf("[DEBUG] paused is not a bool, type: %T, value: %+v\n", pausedInterface, pausedInterface)
		return diag.Errorf("paused is not a valid boolean: %v", pausedInterface)
	}

	// Safe type assertions for field configuration
	fieldBehaviorInterface := d.Get("field_behavior")
	fieldBehavior, _ := fieldBehaviorInterface.(string)

	fieldNormalizationInterface := d.Get("field_normalization")
	fieldNormalization, _ := fieldNormalizationInterface.(string)

	fieldOrderInterface := d.Get("field_order")
	fieldOrder, _ := fieldOrderInterface.(string)

	syncBehaviorFamilyInterface := d.Get("sync_behavior_family")
	syncBehaviorFamily, _ := syncBehaviorFamilyInterface.(string)

	// Safe type assertions for alerts
	alertInterface := d.Get("alert")
	alertSet, ok := alertInterface.(*schema.Set)
	if !ok {
		fmt.Printf("[DEBUG] alert is not a *schema.Set, type: %T, value: %+v\n", alertInterface, alertInterface)
		return diag.Errorf("alert is not a valid set: %v", alertInterface)
	}

	req := &client.UpdateSyncRequest{
		Label:                 label,
		SourceAttributes:      expandSourceAttributes(d.Get("source_attributes").([]interface{})),
		DestinationAttributes: expandStringMap(destAttrs),
		FieldMappings:         expandFieldMappings(fieldMappings),
		Paused:                paused,

		// Mode - live vs triggered with trigger configurations
		Mode: mode,

		// Field configuration
		FieldBehavior:      fieldBehavior,
		FieldNormalization: fieldNormalization,
		FieldOrder:         fieldOrder,

		// Sync behavior family
		SyncBehaviorFamily: syncBehaviorFamily,

		// Advanced configuration
		AdvancedConfiguration: expandAdvancedConfiguration(d.Get("advanced_configuration").(string)),

		// High water mark attribute
		HighWaterMarkAttribute: d.Get("high_water_mark_attribute").(string),

		// Historical sync operation
		HistoricalSyncOperation: d.Get("historical_sync_operation").(string),

		// Mirror strategy
		MirrorStrategy: d.Get("mirror_strategy").(string),

		// Alert configuration
		AlertAttributes: expandAlerts(alertSet.List()),
	}

	fmt.Printf("[DEBUG] Update request: %+v\n", req)

	fmt.Printf("[DEBUG] Calling UpdateSyncWithToken...\n")
	_, err = apiClient.UpdateSyncWithToken(ctx, id, req, workspaceToken)
	if err != nil {
		fmt.Printf("[DEBUG] UpdateSyncWithToken failed: %v\n", err)
		return diag.FromErr(err)
	}

	fmt.Printf("[DEBUG] UpdateSyncWithToken succeeded, calling resourceSyncRead...\n")
	result := resourceSyncRead(ctx, d, meta)
	fmt.Printf("[DEBUG] === resourceSyncUpdate completed ===\n")
	return result
}

func resourceSyncDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*client.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("invalid sync ID: %s", d.Id())
	}

	workspaceId := d.Get("workspace_id").(string)
	workspaceIdInt, err := strconv.Atoi(workspaceId)
	if err != nil {
		return diag.Errorf("invalid workspace ID: %s", workspaceId)
	}

	workspaceToken, err := apiClient.GetWorkspaceAPIKey(ctx, workspaceIdInt)
	if err != nil {
		return diag.FromErr(err)
	}

	err = apiClient.DeleteSyncWithToken(ctx, id, workspaceToken)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

// Helper functions for expanding/flattening complex types

func expandFieldMappings(mappings []interface{}) []client.FieldMapping {
	result := make([]client.FieldMapping, 0, len(mappings))
	for i, mapping := range mappings {
		// Safe type assertion for mapping
		m, ok := mapping.(map[string]interface{})
		if !ok {
			fmt.Printf("[DEBUG] expandFieldMappings: mappings[%d] is not a map[string]interface{}, type: %T, value: %+v\n", i, mapping, mapping)
			continue // Skip invalid entries
		}

		fieldMapping := client.FieldMapping{
			Constant: m["constant"], // This is safe as interface{}
		}

		// Safe type assertions for string fields
		if from, ok := m["from"].(string); ok {
			fieldMapping.From = from
		} else {
			fmt.Printf("[DEBUG] expandFieldMappings: mappings[%d]['from'] is not a string, type: %T, value: %+v\n", i, m["from"], m["from"])
		}

		if to, ok := m["to"].(string); ok {
			fieldMapping.To = to
		} else {
			fmt.Printf("[DEBUG] expandFieldMappings: mappings[%d]['to'] is not a string, type: %T, value: %+v\n", i, m["to"], m["to"])
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
				fmt.Printf("[ERROR] expandFieldMappings: field_mapping[%d] has a constant value but type is '%s'. When using constant, type must be 'constant'.\n", i, mappingType)
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

func flattenFieldMappings(mappings []client.FieldMapping) []interface{} {
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

func expandAlerts(alerts []interface{}) []client.AlertAttribute {
	if len(alerts) == 0 {
		return nil
	}

	result := make([]client.AlertAttribute, 0, len(alerts))
	for i, alert := range alerts {
		m, ok := alert.(map[string]interface{})
		if !ok {
			fmt.Printf("[DEBUG] expandAlerts: alerts[%d] is not a map[string]interface{}, type: %T, value: %+v\n", i, alert, alert)
			continue
		}

		alertAttr := client.AlertAttribute{
			Options: make(map[string]interface{}),
		}

		if alertType, ok := m["type"].(string); ok {
			alertAttr.Type = alertType
		}

		if sendFor, ok := m["send_for"].(string); ok {
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

		result = append(result, alertAttr)
	}
	return result
}

func flattenAlerts(alerts []client.AlertAttribute) []interface{} {
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

func expandSyncSchedule(schedules []interface{}) *client.SyncSchedule {
	fmt.Printf("[DEBUG] expandSyncSchedule called with: %+v\n", schedules)

	if len(schedules) == 0 || schedules[0] == nil {
		fmt.Printf("[DEBUG] expandSyncSchedule returning nil (empty or nil schedule)\n")
		return nil
	}

	// Safe type assertion
	sInterface := schedules[0]
	s, ok := sInterface.(map[string]interface{})
	if !ok {
		fmt.Printf("[DEBUG] expandSyncSchedule: schedules[0] is not a map[string]interface{}, type: %T, value: %+v\n", sInterface, sInterface)
		return nil
	}
	fmt.Printf("[DEBUG] schedule map: %+v\n", s)

	// Safely extract values with defaults
	frequency := ""
	if freq, ok := s["frequency"]; ok && freq != nil {
		if freqStr, ok := freq.(string); ok {
			frequency = freqStr
		} else {
			fmt.Printf("[DEBUG] expandSyncSchedule: frequency is not a string, type: %T, value: %+v\n", freq, freq)
		}
	}

	minute := 0 // default
	if m, ok := s["minute"]; ok && m != nil {
		if minuteInt, ok := m.(int); ok {
			minute = minuteInt
		} else {
			fmt.Printf("[DEBUG] expandSyncSchedule: minute is not an int, type: %T, value: %+v\n", m, m)
		}
	}

	hour := 0 // default
	if h, ok := s["hour"]; ok && h != nil {
		if hourInt, ok := h.(int); ok {
			hour = hourInt
		} else {
			fmt.Printf("[DEBUG] expandSyncSchedule: hour is not an int, type: %T, value: %+v\n", h, h)
		}
	}

	dayOfWeek := 0 // default
	if dow, ok := s["day_of_week"]; ok && dow != nil {
		if dowInt, ok := dow.(int); ok {
			dayOfWeek = dowInt
		} else {
			fmt.Printf("[DEBUG] expandSyncSchedule: day_of_week is not an int, type: %T, value: %+v\n", dow, dow)
		}
	}

	timezone := "UTC" // default
	if tz, ok := s["timezone"]; ok && tz != nil {
		if tzStr, ok := tz.(string); ok {
			timezone = tzStr
		} else {
			fmt.Printf("[DEBUG] expandSyncSchedule: timezone is not a string, type: %T, value: %+v\n", tz, tz)
		}
	}

	result := &client.SyncSchedule{
		Frequency: frequency,
		Minute:    minute,
		Hour:      hour,
		DayOfWeek: dayOfWeek,
		Timezone:  timezone,
	}

	fmt.Printf("[DEBUG] expandSyncSchedule returning: %+v\n", result)
	return result
}

// expandRunMode converts Terraform run_mode config to API SyncMode struct
func expandRunMode(runModes []interface{}) *client.SyncMode {
	fmt.Printf("[DEBUG] expandRunMode called with: %+v\n", runModes)

	if len(runModes) == 0 || runModes[0] == nil {
		fmt.Printf("[DEBUG] expandRunMode returning nil (empty or nil run_mode)\n")
		return nil
	}

	runModeMap, ok := runModes[0].(map[string]interface{})
	if !ok {
		fmt.Printf("[DEBUG] expandRunMode: runModes[0] is not a map[string]interface{}, type: %T\n", runModes[0])
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

	fmt.Printf("[DEBUG] expandRunMode returning: %+v\n", mode)
	return mode
}

// flattenRunMode converts API SyncMode struct to Terraform run_mode config
func flattenRunMode(mode *client.SyncMode) []interface{} {
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

			if mode.Triggers.Schedule.Day != "" {
				scheduleMap["day"] = mode.Triggers.Schedule.Day
			}
			if mode.Triggers.Schedule.Hour != 0 {
				scheduleMap["hour"] = mode.Triggers.Schedule.Hour
			}
			if mode.Triggers.Schedule.Minute != 0 {
				scheduleMap["minute"] = mode.Triggers.Schedule.Minute
			}
			if mode.Triggers.Schedule.CronExpression != "" {
				scheduleMap["cron_expression"] = mode.Triggers.Schedule.CronExpression
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

func expandStringMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return nil
	}
	result := make(map[string]interface{})
	for k, v := range m {
		result[k] = v
	}
	return result
}

func flattenStringMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return nil
	}
	result := make(map[string]interface{})
	for k, v := range m {
		result[k] = v
	}
	return result
}

// cleanEmptyStrings removes empty string and zero values from a map
// to avoid sending invalid data to the Census API
func cleanEmptyStrings(m map[string]interface{}) map[string]interface{} {
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

func expandAdvancedConfiguration(jsonStr string) map[string]interface{} {
	if jsonStr == "" {
		return nil
	}
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		fmt.Printf("[DEBUG] Failed to unmarshal advanced_configuration: %v\n", err)
		return nil
	}
	return result
}

func flattenAdvancedConfiguration(m map[string]interface{}) string {
	if m == nil || len(m) == 0 {
		return ""
	}
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		fmt.Printf("[DEBUG] Failed to marshal advanced_configuration: %v\n", err)
		return ""
	}
	return string(jsonBytes)
}

// flattenSourceAttributes converts API source_attributes map to Terraform list structure
func flattenSourceAttributes(attrs map[string]interface{}) []map[string]interface{} {
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

func expandStringList(list []interface{}) []string {
	result := make([]string, 0, len(list))
	for i, v := range list {
		// Safe type assertion
		if str, ok := v.(string); ok {
			result = append(result, str)
		} else {
			fmt.Printf("[DEBUG] expandStringList: list[%d] is not a string, type: %T, value: %+v\n", i, v, v)
			// Skip non-string values instead of panicking
		}
	}
	return result
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

// validatePrimaryIdentifier ensures exactly one field mapping has is_primary_identifier = true
func validatePrimaryIdentifier(fieldMappings []client.FieldMapping) error {
	primaryCount := 0
	for _, fm := range fieldMappings {
		if fm.IsPrimaryIdentifier {
			primaryCount++
		}
	}

	if primaryCount == 0 {
		return fmt.Errorf("exactly one field_mapping must have is_primary_identifier = true, but found 0")
	}
	if primaryCount > 1 {
		return fmt.Errorf("exactly one field_mapping must have is_primary_identifier = true, but found %d", primaryCount)
	}

	return nil
}

// convertFieldMappingsToMappingAttributes converts Terraform FieldMapping to Census API MappingAttributes
func convertFieldMappingsToMappingAttributes(fieldMappings []client.FieldMapping) []client.MappingAttributes {
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
			// Liquid template transformation
			templateData := map[string]interface{}{
				"liquid_template": fm.LiquidTemplate,
			}
			mappingFrom = client.MappingFrom{
				Type: "liquid_template",
				Data: templateData,
			}

		default:
			// Default to column mapping (direct or hash)
			mappingFrom = client.MappingFrom{
				Type: "column",
				Data: fm.From,
			}
		}

		result[i] = client.MappingAttributes{
			From:                mappingFrom,
			To:                  fm.To,
			IsPrimaryIdentifier: fm.IsPrimaryIdentifier, // Use value from field_mapping directly
			LookupObject:        fm.LookupObject,
			LookupField:         fm.LookupField,
			PreserveValues:      fm.PreserveValues,
			GenerateField:       fm.GenerateField,
			SyncNullValues:      fm.SyncNullValues,
		}
	}

	return result
}

// convertMappingAttributesToFieldMappings converts Census API MappingAttributes back to Terraform FieldMapping
func convertMappingAttributesToFieldMappings(mappings []client.MappingAttributes) []client.FieldMapping {
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
				// Census API returns: {"liquid_template": "{{ record['field'] | upcase }}"}
				if dataMap, ok := ma.From.Data.(map[string]interface{}); ok {
					if template, ok := dataMap["liquid_template"].(string); ok {
						liquidTemplate = template
					}
				}
				from = "" // Empty from field for liquid_template

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

// expandSourceAttributes converts list-based source_attributes from Terraform to map format for API
func expandSourceAttributes(sourceAttrs []interface{}) map[string]interface{} {
	if len(sourceAttrs) == 0 {
		return nil
	}

	// Safe type assertion for source attributes
	attrInterface := sourceAttrs[0]
	attr, ok := attrInterface.(map[string]interface{})
	if !ok {
		fmt.Printf("[DEBUG] expandSourceAttributes: sourceAttrs[0] is not a map[string]interface{}, type: %T, value: %+v\n", attrInterface, attrInterface)
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
					fmt.Printf("[DEBUG] expandSourceAttributes: object extracted from list: %+v\n", obj)
					objectMap = obj
				} else {
					fmt.Printf("[DEBUG] expandSourceAttributes: objList[0] is not a map[string]interface{}, type: %T, value: %+v\n", v[0], v[0])
					return result // Return partial result instead of nil
				}
			}
		case map[string]interface{}:
			// Object is directly a map (direct config)
			fmt.Printf("[DEBUG] expandSourceAttributes: object is direct map: %+v\n", v)
			objectMap = v
		default:
			fmt.Printf("[DEBUG] expandSourceAttributes: object is unexpected type: %T, value: %+v\n", objData, objData)
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
			fmt.Printf("[DEBUG] expandSourceAttributes: Translated segment source - object: %+v, filter_segment_id: %+v\n", translatedObject, result["filter_segment_id"])

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
			fmt.Printf("[DEBUG] expandSourceAttributes: Translated cohort source - object: %+v, cohort_id: %+v\n", translatedObject, result["cohort_id"])

		default:
			// For all other types (model, topic, dataset, table), pass through as-is
			// But clean empty strings to avoid API errors
			result["object"] = cleanEmptyStrings(objectMap)
			fmt.Printf("[DEBUG] expandSourceAttributes: Pass-through object for type %s (cleaned): %+v\n", objectType, result["object"])
		}
	}

	// Remove cohort_id if it's 0 (default value from Terraform, not actually set)
	if cohortId, ok := result["cohort_id"].(int); ok && cohortId == 0 {
		delete(result, "cohort_id")
	}

	return result
}

// expandDestinationAttributes converts list-based destination_attributes from Terraform to map format for API
func expandDestinationAttributes(destAttrs []interface{}) map[string]interface{} {
	if len(destAttrs) == 0 {
		return nil
	}

	// Safe type assertion for destination attributes
	attrInterface := destAttrs[0]
	attr, ok := attrInterface.(map[string]interface{})
	if !ok {
		fmt.Printf("[DEBUG] expandDestinationAttributes: destAttrs[0] is not a map[string]interface{}, type: %T, value: %+v\n", attrInterface, attrInterface)
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

// flattenDestinationAttributes converts API destination_attributes map to Terraform list structure
func flattenDestinationAttributes(attrs map[string]interface{}) []map[string]interface{} {
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

func resourceSyncImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// Support composite format: workspace_id:sync_id
	parts := strings.Split(d.Id(), ":")

	if len(parts) == 2 {
		// Format: workspace_id:sync_id
		workspaceId := parts[0]
		syncId := parts[1]

		d.SetId(syncId)
		d.Set("workspace_id", workspaceId)

		return []*schema.ResourceData{d}, nil
	} else if len(parts) == 1 {
		// Legacy format - provide helpful error
		return nil, fmt.Errorf(`import requires workspace_id. Use format: workspace_id:sync_id

Example:
  terraform import census_sync.contact_sync 69962:123

Where 69962 is the workspace_id and 123 is the sync_id.`)
	}

	return nil, fmt.Errorf("invalid import format. Use: workspace_id:sync_id")
}
