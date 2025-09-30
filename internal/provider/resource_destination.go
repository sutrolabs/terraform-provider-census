package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/sutrolabs/terraform-provider-census/internal/client"
)

func resourceDestination() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Census data destination connection.",

		CreateContext: resourceDestinationCreate,
		ReadContext:   resourceDestinationRead,
		UpdateContext: resourceDestinationUpdate,
		DeleteContext: resourceDestinationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
			// Force diff detection for sensitive connection_config changes
			if d.HasChange("connection_config") {
				d.SetNewComputed("updated_at")
			}
			return nil
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the destination.",
			},
			"workspace_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the workspace this destination belongs to.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the destination connection.",
			},
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The type of destination (e.g., salesforce, hubspot, postgres).",
			},
			"connection_config": {
				Type:        schema.TypeMap,
				Required:    true,
				Sensitive:   true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Connection configuration for the destination. Contents vary by destination type.",
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
			"auto_refresh_objects": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether to automatically refresh object metadata after creation.",
			},
		},
	}
}

func resourceDestinationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*client.Client)

	workspaceId := d.Get("workspace_id").(string)
	name := d.Get("name").(string)
	destinationType := d.Get("type").(string)
	connectionConfig := expandConnectionConfig(d.Get("connection_config").(map[string]interface{}))

	// Get the workspace API key dynamically using the personal access token
	workspaceIdInt, err := strconv.Atoi(workspaceId)
	if err != nil {
		return diag.Errorf("invalid workspace ID: %s", workspaceId)
	}

	workspaceToken, err := apiClient.GetWorkspaceAPIKey(ctx, workspaceIdInt)
	if err != nil {
		return diag.FromErr(err)
	}

	// Validate destination credentials against connector requirements (now with pagination support)
	if err := apiClient.ValidateDestinationCredentials(ctx, destinationType, connectionConfig, workspaceToken); err != nil {
		return diag.Errorf("destination credential validation failed: %v", err)
	}

	req := &client.CreateDestinationRequest{
		Type: destinationType,
		ServiceConnection: client.DestinationConnection{
			Name:        name, // Set name inside service_connection per API requirements
			Type:        destinationType,
			Credentials: connectionConfig,
		},
	}

	// Use the dynamically retrieved workspace token
	destination, err := apiClient.CreateDestinationWithToken(ctx, req, workspaceToken)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(destination.ID))

	// Explicitly set workspace_id from our input since API doesn't return it
	d.Set("workspace_id", workspaceId)

	// Optionally refresh objects after creation
	if d.Get("auto_refresh_objects").(bool) {
		refreshReq := &client.RefreshObjectsRequest{}
		if err := apiClient.RefreshDestinationObjectsWithToken(ctx, destination.ID, refreshReq, workspaceToken); err != nil {
			// Log the error but don't fail the creation
			// The destination was created successfully, object refresh is optional
			return diag.Errorf("destination created successfully but object refresh failed: %v", err)
		}
	}

	return resourceDestinationRead(ctx, d, meta)
}

func resourceDestinationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*client.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("invalid destination ID: %s", d.Id())
	}

	// Get workspace token dynamically if we have workspace_id
	workspaceId := d.Get("workspace_id").(string)
	var destination *client.Destination
	if workspaceId != "" {
		workspaceIdInt, err := strconv.Atoi(workspaceId)
		if err != nil {
			return diag.Errorf("invalid workspace ID: %s", workspaceId)
		}

		workspaceToken, err := apiClient.GetWorkspaceAPIKey(ctx, workspaceIdInt)
		if err != nil {
			return diag.FromErr(err)
		}

		destination, err = apiClient.GetDestinationWithToken(ctx, id, workspaceToken)
	} else {
		// In PAT-only architecture, workspace_id is required for read operations
		return diag.Errorf("workspace_id is required but missing from resource state - please reimport this resource")
	}

	if err != nil {
		// Check if destination was not found
		if IsNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	// Only update workspace_id if API returned it, otherwise preserve what's in state
	if destination.WorkspaceID != "" {
		d.Set("workspace_id", destination.WorkspaceID)
	}
	d.Set("name", destination.Name)
	d.Set("type", destination.Type)
	d.Set("status", destination.Status)
	d.Set("test_status", destination.TestStatus)
	d.Set("created_at", destination.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
	d.Set("updated_at", destination.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"))

	if destination.LastTested != nil {
		d.Set("last_tested", destination.LastTested.Format("2006-01-02T15:04:05Z07:00"))
	}

	// Note: We don't set the connection config from the API response
	// as it may not contain all the original values (especially secrets)
	// Terraform will maintain the connection config from the configuration

	return nil
}

func resourceDestinationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*client.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("invalid destination ID: %s", d.Id())
	}

	// Always get workspace token for the update operation
	workspaceId := d.Get("workspace_id").(string)
	workspaceIdInt, err := strconv.Atoi(workspaceId)
	if err != nil {
		return diag.Errorf("invalid workspace ID: %s", workspaceId)
	}

	workspaceToken, err := apiClient.GetWorkspaceAPIKey(ctx, workspaceIdInt)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &client.UpdateDestinationRequest{}

	// Census API requires service_connection for all updates, so always include it
	connectionConfig := expandConnectionConfig(d.Get("connection_config").(map[string]interface{}))
	destinationType := d.Get("type").(string)

	// If connection changed, validate the new credentials
	if d.HasChange("connection_config") {
		if err := apiClient.ValidateDestinationCredentials(ctx, destinationType, connectionConfig, workspaceToken); err != nil {
			return diag.Errorf("destination credential validation failed: %v", err)
		}
	}

	// Always build ServiceConnection structure since API requires it
	// Note: Don't include Type field as it cannot be modified after creation
	// Note: Don't include SyncEngine as destinations don't have sync engines (sources do)
	req.ServiceConnection = &client.DestinationConnection{
		Name:        d.Get("name").(string), // Set name inside service_connection per API requirements
		Credentials: connectionConfig,
	}

	_, err = apiClient.UpdateDestinationWithToken(ctx, id, req, workspaceToken)
	if err != nil {
		return diag.FromErr(err)
	}

	// Refresh objects if requested and connection changed
	if d.HasChange("connection_config") && d.Get("auto_refresh_objects").(bool) {
		// We need the workspace token for refresh
		workspaceId := d.Get("workspace_id").(string)
		workspaceIdInt, err := strconv.Atoi(workspaceId)
		if err != nil {
			return diag.Errorf("invalid workspace ID: %s", workspaceId)
		}

		workspaceToken, err := apiClient.GetWorkspaceAPIKey(ctx, workspaceIdInt)
		if err != nil {
			return diag.FromErr(err)
		}

		refreshReq := &client.RefreshObjectsRequest{}
		if err := apiClient.RefreshDestinationObjectsWithToken(ctx, id, refreshReq, workspaceToken); err != nil {
			// Log the error but don't fail the update
			return diag.Errorf("destination updated successfully but object refresh failed: %v", err)
		}
	}

	return resourceDestinationRead(ctx, d, meta)
}

func resourceDestinationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*client.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("invalid destination ID: %s", d.Id())
	}

	// Get workspace token dynamically if we have workspace_id
	workspaceId := d.Get("workspace_id").(string)
	if workspaceId != "" {
		workspaceIdInt, err := strconv.Atoi(workspaceId)
		if err != nil {
			return diag.Errorf("invalid workspace ID: %s", workspaceId)
		}

		workspaceToken, err := apiClient.GetWorkspaceAPIKey(ctx, workspaceIdInt)
		if err != nil {
			return diag.FromErr(err)
		}

		err = apiClient.DeleteDestinationWithToken(ctx, id, workspaceToken)
	} else {
		// In PAT-only architecture, workspace_id is required for delete operations
		return diag.Errorf("workspace_id is required but missing from resource state - please reimport this resource")
	}

	if err != nil {
		// If destination is already deleted, don't return error
		if IsNotFoundError(err) {
			return nil
		}
		return diag.FromErr(err)
	}

	return nil
}
