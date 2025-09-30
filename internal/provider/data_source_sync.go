package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/sutrolabs/terraform-provider-census/internal/client"
)

func dataSourceSync() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves information about a Census sync.",
		ReadContext: dataSourceSyncRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the sync.",
			},
			"workspace_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the workspace this sync belongs to.",
			},
			"label": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name/label of the sync.",
			},
			"source_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The ID of the source to sync data from.",
			},
			"destination_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The ID of the destination to sync data to.",
			},
			"source_attributes": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Source-specific configuration (e.g., SQL query, table selection).",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"destination_attributes": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Destination-specific configuration (e.g., object, operation mode).",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"field_mappings": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Field mappings between source and destination.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"from": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Source field name.",
						},
						"to": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Destination field name.",
						},
						"operation": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Mapping operation (direct, hash, etc.).",
						},
						"constant": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Constant value for constant operations.",
						},
					},
				},
			},
			"sync_key": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Fields that uniquely identify records for syncing.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"paused": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the sync is paused.",
			},
			"schedule": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Sync scheduling configuration.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"frequency": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Sync frequency (hourly, daily, weekly).",
						},
						"interval": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Run every N frequency units.",
						},
						"hour": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Hour to run (for daily/weekly schedules, 0-23).",
						},
						"day_of_week": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Day of week to run (for weekly schedules, 0=Sunday).",
						},
						"timezone": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Timezone for scheduling.",
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

func dataSourceSyncRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*client.Client)

	syncID, err := strconv.Atoi(d.Get("id").(string))
	if err != nil {
		return diag.Errorf("invalid sync ID: %s", d.Get("id").(string))
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

	sync, err := apiClient.GetSyncWithToken(ctx, syncID, workspaceToken)
	if err != nil {
		return diag.FromErr(err)
	}

	// Check if sync is nil (API returned successfully but with nil data)
	if sync == nil {
		return diag.Errorf("sync not found: %d", syncID)
	}

	d.SetId(strconv.Itoa(sync.ID))
	d.Set("workspace_id", workspaceId)
	d.Set("label", sync.Label)
	d.Set("status", sync.Status)
	d.Set("paused", sync.Paused)
	d.Set("created_at", sync.CreatedAt.Format("2006-01-02T15:04:05Z"))
	d.Set("updated_at", sync.UpdatedAt.Format("2006-01-02T15:04:05Z"))

	if sync.LastRunAt != nil {
		d.Set("last_run_at", sync.LastRunAt.Format("2006-01-02T15:04:05Z"))
	}
	if sync.NextRunAt != nil {
		d.Set("next_run_at", sync.NextRunAt.Format("2006-01-02T15:04:05Z"))
	}
	if sync.LastRunID != nil {
		d.Set("last_run_id", *sync.LastRunID)
	}

	// Set complex attributes
	if err := d.Set("source_attributes", flattenStringMap(sync.SourceAttributes)); err != nil {
		return diag.Errorf("failed to set source_attributes: %v", err)
	}
	if err := d.Set("destination_attributes", flattenStringMap(sync.DestinationAttributes)); err != nil {
		return diag.Errorf("failed to set destination_attributes: %v", err)
	}
	if err := d.Set("field_mappings", flattenFieldMappings(sync.FieldMappings)); err != nil {
		return diag.Errorf("failed to set field_mappings: %v", err)
	}
	if err := d.Set("sync_key", sync.SyncKey); err != nil {
		return diag.Errorf("failed to set sync_key: %v", err)
	}
	if err := d.Set("schedule", flattenSyncSchedule(sync.Schedule)); err != nil {
		return diag.Errorf("failed to set schedule: %v", err)
	}

	return nil
}
