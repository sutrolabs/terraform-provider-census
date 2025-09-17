package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/your-org/terraform-provider-census/internal/client"
)

func resourceSource() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Census data source connection.",

		CreateContext: resourceSourceCreate,
		ReadContext:   resourceSourceRead,
		UpdateContext: resourceSourceUpdate,
		DeleteContext: resourceSourceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the source.",
			},
			"workspace_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the workspace this source belongs to.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the source connection.",
			},
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The type of source (e.g., snowflake, bigquery, postgres).",
			},
			"connection_config": {
				Type:        schema.TypeMap,
				Required:    true,
				Sensitive:   true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Connection configuration for the source. Contents vary by source type.",
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
			"auto_refresh_tables": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether to automatically refresh table metadata after creation.",
			},
		},
	}
}

func resourceSourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*client.Client)

	workspaceId := d.Get("workspace_id").(string)
	name := d.Get("name").(string)
	sourceType := d.Get("type").(string)
	connectionConfig := expandConnectionConfig(d.Get("connection_config").(map[string]interface{}))

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

	// Validate source credentials against source type requirements
	if err := apiClient.ValidateSourceCredentials(ctx, sourceType, connectionConfig, workspaceToken); err != nil {
		return diag.Errorf("source credential validation failed: %v", err)
	}

	req := &client.CreateSourceRequest{
		Connection: client.SourceConnection{
			Label:       name,
			Type:        sourceType,
			SyncEngine:  "basic", // Default sync engine
			Credentials: connectionConfig,
		},
	}
	
	// Use the dynamically retrieved workspace token
	source, err := apiClient.CreateSourceWithToken(ctx, req, workspaceToken)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(source.ID))
	
	// Explicitly set workspace_id from our input since API doesn't return it
	d.Set("workspace_id", workspaceId)

	// Optionally refresh tables after creation
	if d.Get("auto_refresh_tables").(bool) {
		if err := apiClient.RefreshSourceTablesWithToken(ctx, source.ID, workspaceToken); err != nil {
			// Log the error but don't fail the creation
			// The source was created successfully, table refresh is optional
			return diag.Errorf("source created successfully but table refresh failed: %v", err)
		}
	}

	return resourceSourceRead(ctx, d, meta)
}

func resourceSourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*client.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("invalid source ID: %s", d.Id())
	}

	// Get workspace token dynamically if we have workspace_id
	workspaceId := d.Get("workspace_id").(string)
	var source *client.Source
	if workspaceId != "" {
		workspaceIdInt, err := strconv.Atoi(workspaceId)
		if err != nil {
			return diag.Errorf("invalid workspace ID: %s", workspaceId)
		}
		
		workspaceToken, err := apiClient.GetWorkspaceAPIKey(ctx, workspaceIdInt)
		if err != nil {
			return diag.FromErr(err)
		}
		
		source, err = apiClient.GetSourceWithToken(ctx, id, workspaceToken)
	} else {
		// workspace_id missing - this is a legacy resource that needs manual fixing
		// The simplest fix is to manually add workspace_id to terraform state
		return diag.Errorf(`workspace_id is required but missing from resource state. 

To fix this, add the missing workspace_id to terraform state:
  terraform state rm census_source.marketing_prod_warehouse
  terraform import census_source.marketing_prod_warehouse 69962:2280673

Where 69962 is the workspace_id for marketing_prod workspace.`)
	}
	
	if err != nil {
		// Check if source was not found
		if IsNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	// Only update workspace_id if API returned it, otherwise preserve what's in state
	if source.WorkspaceID != "" {
		d.Set("workspace_id", source.WorkspaceID)
	}
	d.Set("name", source.Name)
	d.Set("type", source.Type)
	d.Set("status", source.Status)
	d.Set("test_status", source.TestStatus)
	d.Set("created_at", source.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
	d.Set("updated_at", source.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"))

	if source.LastTested != nil {
		d.Set("last_tested", source.LastTested.Format("2006-01-02T15:04:05Z07:00"))
	}

	// Note: We don't set the connection config from the API response
	// as it may not contain all the original values (especially secrets)
	// Terraform will maintain the connection config from the configuration

	return nil
}

func resourceSourceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*client.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("invalid source ID: %s", d.Id())
	}

	req := &client.UpdateSourceRequest{}

	if d.HasChange("name") {
		req.Name = d.Get("name").(string)
	}

	if d.HasChange("connection_config") {
		connectionConfig := expandConnectionConfig(d.Get("connection_config").(map[string]interface{}))
		
		// Validate updated credentials if connection changed
		workspaceId := d.Get("workspace_id").(string)
		sourceType := d.Get("type").(string)
		
		workspaceIdInt, err := strconv.Atoi(workspaceId)
		if err != nil {
			return diag.Errorf("invalid workspace ID: %s", workspaceId)
		}
		
		workspaceToken, err := apiClient.GetWorkspaceAPIKey(ctx, workspaceIdInt)
		if err != nil {
			return diag.FromErr(err)
		}
		
		if err := apiClient.ValidateSourceCredentials(ctx, sourceType, connectionConfig, workspaceToken); err != nil {
			return diag.Errorf("source credential validation failed: %v", err)
		}
		
		req.Connection = connectionConfig
		
		// Use the workspace token for the update as well
		_, err = apiClient.UpdateSourceWithToken(ctx, id, req, workspaceToken)
		if err != nil {
			return diag.FromErr(err)
		}
	} else {
		// If no connection config change, still need workspace token for regular update
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
			
			_, err = apiClient.UpdateSourceWithToken(ctx, id, req, workspaceToken)
		} else {
			_, err = apiClient.UpdateSource(ctx, id, req)
		}
		if err != nil {
			return diag.FromErr(err)
		}
	}

	// Refresh tables if requested and connection changed
	if d.HasChange("connection_config") && d.Get("auto_refresh_tables").(bool) {
		// We need the workspace token for refresh, get it again
		workspaceId := d.Get("workspace_id").(string)
		workspaceIdInt, err := strconv.Atoi(workspaceId)
		if err != nil {
			return diag.Errorf("invalid workspace ID: %s", workspaceId)
		}
		
		workspaceToken, err := apiClient.GetWorkspaceAPIKey(ctx, workspaceIdInt)
		if err != nil {
			return diag.FromErr(err)
		}
		
		if err := apiClient.RefreshSourceTablesWithToken(ctx, id, workspaceToken); err != nil {
			// Log the error but don't fail the update
			return diag.Errorf("source updated successfully but table refresh failed: %v", err)
		}
	}

	return resourceSourceRead(ctx, d, meta)
}

func resourceSourceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*client.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("invalid source ID: %s", d.Id())
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
		
		err = apiClient.DeleteSourceWithToken(ctx, id, workspaceToken)
	} else {
		// In PAT-only architecture, workspace_id is required for delete operations
		return diag.Errorf("workspace_id is required but missing from resource state - please reimport this resource")
	}
	
	if err != nil {
		// If source is already deleted, don't return error
		if IsNotFoundError(err) {
			return nil
		}
		return diag.FromErr(err)
	}

	return nil
}


