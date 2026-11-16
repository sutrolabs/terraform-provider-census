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
```

## Argument Reference

* `id` - (Required) The ID of the source.

## Attribute Reference

* `name` - The name of the source.
* `type` - The type of data source connector (e.g., "snowflake", "big_query", "postgres").
* `group_id` - The group ID of the source, used to reference it from activations and datasets.
* `created_at` - When the source was created.
* `status` - The current status of the source. **Deprecated:** This field is unsupported and will be removed in a future version.
* `test_status` - The test status of the source connection. **Deprecated:** This field is unsupported and will be removed in a future version.
* `last_tested` - Timestamp when the source was last tested. **Deprecated:** This field is unsupported and will be removed in a future version.
* `updated_at` - When the source was last updated. **Deprecated:** This field is unsupported and will be removed in a future version.
