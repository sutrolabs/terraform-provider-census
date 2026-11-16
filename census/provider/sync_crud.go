package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/sutrolabs/terraform-provider-census/census/client"
)

func resourceSyncCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*clients).V1

	workspaceId := d.Get("workspace_id").(string)

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

	destinationAttributes := ExpandDestinationAttributes(ctx, d.Get("destination_attributes").([]interface{}))
	fieldMappings := ExpandFieldMappings(ctx, d.Get("field_mapping").([]interface{}))

	// Get operation from top-level field (per OpenAPI spec)
	operation := d.Get("operation").(string)

	// Convert FieldMappings to MappingAttributes for API compliance
	mappings := ConvertFieldMappingsToMappingAttributes(fieldMappings)

	// Handle run_mode
	var mode *client.SyncMode
	runModeRaw := d.Get("run_mode").([]interface{})

	if len(runModeRaw) > 0 {
		mode = ExpandRunMode(ctx, runModeRaw)
	}

	req := &client.CreateSyncRequest{
		// Required fields per OpenAPI spec
		Operation:             operation,
		SourceAttributes:      ExpandSourceAttributes(ctx, d.Get("source_attributes").([]interface{})),
		DestinationAttributes: destinationAttributes,
		Mappings:              mappings,

		// Optional fields
		Label: d.Get("label").(string),

		// Mode - live vs triggered with trigger configurations
		Mode: mode,

		Paused: d.Get("paused").(bool),

		// Field configuration
		FieldBehavior:      d.Get("field_behavior").(string),
		FieldNormalization: d.Get("field_normalization").(string),
		FieldOrder:         d.Get("field_order").(string),

		// Sync behavior family
		SyncBehaviorFamily: d.Get("sync_behavior_family").(string),

		// Advanced configuration
		AdvancedConfiguration: ExpandAdvancedConfiguration(ctx, d.Get("advanced_configuration").(string)),

		// High water mark attribute
		HighWaterMarkAttribute: d.Get("high_water_mark_attribute").(string),

		// Historical sync operation
		HistoricalSyncOperation: d.Get("historical_sync_operation").(string),

		// Mirror strategy
		MirrorStrategy: d.Get("mirror_strategy").(string),

		// Alert configuration
		AlertAttributes: ExpandAlerts(ctx, d.Get("alert").([]interface{})),
	}

	if reqJSON, jsonErr := json.Marshal(req); jsonErr == nil {
		tflog.Debug(ctx, "Creating sync", map[string]interface{}{"request": string(reqJSON)})
	} else {
		tflog.Debug(ctx, "Creating sync (failed to marshal request for logging)", map[string]interface{}{"error": jsonErr.Error()})
	}
	sync, err := apiClient.CreateSyncWithToken(ctx, req, workspaceToken)
	if err != nil {
		return diag.FromErr(err)
	}

	tflog.Debug(ctx, "Sync created successfully", map[string]interface{}{"sync_id": sync.ID})
	d.SetId(strconv.Itoa(sync.ID))

	// Explicitly set workspace_id from our input since API doesn't return it
	d.Set("workspace_id", workspaceId)

	return resourceSyncRead(ctx, d, meta)
}

func resourceSyncRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*clients).V1

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("invalid sync ID: %s", d.Id())
	}

	tflog.Debug(ctx, "Starting resourceSyncRead", map[string]interface{}{"sync_id": id})

	// Get workspace token dynamically if we have workspace_id
	workspaceId := d.Get("workspace_id").(string)

	var sync *client.Sync
	if workspaceId != "" {
		workspaceIdInt, convErr := strconv.Atoi(workspaceId)
		if convErr != nil {
			return diag.Errorf("invalid workspace ID: %s", workspaceId)
		}

		workspaceToken, tokenErr := apiClient.GetWorkspaceAPIKey(ctx, workspaceIdInt)
		if tokenErr != nil {
			tflog.Debug(ctx, "Failed to get workspace API key", map[string]interface{}{"workspace_id": workspaceIdInt, "error": tokenErr.Error()})
			return diag.FromErr(tokenErr)
		}

		sync, err = apiClient.GetSyncWithToken(ctx, id, workspaceToken)
	} else {
		return diag.Errorf(`workspace_id is required but missing from resource state.

To fix this, add the missing workspace_id to terraform state:
  terraform state rm census_sync.example
  terraform import census_sync.example workspace_id:sync_id`)
	}

	if err != nil {
		// Check if sync was not found
		if IsNotFoundError(err) {
			tflog.Debug(ctx, "Sync not found, clearing resource ID", map[string]interface{}{"sync_id": id})
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	// Only update workspace_id if API returned it, otherwise preserve what's in state
	if sync.WorkspaceID != "" {
		d.Set("workspace_id", sync.WorkspaceID)
	}

	d.Set("label", sync.Label)
	d.Set("status", sync.Status)
	d.Set("paused", sync.Paused)

	// Set operation field from API response
	if sync.Operation != "" {
		d.Set("operation", sync.Operation)
	}

	// Set field configuration only if explicitly configured by user
	// This prevents drift when API returns defaults but user hasn't specified them
	if _, ok := d.GetOk("field_behavior"); ok && sync.FieldBehavior != "" {
		d.Set("field_behavior", sync.FieldBehavior)
	}
	if _, ok := d.GetOk("field_normalization"); ok && sync.FieldNormalization != "" {
		d.Set("field_normalization", sync.FieldNormalization)
	}
	if _, ok := d.GetOk("field_order"); ok && sync.FieldOrder != "" {
		d.Set("field_order", sync.FieldOrder)
	}
	if _, ok := d.GetOk("sync_behavior_family"); ok && sync.SyncBehaviorFamily != "" {
		d.Set("sync_behavior_family", sync.SyncBehaviorFamily)
	}

	// Set advanced configuration if present
	if sync.AdvancedConfiguration != nil && len(sync.AdvancedConfiguration) > 0 {
		if err := d.Set("advanced_configuration", FlattenAdvancedConfiguration(ctx, sync.AdvancedConfiguration)); err != nil {
			return diag.Errorf("failed to set advanced_configuration: %v", err)
		}
	}

	// Set high water mark attribute if present
	if sync.HighWaterMarkAttribute != "" {
		d.Set("high_water_mark_attribute", sync.HighWaterMarkAttribute)
	}

	// Set historical sync operation if present
	if sync.HistoricalSyncOperation != "" {
		d.Set("historical_sync_operation", sync.HistoricalSyncOperation)
	}

	// Set mirror strategy if present
	if sync.MirrorStrategy != "" {
		d.Set("mirror_strategy", sync.MirrorStrategy)
	}

	// Set alert attributes if present
	if len(sync.AlertAttributes) > 0 {
		if err := d.Set("alert", FlattenAlerts(sync.AlertAttributes)); err != nil {
			return diag.Errorf("failed to set alert: %v", err)
		}
	}

	// Set time fields only if they are not zero values
	if !sync.CreatedAt.IsZero() {
		d.Set("created_at", sync.CreatedAt.Format("2006-01-02T15:04:05Z"))
	}
	if !sync.UpdatedAt.IsZero() {
		d.Set("updated_at", sync.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	}

	// last_run_at, next_run_at and last_run_id are deprecated: the API does not support them,
	// so they are never set and stay empty.

	// Handle run_mode from API response
	if sync.Mode != nil {
		if err := d.Set("run_mode", FlattenRunMode(sync.Mode)); err != nil {
			return diag.Errorf("failed to set run_mode: %v", err)
		}
	} else if sync.ScheduleFrequency != "" {
		// Handle very old syncs that pre-date Mode API field (created before Census added Mode support)
		tflog.Debug(ctx, "Legacy sync detected - has flat schedule fields but no Mode", map[string]interface{}{"sync_id": id})
		return diag.Errorf("This sync was created with an older version of the Census API that pre-dates run_mode support. Please recreate it using run_mode configuration or contact Census support to migrate it.")
	}

	// Set complex attributes with nil checks
	if err := d.Set("source_attributes", FlattenSourceAttributes(sync.SourceAttributes)); err != nil {
		return diag.Errorf("failed to set source_attributes: %v", err)
	}

	if err := d.Set("destination_attributes", FlattenDestinationAttributes(sync.DestinationAttributes)); err != nil {
		return diag.Errorf("failed to set destination_attributes: %v", err)
	}

	// Convert API Mappings back to Terraform FieldMappings with defensive handling
	var fieldMappings []client.FieldMapping
	if sync.Mappings != nil && len(sync.Mappings) > 0 {
		fieldMappings = ConvertMappingAttributesToFieldMappings(sync.Mappings)
	} else {
		fieldMappings = []client.FieldMapping{} // Empty slice as fallback
	}
	tflog.Debug(ctx, "Converted field mappings", map[string]interface{}{"sync_id": id, "count": len(fieldMappings)})

	// Sort API mappings by position (Census's canonical order)
	// Position is always set by Census API regardless of field_order setting
	// This prevents spurious diffs when the Census API returns mappings in a different order
	if len(fieldMappings) > 0 {
		sort.Slice(fieldMappings, func(i, j int) bool {
			return fieldMappings[i].Position < fieldMappings[j].Position
		})
	}

	// Filter out Census-managed compound_key mappings
	// These are auto-generated by Census for certain destinations (e.g., Google Ads with compound keys)
	// Users cannot create or modify these mappings, so they should not be tracked in Terraform state
	userManagedMappings := make([]client.FieldMapping, 0, len(fieldMappings))
	for _, mapping := range fieldMappings {
		if mapping.Type != "compound_key" {
			userManagedMappings = append(userManagedMappings, mapping)
		} else {
			tflog.Debug(ctx, "Filtering out Census-managed compound_key mapping", map[string]interface{}{"from": mapping.From, "to": mapping.To})
		}
	}
	fieldMappings = userManagedMappings

	if err := d.Set("field_mapping", FlattenFieldMappings(fieldMappings)); err != nil {
		return diag.Errorf("failed to set field_mapping: %v", err)
	}

	if sync.SyncKey != nil {
		if err := d.Set("sync_key", sync.SyncKey); err != nil {
			return diag.Errorf("failed to set sync_key: %v", err)
		}
	}

	return nil
}

func resourceSyncUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*clients).V1

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("invalid sync ID: %s", d.Id())
	}
	tflog.Debug(ctx, "Updating sync", map[string]interface{}{"sync_id": id})

	// Safe type assertion for workspace_id
	workspaceIdInterface := d.Get("workspace_id")
	workspaceId, ok := workspaceIdInterface.(string)
	if !ok {
		return diag.Errorf("workspace_id is not a valid string: %v", workspaceIdInterface)
	}

	workspaceIdInt, err := strconv.Atoi(workspaceId)
	if err != nil {
		return diag.Errorf("invalid workspace ID: %s", workspaceId)
	}

	workspaceToken, err := apiClient.GetWorkspaceAPIKey(ctx, workspaceIdInt)
	if err != nil {
		return diag.FromErr(err)
	}

	// Handle run_mode
	var mode *client.SyncMode
	runModeRaw := d.Get("run_mode").([]interface{})

	if len(runModeRaw) > 0 {
		mode = ExpandRunMode(ctx, runModeRaw)
		tflog.Debug(ctx, "Using run_mode", map[string]interface{}{"mode": fmt.Sprintf("%+v", mode)})
	}

	// Safe type assertions for all fields
	labelInterface := d.Get("label")
	label, ok := labelInterface.(string)
	if !ok {
		return diag.Errorf("label is not a valid string: %v", labelInterface)
	}

	sourceAttrsInterface := d.Get("source_attributes")
	var sourceAttrs map[string]interface{}

	// Handle both map and list formats
	switch v := sourceAttrsInterface.(type) {
	case map[string]interface{}:
		sourceAttrs = v
	case []interface{}:
		if len(v) > 0 {
			if m, ok := v[0].(map[string]interface{}); ok {
				sourceAttrs = m
			} else {
				return diag.Errorf("source_attributes list element is not a valid map: %v", v[0])
			}
		} else {
			sourceAttrs = make(map[string]interface{})
		}
	default:
		return diag.Errorf("source_attributes is not a valid map or list: %v", sourceAttrsInterface)
	}

	// Additional processing to extract object field from list format
	if sourceAttrs != nil {
		if objData, exists := sourceAttrs["object"]; exists {
			switch v := objData.(type) {
			case []interface{}:
				// Object stored as list in Terraform state - extract first element
				if len(v) > 0 {
					if obj, ok := v[0].(map[string]interface{}); ok {
						sourceAttrs["object"] = obj
					} else {
						tflog.Debug(ctx, "source_attributes object list element is not a map", map[string]interface{}{"type": fmt.Sprintf("%T", v[0])})
					}
				}
			case map[string]interface{}:
				// Object is already a direct map - no change needed
			default:
				tflog.Debug(ctx, "source_attributes object is an unexpected type", map[string]interface{}{"type": fmt.Sprintf("%T", v)})
			}
		}
	}

	destAttrsInterface := d.Get("destination_attributes")
	var destAttrs map[string]interface{}

	// Handle both map and list formats
	switch v := destAttrsInterface.(type) {
	case map[string]interface{}:
		destAttrs = v
	case []interface{}:
		if len(v) > 0 {
			if m, ok := v[0].(map[string]interface{}); ok {
				destAttrs = m
			} else {
				return diag.Errorf("destination_attributes list element is not a valid map: %v", v[0])
			}
		} else {
			destAttrs = make(map[string]interface{})
		}
	default:
		return diag.Errorf("destination_attributes is not a valid map or list: %v", destAttrsInterface)
	}

	fieldMappingsInterface := d.Get("field_mapping")
	fieldMappings, ok := fieldMappingsInterface.([]interface{})
	if !ok {
		return diag.Errorf("field_mapping is not a valid list: %v", fieldMappingsInterface)
	}

	pausedInterface := d.Get("paused")
	paused, ok := pausedInterface.(bool)
	if !ok {
		return diag.Errorf("paused is not a valid boolean: %v", pausedInterface)
	}

	// Safe type assertions for field configuration
	fieldBehaviorInterface := d.Get("field_behavior")
	fieldBehavior, _ := fieldBehaviorInterface.(string)

	fieldNormalizationInterface := d.Get("field_normalization")
	fieldNormalization, _ := fieldNormalizationInterface.(string)

	fieldOrderInterface := d.Get("field_order")
	fieldOrder, _ := fieldOrderInterface.(string)

	syncBehaviorFamilyInterface := d.Get("sync_behavior_family")
	syncBehaviorFamily, _ := syncBehaviorFamilyInterface.(string)

	req := &client.UpdateSyncRequest{
		Label:                 label,
		Operation:             d.Get("operation").(string),
		SourceAttributes:      ExpandSourceAttributes(ctx, d.Get("source_attributes").([]interface{})),
		DestinationAttributes: ExpandStringMap(destAttrs),
		Mappings:              ConvertFieldMappingsToMappingAttributes(ExpandFieldMappings(ctx, fieldMappings)),
		Paused:                paused,

		// Mode - live vs triggered with trigger configurations
		Mode: mode,

		// Field configuration
		FieldBehavior:      fieldBehavior,
		FieldNormalization: fieldNormalization,
		FieldOrder:         fieldOrder,

		// Sync behavior family
		SyncBehaviorFamily: syncBehaviorFamily,

		// Advanced configuration
		AdvancedConfiguration: ExpandAdvancedConfiguration(ctx, d.Get("advanced_configuration").(string)),

		// High water mark attribute
		HighWaterMarkAttribute: d.Get("high_water_mark_attribute").(string),

		// Historical sync operation
		HistoricalSyncOperation: d.Get("historical_sync_operation").(string),

		// Mirror strategy
		MirrorStrategy: d.Get("mirror_strategy").(string),

		// Alert configuration
		AlertAttributes: ExpandAlerts(ctx, d.Get("alert").([]interface{})),
	}

	if reqJSON, jsonErr := json.Marshal(req); jsonErr == nil {
		tflog.Debug(ctx, "Updating sync", map[string]interface{}{"sync_id": id, "request": string(reqJSON)})
	} else {
		tflog.Debug(ctx, "Updating sync (failed to marshal request for logging)", map[string]interface{}{"sync_id": id, "error": jsonErr.Error()})
	}
	_, err = apiClient.UpdateSyncWithToken(ctx, id, req, workspaceToken)
	if err != nil {
		return diag.FromErr(err)
	}

	result := resourceSyncRead(ctx, d, meta)
	return result
}

func resourceSyncDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*clients).V1

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("invalid sync ID: %s", d.Id())
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

	err = apiClient.DeleteSyncWithToken(ctx, id, workspaceToken)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceSyncImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// Support composite format: workspace_id:sync_id
	parts := strings.Split(d.Id(), ":")

	if len(parts) == 2 {
		// Format: workspace_id:sync_id
		workspaceId := parts[0]
		syncId := parts[1]

		d.SetId(syncId)
		d.Set("workspace_id", workspaceId)

		return []*schema.ResourceData{d}, nil
	} else if len(parts) == 1 {
		// Legacy format - provide helpful error
		return nil, fmt.Errorf(`import requires workspace_id. Use format: workspace_id:sync_id

Example:
  terraform import census_sync.contact_sync 69962:123

Where 69962 is the workspace_id and 123 is the sync_id.`)
	}

	return nil, fmt.Errorf("invalid import format. Use: workspace_id:sync_id")
}
