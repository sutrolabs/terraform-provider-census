package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/sutrolabs/terraform-provider-census/internal/client"
)

func resourceWorkspace() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Census workspace.",

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
				Required:    true,
				Description: "The name of the workspace. Must be unique within the organization.",
			},
			"notification_emails": {
				Type:        schema.TypeSet,
				Optional:    true,
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
	apiClient := meta.(*client.Client)

	name := d.Get("name").(string)
	notificationEmails := expandStringSet(d.Get("notification_emails").(*schema.Set))
	returnAPIKey := d.Get("return_workspace_api_key").(bool)

	req := &client.CreateWorkspaceRequest{
		Name:                  name,
		NotificationEmails:    notificationEmails,
		ReturnWorkspaceAPIKey: returnAPIKey,
	}

	workspace, err := apiClient.CreateWorkspace(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(workspace.ID))

	// Set the API key if it was returned
	if workspace.APIKey != "" {
		d.Set("api_key", workspace.APIKey)
	}

	return resourceWorkspaceRead(ctx, d, meta)
}

func resourceWorkspaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*client.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("invalid workspace ID: %s", d.Id())
	}

	workspace, err := apiClient.GetWorkspace(ctx, id)
	if err != nil {
		// Check if workspace was not found
		if IsNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	// Check if workspace is nil (API returned successfully but with nil data)
	if workspace == nil {
		d.SetId("")
		return nil
	}

	d.Set("name", workspace.Name)
	d.Set("organization_id", workspace.OrganizationID)

	// Set time field only if it's not a zero value
	if !workspace.CreatedAt.IsZero() {
		d.Set("created_at", workspace.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
	}

	if err := d.Set("notification_emails", workspace.NotificationEmails); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceWorkspaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*client.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("invalid workspace ID: %s", d.Id())
	}

	name := d.Get("name").(string)
	notificationEmails := expandStringSet(d.Get("notification_emails").(*schema.Set))

	req := &client.UpdateWorkspaceRequest{
		Name:               name,
		NotificationEmails: notificationEmails,
	}

	_, err = apiClient.UpdateWorkspace(ctx, id, req)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceWorkspaceRead(ctx, d, meta)
}

func resourceWorkspaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*client.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("invalid workspace ID: %s", d.Id())
	}

	err = apiClient.DeleteWorkspace(ctx, id)
	if err != nil {
		// If workspace is already deleted, don't return error
		if IsNotFoundError(err) {
			return nil
		}
		return diag.FromErr(err)
	}

	return nil
}

// expandStringSet converts a Terraform Set to a slice of strings
func expandStringSet(set *schema.Set) []string {
	list := set.List()
	result := make([]string, len(list))
	for i, v := range list {
		result[i] = v.(string)
	}
	return result
}
