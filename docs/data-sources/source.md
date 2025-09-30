# census_source Data Source

Retrieves information about an existing Census data source connection.

## Example Usage

```hcl
data "census_source" "warehouse" {
  id           = "67890"
  workspace_id = census_workspace.main.id
}

output "source_status" {
  value = data.census_source.warehouse.connection_status
}
```

## Argument Reference

* `id` - (Required) The ID of the source.
* `workspace_id` - (Required) The ID of the workspace this source belongs to.

## Attribute Reference

* `name` - The name of the source.
* `type` - The type of data source connector (e.g., "snowflake", "big_query", "postgres").
* `connection_status` - The current connection status of the source.
* `credentials` - The credentials for the source connection (sensitive, not fully populated in reads).