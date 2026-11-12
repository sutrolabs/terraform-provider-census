# census_activation Resource

Manages a Fivetran Activations activation, which pairs a source with a destination. Creating an activation also creates its destination connection from `type` and `credentials`. Use `census_activation` in place of `census_destination`, which can no longer be created.

## Example Usage

```hcl
resource "census_activation" "salesforce" {
  group_id = census_source.warehouse.group_id
  type     = "salesforce"
  name     = "Production Salesforce"

  credentials = {
    instance_url = "https://mycompany.my.salesforce.com"
    username     = var.salesforce_username
    password     = var.salesforce_password
  }
}
```

## Argument Reference

* `group_id` - (Required, Forces new resource) The `group_id` of the activation's source connection, usually `census_source.<name>.group_id`.
* `type` - (Required, Forces new resource) The destination type: the `service_name` of an activation type from the `GET /activation_types` API that can be created through the API.
* `credentials` - (Required, Sensitive) Credentials for the destination connection. The keys depend on `type`; see `GET /activation_types`. Values that are JSON objects or arrays (for example from `jsonencode`) are sent as JSON; all other values are sent as strings. The API does not return credentials, so they are not read back and Terraform keeps the values in your configuration.
* `name` - (Optional) The name of the activation. Defaults to the destination connection's name.

## Attribute Reference

* `id` - The ID of the activation.
* `last_test_succeeded` - Whether the last test of the destination connection succeeded. `false` if it has not been tested.
* `last_tested_at` - When the destination connection was last tested (RFC 3339), or empty if it has not been tested.
* `last_test_results` - The steps of the last test of the destination connection, each with:
  * `step` - The name of the step.
  * `status` - `success` or `error`.
  * `error_message` - The error message if the step failed.
* `connection_details` - The destination connection's credential values that the API returns. Non-string values are JSON-encoded.
* `is_file_system_activation` - Whether the destination is a file system destination (for example S3, SFTP or Google Cloud Storage). Syncs to a file system destination use the file path as the object name.

## Deleting

Destroying an activation also deletes its syncs and anything else that depends on it. Its destination connection is deleted too, unless another activation uses the same destination connection.

## Import

Import an activation by its ID:

```shell
terraform import census_activation.salesforce <activation_id>
```

### Migrating from `census_destination`

To import the activation for a source and a destination previously managed with `census_destination`, use the source's `group_id` and the destination's numeric ID, separated by a colon:

```shell
terraform import census_activation.salesforce <group_id>:<destination_id>
```

Because credentials are not read back, the first `terraform plan` after an import shows an in-place update to `credentials`. Applying it sends the configured credentials.
