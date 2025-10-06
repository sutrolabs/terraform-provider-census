package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/sutrolabs/terraform-provider-census/census/client"
)

func dataSourceDataset() *schema.Resource {
	return &schema.Resource{
		Description: "Fetches information about a Census SQL dataset.",

		ReadContext: dataSourceDatasetRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the dataset.",
			},
			"workspace_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the workspace this dataset belongs to.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the dataset.",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The type of dataset.",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of the dataset.",
			},
			"query": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "SQL query that defines the dataset.",
			},
			"source_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "ID of the source connection.",
			},
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
		},
	}
}

func dataSourceDatasetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*client.Client)

	id, err := strconv.Atoi(d.Get("id").(string))
	if err != nil {
		return diag.Errorf("invalid dataset ID: %s", d.Get("id").(string))
	}

	workspaceId := d.Get("workspace_id").(string)
	workspaceIdInt, err := strconv.Atoi(workspaceId)
	if err != nil {
		return diag.Errorf("invalid workspace ID: %s", workspaceId)
	}

	workspaceToken, err := apiClient.GetWorkspaceAPIKey(ctx, workspaceIdInt)
	if err != nil {
		return diag.Errorf("failed to get workspace API key: %v", err)
	}

	dataset, err := apiClient.GetDatasetWithToken(ctx, id, workspaceToken)
	if err != nil {
		return diag.FromErr(err)
	}

	if dataset == nil {
		return diag.Errorf("dataset not found: %d", id)
	}

	d.SetId(strconv.Itoa(dataset.ID))
	d.Set("workspace_id", workspaceId)
	d.Set("name", dataset.Name)
	d.Set("type", dataset.Type)
	d.Set("query", dataset.Query)
	d.Set("source_id", dataset.SourceID)
	d.Set("resource_identifier", dataset.ResourceIdentifier)

	// Set optional fields with nil checks
	if dataset.Description != nil {
		d.Set("description", *dataset.Description)
	}

	if dataset.CachedRecordCount != nil {
		d.Set("cached_record_count", *dataset.CachedRecordCount)
	}

	// Set time fields
	if !dataset.CreatedAt.IsZero() {
		d.Set("created_at", dataset.CreatedAt.Format("2006-01-02T15:04:05Z"))
	}
	if !dataset.UpdatedAt.IsZero() {
		d.Set("updated_at", dataset.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	}

	// Set columns - always set to avoid drift, use empty list if no columns
	columns := make([]map[string]interface{}, 0)
	if dataset.Columns != nil && len(dataset.Columns) > 0 {
		columns = make([]map[string]interface{}, len(dataset.Columns))
		for i, col := range dataset.Columns {
			columns[i] = map[string]interface{}{
				"name":      col.Name,
				"data_type": col.DataType,
			}
		}
	}
	if err := d.Set("columns", columns); err != nil {
		return diag.Errorf("failed to set columns: %v", err)
	}

	return nil
}
