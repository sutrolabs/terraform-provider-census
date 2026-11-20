package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/sutrolabs/terraform-provider-census/census/client/v2"
)

func dataSourceSync() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves information about a Fivetran Activations sync.",
		ReadContext: dataSourceSyncRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the sync.",
			},
			"activation_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the activation this sync belongs to.",
			},
			"label": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name/label of the sync.",
			},
			"source_object": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Source object configuration (JSON string for complex objects like table/dataset).",
			},
			"destination_object": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The destination object name.",
			},
			"paused": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the sync is paused.",
			},
			"run_mode": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Run mode configuration for the sync (live vs triggered with various trigger types).",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Mode type: 'live' for continuous syncing or 'triggered' for event-based syncing.",
						},
						"triggers": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Trigger configurations (only for 'triggered' mode).",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"schedule": {
										Type:        schema.TypeList,
										Computed:    true,
										Description: "Schedule-based trigger configuration.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"frequency": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "Sync frequency: never, continuous, quarter_hourly, hourly, daily, weekly, or expression (for cron).",
												},
												"day": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "Day of week (Sunday-Saturday, for weekly schedules).",
												},
												"hour": {
													Type:        schema.TypeInt,
													Computed:    true,
													Description: "Hour to run (0-24).",
												},
												"minute": {
													Type:        schema.TypeInt,
													Computed:    true,
													Description: "Minute to run (0-59).",
												},
												"cron_expression": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "Cron expression (only valid when frequency is 'expression').",
												},
											},
										},
									},
									"dbt_cloud": {
										Type:        schema.TypeList,
										Computed:    true,
										Description: "dbt Cloud job trigger configuration.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"project_id": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "dbt Cloud project ID.",
												},
												"job_id": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "dbt Cloud job ID.",
												},
											},
										},
									},
									"fivetran": {
										Type:        schema.TypeList,
										Computed:    true,
										Description: "Fivetran connector trigger configuration.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"job_id": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "Fivetran job ID.",
												},
												"job_name": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "Fivetran job name.",
												},
											},
										},
									},
									"sync_sequence": {
										Type:        schema.TypeList,
										Computed:    true,
										Description: "Sync dependency trigger configuration (triggers after another sync completes).",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"sync_id": {
													Type:        schema.TypeInt,
													Computed:    true,
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
			// Computed status fields
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
		},
	}
}

func dataSourceSyncRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*clients).V2

	syncID, err := strconv.Atoi(d.Get("id").(string))
	if err != nil {
		return diag.Errorf("invalid sync ID: %s", d.Get("id").(string))
	}

	sync, err := v2.GetSync(ctx, apiClient, syncID)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(sync.ID))
	d.Set("activation_id", sync.ActivationID)
	d.Set("label", sync.Label)
	d.Set("status", sync.Status)
	d.Set("paused", sync.Paused)
	d.Set("created_at", sync.CreatedAt.Format(time.RFC3339))
	d.Set("updated_at", sync.UpdatedAt.Format(time.RFC3339))

	// last_run_at, next_run_at and last_run_id are deprecated: the API does not support them,
	// so they are never set and stay empty.

	// Set complex attributes
	source_object, err := flattenSourceAttributesForDataSource(sync.SourceAttributes)
	if err != nil {
		return diag.Errorf("failed to read source_object: %v", err)
	}

	destination_object, err := flattenDestinationAttributesForDataSource(sync.DestinationAttributes)
	if err != nil {
		return diag.Errorf("failed to read destination_object: %v", err)
	}

	d.Set("source_object", source_object)
	d.Set("destination_object", destination_object)

	// field_mapping is deprecated: the API does not support it, so it is always an empty list.
	if err := d.Set("field_mapping", []interface{}{}); err != nil {
		return diag.Errorf("failed to set field_mapping: %v", err)
	}
	if sync.Mode != nil {
		if err := d.Set("run_mode", FlattenRunMode(sync.Mode)); err != nil {
			return diag.Errorf("failed to set run_mode: %v", err)
		}
	}

	return nil
}

// flattenSourceAttributesForDataSource converts API source_attributes map to Terraform TypeList format
func flattenSourceAttributesForDataSource(attrs map[string]interface{}) (string, error) {
	if attrs == nil {
		return "", nil
	}

	// Serialize object as JSON string
	if obj, ok := attrs["object"]; ok {
		if objMap, isMap := obj.(map[string]interface{}); isMap {
			jsonBytes, err := json.Marshal(objMap)
			if err == nil {
				return string(jsonBytes), nil
			} else {
				return "", err
			}
		} else {
			return fmt.Sprintf("%v", obj), nil
		}
	}

	return "", nil
}

// flattenDestinationAttributesForDataSource converts API destination_attributes map to Terraform TypeList format
func flattenDestinationAttributesForDataSource(attrs *v2.SyncDestinationAttributes) (string, error) {
	if attrs == nil {
		return "", nil
	}

	return fmt.Sprintf("%v", attrs.Object), nil
}
