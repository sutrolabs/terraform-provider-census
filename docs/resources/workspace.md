# census_workspace Resource

Manages a Census workspace. Workspaces in Census are containers that organize your data syncs, destinations, and sources.

## Example Usage

```hcl
resource "census_workspace" "example" {
  name = "Production Workspace"
  notification_emails = [
    "alerts@company.com",
    "data-team@company.com"
  ]
  return_workspace_api_key = true
}
```

## Argument Reference

* `name` - (Required) The name of the workspace. Must be unique within the organization.
* `notification_emails` - (Optional) A list of email addresses that will receive alerts from the workspace.
* `return_workspace_api_key` - (Optional) Whether to return the workspace API key in the response during creation. Defaults to `false`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the workspace.
* `organization_id` - The ID of the organization the workspace belongs to.
* `created_at` - The timestamp when the workspace was created.
* `api_key` - The API key of the workspace. Only available during creation if `return_workspace_api_key` is `true`. This value is sensitive.

## Import

Workspaces can be imported using their ID:

```bash
terraform import census_workspace.example 123
```

## Notes

- The `api_key` attribute is only populated during resource creation when `return_workspace_api_key` is set to `true`.
- The API key is marked as sensitive and will not be displayed in Terraform output unless explicitly requested.
- Workspace names must be unique within your Census organization.
- Deleting a workspace will also delete all associated syncs, destinations, and sources. Use caution when destroying workspace resources.