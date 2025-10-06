# census_sync Data Source

Retrieves information about an existing Census sync.

## Example Usage

```hcl
data "census_sync" "user_sync" {
  id           = "98765"
  workspace_id = census_workspace.main.id
}

output "sync_status" {
  value = data.census_sync.user_sync.status
}

output "sync_paused" {
  value = data.census_sync.user_sync.paused
}

output "sync_run_mode" {
  value = data.census_sync.user_sync.run_mode
}
```

## Argument Reference

* `id` - (Required) The ID of the sync.
* `workspace_id` - (Required) The ID of the workspace this sync belongs to.

## Attribute Reference

* `label` - The name/label of the sync.
* `source_attributes` - List containing source configuration block:
  * `connection_id` - The source connection ID.
  * `cohort_id` - The cohort ID (if applicable).
  * `filter_segment_id` - The filter segment ID (if applicable).
  * `object` - List containing source object configuration:
    * `type` - Object type (table, dataset, model, topic, segment, cohort).
    * `table_name` - Table name (for table type).
    * `table_schema` - Table schema (for table type).
    * `table_catalog` - Table catalog (for table type).
    * `id` - Object ID (for dataset, model, etc.).
* `destination_attributes` - List containing destination configuration block:
  * `connection_id` - The destination connection ID.
  * `object` - The destination object name.
  * `lead_union_insert_to` - Lead union insertion target (Salesforce only).
* `field_mapping` - Set of field mappings between source and destination:
  * `from` - Source field name.
  * `to` - Destination field name.
  * `type` - Mapping type: "direct" (default), "hash", "constant", "sync_metadata", "segment_membership", or "liquid_template".
  * `constant` - Constant value (when type="constant").
  * `sync_metadata_key` - Metadata key (when type="sync_metadata").
  * `segment_identify_by` - Segment identifier (when type="segment_membership").
  * `liquid_template` - Liquid template (when type="liquid_template").
  * `is_primary_identifier` - Whether this field is the primary identifier.
  * `lookup_object` - Lookup object for relationship mapping.
  * `lookup_field` - Lookup field for relationship mapping.
  * `preserve_values` - Whether to preserve existing destination values.
  * `generate_field` - Whether Census should generate this field.
  * `sync_null_values` - Whether to sync null values.
  * `array_field` - Whether the field is an array type.
  * `field_type` - Field type (when generate_field=true).
  * `follow_source_type` - Whether to follow source type changes.
* `operation` - Sync mode: "upsert", "append", "mirror", "update", or "insert".
* `run_mode` - Run mode configuration block:
  * `type` - Mode type: "live" or "triggered".
  * `triggers` - Trigger configurations (for triggered mode):
    * `schedule` - Schedule-based trigger:
      * `frequency` - How often to run: "never", "continuous", "quarter_hourly", "hourly", "daily", "weekly", or "expression".
      * `day` - Day of week (for weekly schedules).
      * `hour` - Hour to run (0-24).
      * `minute` - Minute to run (0-59).
      * `cron_expression` - Cron expression (when frequency="expression").
    * `dbt_cloud` - dbt Cloud trigger configuration (if configured).
    * `fivetran` - Fivetran trigger configuration (if configured).
    * `sync_sequence` - Sync sequence trigger configuration (if configured).
* `field_behavior` - Field syncing behavior: "specific_properties" or "sync_all_properties".
* `field_normalization` - Field name normalization (when field_behavior="sync_all_properties").
* `field_order` - Field ordering: "alphabetical_column_name" or "mapping_order".
* `sync_behavior_family` - Behavior family: "activateEvents" or "mapRecords".
* `advanced_configuration` - Advanced configuration JSON string.
* `high_water_mark_attribute` - High water mark column name.
* `historical_sync_operation` - Historical sync operation: "skip_current_records" or "backfill_all_records".
* `mirror_strategy` - Mirror strategy: "sync_updates_and_deletes", "sync_updates_and_nulls", or "upload_and_swap".
* `alert` - Set of alert configurations (if configured).
* `paused` - Whether the sync is currently paused.
* `status` - The current status of the sync.
* `created_at` - When the sync was created.
* `updated_at` - When the sync was last updated.
* `last_run_at` - When the sync last ran.
* `next_run_at` - When the sync will run next.
* `last_run_id` - ID of the last sync run.