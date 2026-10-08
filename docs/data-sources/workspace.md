# census_workspace Data Source

~> **Deprecated:** Workspace management has been removed from the version of the Census API this provider targets. You may continue to include census_workspace data sources in your configurations, but no data will be fetched from the Census API. Do not include census_workspace data sources in new configurations.

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

- This data source requires a personal access token in the provider configuration.
- The provider automatically handles workspace-level authentication.