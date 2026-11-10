package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/sutrolabs/terraform-provider-census/census/client/v2"
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

		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceSourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceSourceStateUpgradeV0,
				Version: 0,
			},
		},

		CustomizeDiff: func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
			// Force diff detection for connection changes so updated_at is
			// recomputed, including write-only credential rotations signalled by a
			// connection_config_wo_version bump.
			if sourceConnectionChanged(d) {
				d.SetNewComputed("updated_at")
			}
			return nil
		},

		Schema: sourceSchemaMap(),
	}
}

func sourceSchemaMap() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The ID of the source.",
		},
		"group_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The group ID of the source, used to reference it from activations and datasets.",
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
			Deprecated:  "This field is unsupported and will be removed in a future version.",
		},
		"test_status": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The status of the last connection test.",
			Deprecated:  "This field is unsupported and will be removed in a future version.",
		},
		"last_tested": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Timestamp when the source was last tested.",
			Deprecated:  "This field is unsupported and will be removed in a future version.",
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
			Deprecated:  "This field is unsupported and will be removed in a future version.",
		},
		"auto_refresh_tables": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Whether to automatically refresh table metadata after creation.",
		},
		"region": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "The region the source runs in (e.g. `AWS-US-EAST-1`). Region availability depends on your account tier. Leave unset to use the account default.",
		},
		"notification_emails": {
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "Email addresses notified about this source.",
		},
		"storage_backend": {
			Type:        schema.TypeMap,
			Optional:    true,
			Sensitive:   true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "Custom storage backend for the source. Set `type` to `s3`, `gcs`, or `azure` plus that backend's fields (S3: `bucket_name`, optionally `role_arn`/`external_id`; GCS: `bucket_name`, optionally `project_id`/`service_account_email`/`service_account_key`; Azure: `storage_account_name`, `storage_container_name`, optionally `sas_token`). Like `connection_config`, it is not read back from the API.",
		},
	}
}

// sourceAttributesAddedInV1 are the attributes the v2 API cutover added to
// census_source. Every other attribute kept its type, so v0 is the current
// schema without these.
var sourceAttributesAddedInV1 = []string{"group_id", "region", "notification_emails", "storage_backend"}

func resourceSourceV0() *schema.Resource {
	s := sourceSchemaMap()
	for _, k := range sourceAttributesAddedInV1 {
		delete(s, k)
	}
	return &schema.Resource{Schema: s}
}

// resourceSourceStateUpgradeV0 fills in notification_emails for state written
// before it existed. It is Optional + Computed, and SDKv2 plans an absent
// Computed list as "known after apply" even when nothing has changed, so
// without this an upgraded source shows a spurious update until its first
// refresh. An empty list is safe: the next refresh replaces it with whatever
// Census reports.
func resourceSourceStateUpgradeV0(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	if rawState == nil {
		rawState = map[string]interface{}{}
	}
	if rawState["notification_emails"] == nil {
		rawState["notification_emails"] = []interface{}{}
	}
	return rawState, nil
}

func resourceSourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*clients).V2

	sourceType := d.Get("type").(string)
	connectionConfig := expandConnectionConfig(d.Get("connection_config").(map[string]interface{}))
	if err := applyWriteOnlyCredentials(d, connectionConfig); err != nil {
		return diag.FromErr(err)
	}

	// Validate source credentials against source type requirements
	if err := v2.ValidateSourceCredentials(ctx, apiClient, sourceType, connectionConfig); err != nil {
		return diag.Errorf("source credential validation failed: %v", err)
	}

	connection := expandSourceConnection(d, connectionConfig)
	connection.Type = sourceType
	connection.SyncEngine = getConfiguredSourceSyncEngine(d)

	source, err := v2.CreateSource(ctx, apiClient, &v2.CreateSourceRequest{Connection: connection})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(source.ID))

	// Optionally refresh tables after creation
	if d.Get("auto_refresh_tables").(bool) {
		if err := v2.RefreshSourceTables(ctx, apiClient, source.ID); err != nil {
			// The source was created successfully, table refresh is optional
			return diag.Errorf("source created successfully but table refresh failed: %v", err)
		}
	}

	return resourceSourceRead(ctx, d, meta)
}

func resourceSourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*clients).V2

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("invalid source ID: %s", d.Id())
	}

	source, err := v2.GetSource(ctx, apiClient, id)
	if err != nil {
		if v2.IsNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	if source.SyncEngine != "" {
		d.Set("sync_engine", source.SyncEngine)
	}
	if source.WarehouseWritebackRetentionInDays != nil {
		d.Set("warehouse_writeback_retention_in_days", *source.WarehouseWritebackRetentionInDays)
	}
	d.Set("name", source.Name)
	d.Set("type", source.Type)
	d.Set("group_id", source.GroupID)
	d.Set("region", source.Region)
	d.Set("notification_emails", source.NotificationEmails)
	d.Set("created_at", source.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))

	// status, test_status, last_tested and updated_at are deprecated: the API does not support them.
	// They are set to the same values they have always had (empty, and the zero timestamp for updated_at).
	d.Set("status", "")
	d.Set("test_status", "")
	d.Set("updated_at", "0001-01-01T00:00:00Z")

	// Note: connection_config and storage_backend are not set from the API
	// response, as it may not contain all the original values (especially
	// secrets). Terraform maintains them from the configuration.

	return nil
}

func resourceSourceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*clients).V2

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("invalid source ID: %s", d.Id())
	}

	sourceType := d.Get("type").(string)
	connectionConfig := expandConnectionConfig(d.Get("connection_config").(map[string]interface{}))
	if err := applyWriteOnlyCredentials(d, connectionConfig); err != nil {
		return diag.FromErr(err)
	}

	// Validate when the connection changes, including when write-only
	// credentials are rotated via connection_config_wo_version.
	if sourceConnectionChanged(d) {
		if err := v2.ValidateSourceCredentials(ctx, apiClient, sourceType, connectionConfig); err != nil {
			return diag.Errorf("source credential validation failed: %v", err)
		}
	}

	// Type and sync_engine cannot be modified after creation.
	connection := expandSourceConnection(d, connectionConfig)

	if _, err := v2.UpdateSource(ctx, apiClient, id, &v2.UpdateSourceRequest{Connection: connection}); err != nil {
		return diag.FromErr(err)
	}

	// Refresh tables if requested and connection changed, including when
	// write-only credentials are rotated via connection_config_wo_version.
	if sourceConnectionChanged(d) && d.Get("auto_refresh_tables").(bool) {
		if err := v2.RefreshSourceTables(ctx, apiClient, id); err != nil {
			return diag.Errorf("source updated successfully but table refresh failed: %v", err)
		}
	}

	return resourceSourceRead(ctx, d, meta)
}

func resourceSourceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*clients).V2

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("invalid source ID: %s", d.Id())
	}

	if err := v2.DeleteSource(ctx, apiClient, id); err != nil {
		// If source is already deleted, don't return error
		if v2.IsNotFoundError(err) {
			return nil
		}
		return diag.FromErr(err)
	}

	return nil
}

// expandSourceConnection builds the attributes shared by create and update.
func expandSourceConnection(d *schema.ResourceData, credentials map[string]interface{}) v2.SourceConnection {
	connection := v2.SourceConnection{
		Name:                              d.Get("name").(string),
		WarehouseWritebackRetentionInDays: getConfiguredWarehouseWritebackRetention(d),
		Credentials:                       credentials,
	}

	if region, ok := d.GetOk("region"); ok {
		r := region.(string)
		connection.Region = &r
	}

	if raw, ok := d.GetOk("notification_emails"); ok {
		for _, email := range raw.([]interface{}) {
			connection.NotificationEmails = append(connection.NotificationEmails, email.(string))
		}
	}

	if raw, ok := d.GetOk("storage_backend"); ok {
		connection.StorageBackend = raw.(map[string]interface{})
	}

	return connection
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
