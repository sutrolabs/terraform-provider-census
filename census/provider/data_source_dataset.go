package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/sutrolabs/terraform-provider-census/census/client/v2"
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
			"group_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The group ID of the source connection the query runs against.",
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
	apiClient := meta.(*clients).V2

	id, err := strconv.Atoi(d.Get("id").(string))
	if err != nil {
		return diag.Errorf("invalid dataset ID: %s", d.Get("id").(string))
	}

	dataset, err := v2.GetDataset(ctx, apiClient, id)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(dataset.ID))
	if err := setDatasetAttributes(d, dataset); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
