package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/sutrolabs/terraform-provider-census/census/client"
)

func resourceDataset() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Census SQL dataset for data transformation and modeling.",

		CreateContext: resourceDatasetCreate,
		ReadContext:   resourceDatasetRead,
		UpdateContext: resourceDatasetUpdate,
		DeleteContext: resourceDatasetDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceDatasetImport,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the dataset.",
			},
			"workspace_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the workspace this dataset belongs to.",
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
			"source_id": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the source connection to run the query against.",
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
		},
	}
}

func resourceDatasetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	// Build description pointer
	var description *string
	if desc, ok := d.GetOk("description"); ok {
		descStr := desc.(string)
		description = &descStr
	}

	req := &client.CreateDatasetRequest{
		Name:        d.Get("name").(string),
		Type:        d.Get("type").(string),
		Description: description,
		Query:       d.Get("query").(string),
		SourceID:    d.Get("source_id").(int),
	}

	dataset, err := apiClient.CreateDatasetWithToken(ctx, req, workspaceToken)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(dataset.ID))

	// Explicitly set workspace_id from our input since API doesn't return it
	d.Set("workspace_id", workspaceId)

	return resourceDatasetRead(ctx, d, meta)
}

func resourceDatasetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*client.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("invalid dataset ID: %s", d.Id())
	}

	// Get workspace token dynamically if we have workspace_id
	workspaceId := d.Get("workspace_id").(string)

	var dataset *client.Dataset
	if workspaceId != "" {
		workspaceIdInt, err := strconv.Atoi(workspaceId)
		if err != nil {
			return diag.Errorf("invalid workspace ID: %s", workspaceId)
		}

		workspaceToken, err := apiClient.GetWorkspaceAPIKey(ctx, workspaceIdInt)
		if err != nil {
			return diag.FromErr(err)
		}

		dataset, err = apiClient.GetDatasetWithToken(ctx, id, workspaceToken)
	} else {
		return diag.Errorf(`workspace_id is required but missing from resource state.

To fix this, add the missing workspace_id to terraform state:
  terraform state rm census_dataset.example
  terraform import census_dataset.example workspace_id:dataset_id`)
	}

	if err != nil {
		// Check if dataset was not found
		if IsNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	// Check if dataset is nil (API returned successfully but with nil data)
	if dataset == nil {
		d.SetId("")
		return nil
	}

	// Only update workspace_id if API returned it, otherwise preserve what's in state
	if dataset.WorkspaceID != "" {
		d.Set("workspace_id", dataset.WorkspaceID)
	}

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

	// Set time fields only if they are not zero values
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

func resourceDatasetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*client.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("invalid dataset ID: %s", d.Id())
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

	req := &client.UpdateDatasetRequest{}

	// Only include changed fields
	if d.HasChange("name") {
		name := d.Get("name").(string)
		req.Name = &name
	}

	if d.HasChange("description") {
		desc := d.Get("description").(string)
		req.Description = &desc
	}

	if d.HasChange("query") {
		query := d.Get("query").(string)
		req.Query = &query
	}

	_, err = apiClient.UpdateDatasetWithToken(ctx, id, req, workspaceToken)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceDatasetRead(ctx, d, meta)
}

func resourceDatasetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*client.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("invalid dataset ID: %s", d.Id())
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

	err = apiClient.DeleteDatasetWithToken(ctx, id, workspaceToken)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceDatasetImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// Support composite format: workspace_id:dataset_id
	parts := strings.Split(d.Id(), ":")

	if len(parts) == 2 {
		// Format: workspace_id:dataset_id
		workspaceId := parts[0]
		datasetId := parts[1]

		d.SetId(datasetId)
		d.Set("workspace_id", workspaceId)

		return []*schema.ResourceData{d}, nil
	} else if len(parts) == 1 {
		// Legacy format - provide helpful error
		return nil, fmt.Errorf(`import requires workspace_id. Use format: workspace_id:dataset_id

Example:
  terraform import census_dataset.all_users 69962:789

Where 69962 is the workspace_id and 789 is the dataset_id.`)
	}

	return nil, fmt.Errorf("invalid import format. Use: workspace_id:dataset_id")
}
