package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/sutrolabs/terraform-provider-census/internal/client"
)

func dataSourceDestination() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get information about a Census destination connection.",

		ReadContext: dataSourceDestinationRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the destination to retrieve.",
			},
			"workspace_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the workspace this destination belongs to.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the destination connection.",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The type of destination (e.g., salesforce, hubspot, postgres).",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The current status of the destination connection.",
			},
			"test_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the last connection test.",
			},
			"last_tested": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp when the destination was last tested.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp when the destination was created.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp when the destination was last updated.",
			},
		},
	}
}

func dataSourceDestinationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*client.Client)

	id, err := strconv.Atoi(d.Get("id").(string))
	if err != nil {
		return diag.Errorf("invalid destination ID: %s", d.Get("id").(string))
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

	// Get the destination using the workspace token
	destination, err := apiClient.GetDestinationWithToken(ctx, id, workspaceToken)
	if err != nil {
		return diag.FromErr(err)
	}

	// Check if destination is nil (API returned successfully but with nil data)
	if destination == nil {
		return diag.Errorf("destination not found: %d", id)
	}

	d.SetId(strconv.Itoa(destination.ID))
	// Note: workspace_id is a Required input field, don't overwrite it with API response
	d.Set("name", destination.Name)
	d.Set("type", destination.Type)
	d.Set("status", destination.Status)
	d.Set("test_status", destination.TestStatus)
	d.Set("created_at", destination.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
	d.Set("updated_at", destination.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"))

	if destination.LastTested != nil {
		d.Set("last_tested", destination.LastTested.Format("2006-01-02T15:04:05Z07:00"))
	}

	return nil
}
