package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceWorkspace() *schema.Resource {
	return &schema.Resource{
		DeprecationMessage: "This resource is deprecated and no longer manages anything in Fivetran Activations. Workspaces are no longer needed; remove it from your configuration.",

		Description: "Previously managed a workspace in Fivetran Activations (Census).",

		CreateContext: resourceWorkspaceCreate,
		ReadContext:   resourceWorkspaceRead,
		UpdateContext: resourceWorkspaceUpdate,
		DeleteContext: resourceWorkspaceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the workspace.",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The name of the workspace. Must be unique within the organization.",
			},
			"notification_emails": {
				Type:        schema.TypeSet,
				Optional:    true,
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
			"api_key": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The API key of the workspace (only available during creation if return_workspace_api_key is true).",
			},
			"return_workspace_api_key": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether to return the workspace API key in the response during creation.",
			},
		},
	}
}

func resourceWorkspaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.Errorf("census_workspace can no longer be created; workspace management has been removed from this provider")
}

func resourceWorkspaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceWorkspaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceWorkspaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
