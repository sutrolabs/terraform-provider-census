# census_source Data Source

Retrieves information about an existing Census data source connection.

## Example Usage

```hcl
data "census_source" "warehouse" {
  id = "67890"
}

output "source_group_id" {
  value = data.census_source.warehouse.group_id
}

output "source_bucket_type" {
  value     = data.census_source.warehouse.storage_backend["type"]
  sensitive = true
}
```

## Argument Reference

* `id` - (Required) The ID of the source.
* `workspace_id` - (Optional, **Deprecated**) Ignored. Sources are looked up with the personal access token alone; remove it from your configuration.

## Attribute Reference

* `name` - The name of the source.
* `type` - The type of data source connector (e.g., "snowflake", "big_query", "postgres").
* `group_id` - The group ID of the source, used to reference it from activations and datasets.
* `region` - The region the source runs in, e.g. `AWS-US-EAST-1`.
* `notification_emails` - Email addresses that receive sync alerts for this source.
* `storage_backend` - (Sensitive) The source's custom storage bucket (bring your own bucket), if one is configured: `type` (`s3`, `gcs`, or `azure`) plus that backend's fields. Empty when the source uses Census-managed storage.
* `created_at` - When the source was created.
* `status` - The current status of the source. **Deprecated:** This field is unsupported and will be removed in a future version.
* `test_status` - The test status of the source connection. **Deprecated:** This field is unsupported and will be removed in a future version.
* `last_tested` - Timestamp when the source was last tested. **Deprecated:** This field is unsupported and will be removed in a future version.
* `updated_at` - When the source was last updated. **Deprecated:** This field is unsupported and will be removed in a future version.
