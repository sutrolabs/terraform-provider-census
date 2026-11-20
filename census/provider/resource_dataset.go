package provider

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/sutrolabs/terraform-provider-census/census/client/v2"
)

func resourceDataset() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Census SQL dataset for data transformation and modeling.",

		CreateContext: resourceDatasetCreate,
		ReadContext:   resourceDatasetRead,
		UpdateContext: resourceDatasetUpdate,
		DeleteContext: resourceDatasetDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the dataset.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the dataset.",
			},
			"type": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "sql",
				ForceNew:    true,
				Description: "The type of dataset (currently only 'sql' is supported).",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Optional description of the dataset.",
			},
			"query": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "SQL query that defines the dataset.",
			},
			"group_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The group ID of the source connection to run the query against.",
				// State written before group_id existed has none until the
				// next refresh reads it back; that is not a reason to replace.
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return old == "" && d.Id() != ""
				},
			},
			// Computed fields
			"resource_identifier": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Resource identifier for the dataset.",
			},
			"cached_record_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Cached count of records in the dataset.",
			},
			"columns": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Columns in the dataset.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the column.",
						},
						"data_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The data type of the column.",
						},
					},
				},
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp when the dataset was created.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp when the dataset was last updated.",
			},
			"wait_for_metadata_refresh": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Wait for metadata refresh to complete before returning. Set to true when creating syncs immediately after dataset creation to avoid timeout issues. Changing from false to true on an existing dataset will trigger metadata refresh (skipped if metadata is already available). Defaults to false. In general, this does not need to be set to true unless the dataset takes a long time (>60s) to query column metadata for from the source.",
			},
			"metadata_ready": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether the dataset metadata (columns) has been refreshed.",
			},
		},
	}
}

func waitForMetadataRefresh(ctx context.Context, apiClient *v2.Client, datasetID int) diag.Diagnostics {
	refreshKey, err := v2.RefreshDatasetColumns(ctx, apiClient, datasetID)
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to trigger metadata refresh",
				Detail: fmt.Sprintf(
					"Dataset %d exists, but metadata refresh could not be triggered: %v\n\n"+
						"To resolve:\n"+
						"1. Run 'terraform apply' again to retry metadata refresh\n"+
						"2. Or remove 'wait_for_metadata_refresh = true' and manually trigger metadata refresh in the Census UI or API\n",
					datasetID, err),
			},
		}
	}

	// Poll until metadata refresh completes using StateChangeConf.
	// A failed refresh comes back as an API error rather than a status.
	stateConf := &retry.StateChangeConf{
		Pending: []string{"processing"},
		Target:  []string{"completed"},
		Refresh: func() (interface{}, string, error) {
			status, err := v2.GetDatasetRefreshStatus(ctx, apiClient, datasetID, refreshKey)
			if err != nil {
				return nil, "", fmt.Errorf("metadata refresh failed: %w", err)
			}
			return status, status, nil
		},
		Timeout:    30 * time.Minute, // Default 30 minutes
		Delay:      5 * time.Second,  // Wait 5 seconds before first poll
		MinTimeout: 3 * time.Second,  // Minimum 3 seconds between polls
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Metadata refresh did not complete in time",
				Detail: fmt.Sprintf(
					"Dataset %d metadata refresh did not complete: %v\n\n"+
						"To resolve:\n"+
						"1. Run 'terraform apply' again to wait for metadata refresh to complete\n"+
						"2. Check metadata refresh status in the Census UI\n",
					datasetID, err),
			},
		}
	}

	return nil
}

func resourceDatasetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*clients).V2

	req := &v2.CreateDatasetRequest{
		Name:    d.Get("name").(string),
		Type:    d.Get("type").(string),
		Query:   d.Get("query").(string),
		GroupID: d.Get("group_id").(string),
	}
	if desc, ok := d.GetOk("description"); ok {
		description := desc.(string)
		req.Description = &description
	}

	dataset, err := v2.CreateDataset(ctx, apiClient, req)
	if err != nil {
		if v2.IsNotFoundError(err) {
			return diag.Errorf("no source connection with group_id %q: %v", req.GroupID, err)
		}
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(dataset.ID))

	if d.Get("wait_for_metadata_refresh").(bool) {
		if diags := waitForMetadataRefresh(ctx, apiClient, dataset.ID); diags != nil {
			return diags
		}
	}

	return resourceDatasetRead(ctx, d, meta)
}

func resourceDatasetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*clients).V2

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("invalid dataset ID: %s", d.Id())
	}

	dataset, err := v2.GetDataset(ctx, apiClient, id)
	if err != nil {
		if v2.IsNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	if err := setDatasetAttributes(d, dataset); err != nil {
		return diag.FromErr(err)
	}
	d.Set("metadata_ready", len(dataset.Columns) > 0)

	return nil
}

func resourceDatasetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*clients).V2

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("invalid dataset ID: %s", d.Id())
	}

	req := &v2.UpdateDatasetRequest{}
	if d.HasChange("name") {
		name := d.Get("name").(string)
		req.Name = &name
	}
	if d.HasChange("description") {
		description := d.Get("description").(string)
		req.Description = &description
	}
	if d.HasChange("query") {
		query := d.Get("query").(string)
		req.Query = &query
	}

	if req.Name != nil || req.Description != nil || req.Query != nil {
		if _, err := v2.UpdateDataset(ctx, apiClient, id, req); err != nil {
			return diag.FromErr(err)
		}
	}

	// If wait_for_metadata_refresh changed from false to true and metadata isn't ready,
	// trigger the metadata refresh flow
	if d.HasChange("wait_for_metadata_refresh") {
		old, new := d.GetChange("wait_for_metadata_refresh")
		if !old.(bool) && new.(bool) && !d.Get("metadata_ready").(bool) {
			if diags := waitForMetadataRefresh(ctx, apiClient, id); diags != nil {
				return diags
			}
		}
	}

	return resourceDatasetRead(ctx, d, meta)
}

func resourceDatasetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*clients).V2

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("invalid dataset ID: %s", d.Id())
	}

	if err := v2.DeleteDataset(ctx, apiClient, id); err != nil {
		if v2.IsNotFoundError(err) {
			return nil
		}
		return diag.FromErr(err)
	}

	return nil
}

// setDatasetAttributes sets the attributes the resource and data source share.
func setDatasetAttributes(d *schema.ResourceData, dataset *v2.Dataset) error {
	d.Set("name", dataset.Name)
	d.Set("type", dataset.Type)
	d.Set("query", dataset.Query)
	d.Set("group_id", dataset.GroupID)
	d.Set("resource_identifier", dataset.ResourceIdentifier)

	if dataset.Description != nil {
		d.Set("description", *dataset.Description)
	}
	if dataset.CachedRecordCount != nil {
		d.Set("cached_record_count", *dataset.CachedRecordCount)
	}

	if !dataset.CreatedAt.IsZero() {
		d.Set("created_at", dataset.CreatedAt.Format("2006-01-02T15:04:05Z"))
	}
	if !dataset.UpdatedAt.IsZero() {
		d.Set("updated_at", dataset.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	}

	// Always set columns, as an empty list when there are none, to avoid drift.
	columns := make([]map[string]interface{}, len(dataset.Columns))
	for i, col := range dataset.Columns {
		columns[i] = map[string]interface{}{
			"name":      col.Name,
			"data_type": col.DataType,
		}
	}
	if err := d.Set("columns", columns); err != nil {
		return fmt.Errorf("failed to set columns: %w", err)
	}

	return nil
}
