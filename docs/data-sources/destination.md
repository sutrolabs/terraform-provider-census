# census_destination Data Source

~> **Deprecated:** This data source is deprecated and no longer returns data from Fivetran Activations. Use the `census_activation` data source instead.

Previously retrieved a destination connection from Fivetran Activations. To look up a destination connection, use the [`census_activation`](activation.md) data source.

## Behavior

This data source remains in the provider only so that existing configurations keep working. It will not be removed.

It makes no API calls. It returns the `id` you give it, empty values for every other attribute, and a warning.

## Example Usage

```hcl
data "census_destination" "salesforce" {
  id = "12345"
}
```

## Argument Reference

* `id` - (Required) The ID of the destination. Returned unchanged.
* `workspace_id` - (Optional) Ignored.

## Attribute Reference

* `id` - The ID given in the configuration.
* `name`, `type`, `status`, `test_status`, `last_tested`, `created_at`, `updated_at` - Always empty.
