# census_workspace Resource

~> **Deprecated:** This resource is deprecated and no longer manages anything in Fivetran Activations. Workspaces are no longer needed; remove it from your configuration.

Previously managed a workspace in Fivetran Activations. The provider now authenticates with a personal access token alone, and sources, activations and syncs no longer belong to a workspace.

## Behavior

This resource remains in the provider only so that existing state and configurations keep working. It will not be removed.

* **Create** fails with an error.
* **Read, Update and Delete** do nothing and make no API calls. Changing the configuration of an existing `census_workspace` has no effect. Running `terraform destroy`, or removing the block, leaves the workspace in Fivetran Activations untouched.
* All arguments are retained so existing configurations continue to validate, but none of them are sent to Fivetran Activations.

## Migrating

Remove the `census_workspace` block from your configuration, along with any `workspace_id = census_workspace.<name>.id` references on other resources, which are now ignored. No changes are made in Fivetran Activations.

## Argument Reference

All arguments are ignored.

* `name` - (Optional) Ignored.
* `notification_emails` - (Optional) Ignored.
* `return_workspace_api_key` - (Optional) Ignored.

## Attribute Reference

* `id` - The ID of the workspace, as recorded in state.
* `organization_id`, `created_at`, `api_key` - Values recorded in state by earlier provider versions. Never refreshed.

## Import

Existing workspaces can be imported by ID. Importing only records the ID in state; nothing is read from Fivetran Activations.

```bash
terraform import census_workspace.example 123
```
