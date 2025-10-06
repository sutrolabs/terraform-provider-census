package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/sutrolabs/terraform-provider-census/census/client"
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
				Required:    true,
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

	workspaceId := d.Get("workspace_id").(string)
	workspaceIdInt, err := strconv.Atoi(workspaceId)
	if err != nil {
		return diag.Errorf("invalid workspace ID: %s", workspaceId)
	}

	// Get workspace token using personal access token
	workspaceToken, err := apiClient.GetWorkspaceAPIKey(ctx, workspaceIdInt)
	if err != nil {
		return diag.Errorf("failed to get workspace API key for workspace %d: %v", workspaceIdInt, err)
	}

	// Get the source using the workspace token
	source, err := apiClient.GetSourceWithToken(ctx, id, workspaceToken)
	if err != nil {
		return diag.FromErr(err)
	}

	// Check if source is nil (API returned successfully but with nil data)
	if source == nil {
		return diag.Errorf("source not found: %d", id)
	}

	d.SetId(strconv.Itoa(source.ID))
	// Note: workspace_id is a Required input field, don't overwrite it with API response
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
