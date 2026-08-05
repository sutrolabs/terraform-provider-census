package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/sutrolabs/terraform-provider-census/census/client"
)

func resourceSource() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Census data source connection.",

		CreateContext: resourceSourceCreate,
		ReadContext:   resourceSourceRead,
		UpdateContext: resourceSourceUpdate,
		DeleteContext: resourceSourceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceSourceImport,
		},

		CustomizeDiff: func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
			// Force diff detection for sensitive connection_config changes.
			// Write-only credential values are never stored in state, so a
			// change to them cannot be detected directly; users signal a
			// rotation by bumping connection_config_wo_version, which we treat
			// as a connection change here.
			if d.HasChange("connection_config") || d.HasChange("connection_config_wo_version") {
				d.SetNewComputed("updated_at")
			}
			return nil
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
			"sync_engine": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "The sync engine to use for the source. New sources default to `advanced` to match the Census UI. Leave this unset to preserve the current engine on existing sources, or set it explicitly to `basic` or `advanced` to manage it in Terraform.",
			},
			"warehouse_writeback_retention_in_days": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Enables Warehouse Writeback for the source and sets the sync log retention period (in days). Only supported on the advanced sync engine and on source types that support sync logs (Snowflake, BigQuery, Databricks, Redshift). Setting this enables Warehouse Writeback. Once enabled, the Census API does not currently support disabling it via this attribute; remove the attribute in Terraform to stop managing the value, in which case Terraform will preserve whatever value the API reports.",
			},
			"connection_config": {
				Type:        schema.TypeMap,
				Required:    true,
				Sensitive:   true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Connection configuration for the source. Contents vary by source type. Secret values may instead be supplied through the write-only `connection_config_wo` argument so they are never persisted in Terraform state or plan files.",
			},
			"connection_config_wo": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				WriteOnly:    true,
				RequiredWith: []string{"connection_config_wo_version"},
				Description:  "Write-only, `jsonencode`-d object of secret connection credentials (e.g. `password`, Snowflake `private_key_pkcs8` / `private_key_passphrase`, BigQuery `service_account_key`, etc.). Works for any source type without enumerating fields. Its value is never stored in Terraform state or plan. Each key/value is merged into `connection_config` when the source is created or updated, taking precedence over the same key set in `connection_config`. Requires Terraform 1.11+. Bump `connection_config_wo_version` to have Terraform apply new values.",
			},
			"connection_config_wo_version": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Trigger for updating the write-only `connection_config_wo` argument. Because write-only values are never stored in state, Terraform cannot detect when they change. Increment this value to signal that the current write-only credentials should be re-applied on the next apply.",
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
	syncEngine := getConfiguredSourceSyncEngine(d)
	connectionConfig := expandConnectionConfig(d.Get("connection_config").(map[string]interface{}))
	if err := applyWriteOnlyCredentials(d, connectionConfig); err != nil {
		return diag.FromErr(err)
	}

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
			Name:                              name, // Set name inside connection per API requirements
			Type:                              sourceType,
			SyncEngine:                        syncEngine,
			WarehouseWritebackRetentionInDays: getConfiguredWarehouseWritebackRetention(d),
			Credentials:                       connectionConfig,
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
		workspaceIdInt, convErr := strconv.Atoi(workspaceId)
		if convErr != nil {
			return diag.Errorf("invalid workspace ID: %s", workspaceId)
		}

		workspaceToken, tokenErr := apiClient.GetWorkspaceAPIKey(ctx, workspaceIdInt)
		if tokenErr != nil {
			return diag.FromErr(tokenErr)
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

	// Check if source is nil (API returned successfully but with nil data)
	if source == nil {
		d.SetId("")
		return nil
	}

	// Only update workspace_id if API returned it, otherwise preserve what's in state
	if source.WorkspaceID != "" {
		d.Set("workspace_id", source.WorkspaceID)
	}
	if syncEngine := getSourceSyncEngine(source); syncEngine != "" {
		d.Set("sync_engine", syncEngine)
	}
	if source.WarehouseWritebackRetentionInDays != nil {
		d.Set("warehouse_writeback_retention_in_days", *source.WarehouseWritebackRetentionInDays)
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

func getSourceSyncEngine(source *client.Source) string {
	if source == nil {
		return ""
	}

	if source.SyncEngine != "" {
		return source.SyncEngine
	}

	if source.Connection == nil {
		return ""
	}

	syncEngine, ok := source.Connection["sync_engine"].(string)
	if !ok {
		return ""
	}

	return syncEngine
}

func getConfiguredSourceSyncEngine(d *schema.ResourceData) string {
	if syncEngine, ok := d.GetOk("sync_engine"); ok {
		if configuredSyncEngine := syncEngine.(string); configuredSyncEngine != "" {
			return configuredSyncEngine
		}
	}

	return "advanced"
}

func getConfiguredWarehouseWritebackRetention(d *schema.ResourceData) *int {
	raw, ok := d.GetOk("warehouse_writeback_retention_in_days")
	if !ok {
		return nil
	}

	retention, ok := raw.(int)
	if !ok || retention <= 0 {
		return nil
	}

	return &retention
}

func resourceSourceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*client.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("invalid source ID: %s", d.Id())
	}

	// Get current values for structured update
	name := d.Get("name").(string)
	sourceType := d.Get("type").(string)
	connectionConfig := expandConnectionConfig(d.Get("connection_config").(map[string]interface{}))
	if err := applyWriteOnlyCredentials(d, connectionConfig); err != nil {
		return diag.FromErr(err)
	}
	workspaceId := d.Get("workspace_id").(string)

	workspaceIdInt, err := strconv.Atoi(workspaceId)
	if err != nil {
		return diag.Errorf("invalid workspace ID: %s", workspaceId)
	}

	workspaceToken, err := apiClient.GetWorkspaceAPIKey(ctx, workspaceIdInt)
	if err != nil {
		return diag.FromErr(err)
	}

	// Always build complete connection structure for updates. Validate when the
	// connection changes, including when write-only credentials are rotated via
	// connection_config_wo_version.
	if d.HasChange("connection_config") || d.HasChange("connection_config_wo_version") {
		if err := apiClient.ValidateSourceCredentials(ctx, sourceType, connectionConfig, workspaceToken); err != nil {
			return diag.Errorf("source credential validation failed: %v", err)
		}
	}

	req := &client.UpdateSourceRequest{
		Connection: client.SourceConnection{
			Name: name, // Set name in connection structure per API requirements
			// Note: Type and SyncEngine cannot be modified after creation per Census API
			WarehouseWritebackRetentionInDays: getConfiguredWarehouseWritebackRetention(d),
			Credentials:                       connectionConfig,
		},
	}

	// Use the workspace token for the update
	_, err = apiClient.UpdateSourceWithToken(ctx, id, req, workspaceToken)
	if err != nil {
		return diag.FromErr(err)
	}

	// Refresh tables if requested and connection changed, including when
	// write-only credentials are rotated via connection_config_wo_version.
	if (d.HasChange("connection_config") || d.HasChange("connection_config_wo_version")) && d.Get("auto_refresh_tables").(bool) {
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

func resourceSourceImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// Support composite format: workspace_id:source_id
	parts := strings.Split(d.Id(), ":")

	if len(parts) == 2 {
		// Format: workspace_id:source_id
		workspaceId := parts[0]
		sourceId := parts[1]

		d.SetId(sourceId)
		d.Set("workspace_id", workspaceId)

		return []*schema.ResourceData{d}, nil
	} else if len(parts) == 1 {
		// Legacy format - provide helpful error
		return nil, fmt.Errorf(`import requires workspace_id. Use format: workspace_id:source_id

Example:
  terraform import census_source.snowflake_basic 69962:828

Where 69962 is the workspace_id and 828 is the source_id.`)
	}

	return nil, fmt.Errorf("invalid import format. Use: workspace_id:source_id")
}
