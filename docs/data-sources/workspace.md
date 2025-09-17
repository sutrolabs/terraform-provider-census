# census_workspace Data Source

Retrieves information about a Census workspace.

## Example Usage

```hcl
data "census_workspace" "example" {
  id = "123"
}

output "workspace_name" {
  value = data.census_workspace.example.name
}
```

## Argument Reference

* `id` - (Required) The ID of the workspace to retrieve.

## Attribute Reference

The following attributes are exported:

* `id` - The ID of the workspace.
* `name` - The name of the workspace.
* `organization_id` - The ID of the organization the workspace belongs to.
* `created_at` - The timestamp when the workspace was created.
* `notification_emails` - The list of email addresses that will receive alerts from the workspace.

## Notes

- This data source requires either a personal access token or workspace access token in the provider configuration.
- When using a workspace access token, you can only retrieve information about the workspace associated with that token.