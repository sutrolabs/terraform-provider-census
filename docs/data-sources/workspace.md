# census_workspace Data Source

~> **Deprecated:** This data source is deprecated and no longer returns data from Fivetran Activations. Workspaces are no longer needed; remove it from your configuration.

Previously retrieved a workspace from Fivetran Activations.

## Behavior

This data source remains in the provider only so that existing configurations keep working. It will not be removed.

It makes no API calls. It returns the `id` you give it, empty values for every other attribute, and a warning.

## Example Usage

```hcl
data "census_workspace" "example" {
  id = "123"
}
```

## Argument Reference

* `id` - (Required) The ID of the workspace. Returned unchanged.

## Attribute Reference

* `id` - The ID given in the configuration.
* `name`, `organization_id`, `created_at`, `notification_emails` - Always empty.
