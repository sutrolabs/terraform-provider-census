package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceWorkspace() *schema.Resource {
	return &schema.Resource{
		DeprecationMessage: "This data source is deprecated and no longer returns data from Fivetran Activations. Workspaces are no longer needed; remove it from your configuration.",

		Description: "Previously retrieved a workspace from Fivetran Activations (Census).",

		ReadContext: dataSourceWorkspaceRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the workspace.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the workspace.",
			},
			"notification_emails": {
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Description: "The list of emails that will receive alerts from the workspace.",
			},
			"organization_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The ID of the organization the workspace belongs to.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The timestamp when the workspace was created.",
			},
		},
	}
}

func dataSourceWorkspaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.SetId(d.Get("id").(string))

	return diag.Diagnostics{{
		Severity: diag.Warning,
		Summary:  "census_workspace data source is deprecated",
		Detail:   "This data source no longer queries Fivetran Activations and returns only its id. Workspaces are no longer needed; remove it from your configuration.",
	}}
}
