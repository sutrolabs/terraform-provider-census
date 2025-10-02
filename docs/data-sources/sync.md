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
```

## Argument Reference

* `id` - (Required) The ID of the sync.
* `workspace_id` - (Required) The ID of the workspace this sync belongs to.

## Attribute Reference

* `name` - The name of the sync.
* `source_attributes` - JSON-encoded configuration for the source object.
* `destination_object` - The destination object name.
* `field_mapping` - Set of field mappings between source and destination:
  * `from` - Source field name.
  * `to` - Destination field name.
  * `operation` - Mapping operation ("direct", "hash", or "constant").
  * `constant` - Constant value (for constant operations).
* `operation` - Sync mode ("upsert", "append", or "mirror").
* `trigger` - JSON-encoded trigger configuration for scheduling.
* `paused` - Whether the sync is currently paused.
* `status` - The current status of the sync.