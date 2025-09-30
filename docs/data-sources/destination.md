# census_destination Data Source

Retrieves information about an existing Census destination connection.

## Example Usage

```hcl
data "census_destination" "salesforce" {
  id           = "12345"
  workspace_id = census_workspace.main.id
}

output "destination_status" {
  value = data.census_destination.salesforce.connection_status
}
```

## Argument Reference

* `id` - (Required) The ID of the destination.
* `workspace_id` - (Required) The ID of the workspace this destination belongs to.

## Attribute Reference

* `name` - The name of the destination.
* `type` - The type of destination connector (e.g., "salesforce", "hubspot", "intercom").
* `connection_status` - The current connection status of the destination.
* `credentials` - The credentials for the destination connection (sensitive, not fully populated in reads).