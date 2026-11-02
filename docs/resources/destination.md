# census_destination Resource

~> **Deprecated:** This resource is deprecated and no longer manages anything in Fivetran Activations. Use the `census_activation` resource instead.

Previously managed a destination connection in Fivetran Activations. Destinations are now represented by [`census_activation`](activation.md), which connects one source to one destination connection.

## Behavior

This resource remains in the provider only so that existing state and configurations keep working. It will not be removed.

* **Create** fails with an error.
* **Read, Update and Delete** do nothing and make no API calls. Changing the configuration of an existing `census_destination` has no effect. Running `terraform destroy`, or removing the block, leaves the destination in Fivetran Activations untouched.
* All arguments are retained so existing configurations continue to validate, but none of them are sent to Fivetran Activations.

## Migrating to `census_activation`

1. Import the existing activation for each (source, destination) pair you use, rather than creating a new one. Fivetran Activations has already created an activation for every existing pairing.

   ```bash
   terraform import census_activation.salesforce "source_id:destination_id"
   ```

   `source_id` is the ID of the `census_source` and `destination_id` is the ID of the legacy `census_destination`. This is the only place the legacy destination ID is still used.

   ```hcl
   resource "census_activation" "salesforce" {
     group_id = census_source.warehouse.group_id
     type     = "salesforce"
     name     = "Production Salesforce"

     credentials = {
       username       = "census@company.com"
       password       = var.salesforce_password
       security_token = var.salesforce_security_token
       instance_url   = "https://company.my.salesforce.com"
     }
   }
   ```

2. Point each `census_sync` at the activation with `activation_id = census_activation.salesforce.id`. The `destination_attributes.connection_id` argument on `census_sync` is no longer used.
3. Remove the `census_destination` block from your configuration. No changes are made in Fivetran Activations.

## Argument Reference

All arguments are ignored.

* `workspace_id` - (Optional) Ignored.
* `name` - (Required) Ignored.
* `type` - (Required) Ignored.
* `connection_config` - (Required, Sensitive) Ignored.
* `auto_refresh_objects` - (Optional) Ignored.

## Attribute Reference

* `id` - The ID of the destination, as recorded in state.
* `status`, `test_status`, `last_tested`, `created_at`, `updated_at` - Values recorded in state by earlier provider versions. Never refreshed.

## Import

Existing destinations can be imported by ID. Importing only records the ID in state; nothing is read from Fivetran Activations. To manage the destination going forward, import a [`census_activation`](activation.md) instead.

```bash
terraform import census_destination.salesforce destination_id
```
