package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/your-org/terraform-provider-census/internal/client"
)

func dataSourceSource() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get information about a Census source connection.",

		ReadContext: dataSourceSourceRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the source to retrieve.",
			},
			"workspace_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the workspace this source belongs to.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the source connection.",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The type of source (e.g., snowflake, bigquery, postgres).",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The current status of the source connection.",
			},
			"test_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the last connection test.",
			},
			"last_tested": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp when the source was last tested.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp when the source was created.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp when the source was last updated.",
			},
		},
	}
}

func dataSourceSourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*client.Client)

	id, err := strconv.Atoi(d.Get("id").(string))
	if err != nil {
		return diag.Errorf("invalid source ID: %s", d.Get("id").(string))
	}

	// Get the source using WithToken method for PAT-only authentication
	// Note: Data sources will need to be provided with workspace_id in the future 
	// or we can derive it from the source itself
	source, err := apiClient.GetSourceWithToken(ctx, id, "")
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(source.ID))
	d.Set("workspace_id", source.WorkspaceID)
	d.Set("name", source.Name)
	d.Set("type", source.Type)
	d.Set("status", source.Status)
	d.Set("test_status", source.TestStatus)
	d.Set("created_at", source.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
	d.Set("updated_at", source.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"))

	if source.LastTested != nil {
		d.Set("last_tested", source.LastTested.Format("2006-01-02T15:04:05Z07:00"))
	}

	return nil
}