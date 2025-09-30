package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/sutrolabs/terraform-provider-census/internal/client"
)

func dataSourceWorkspace() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves information about a Census workspace.",

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
	apiClient := meta.(*client.Client)

	id, err := strconv.Atoi(d.Get("id").(string))
	if err != nil {
		return diag.Errorf("invalid workspace ID: %s", d.Get("id").(string))
	}

	workspace, err := apiClient.GetWorkspace(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}

	// Check if workspace is nil (API returned successfully but with nil data)
	if workspace == nil {
		return diag.Errorf("workspace not found: %d", id)
	}

	d.SetId(strconv.Itoa(workspace.ID))
	d.Set("name", workspace.Name)
	d.Set("organization_id", workspace.OrganizationID)
	d.Set("created_at", workspace.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))

	if err := d.Set("notification_emails", workspace.NotificationEmails); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
