# census_activation Data Source

Looks up an existing Fivetran Activations activation, either by its ID or by its source's `group_id` and the destination's numeric ID from `census_destination`.

## Example Usage

```hcl
data "census_activation" "by_id" {
  id = var.activation_id
}

data "census_activation" "by_source_and_destination" {
  group_id       = census_source.warehouse.group_id
  destination_id = 12345
}
```

## Argument Reference

Set either `id`, or both `group_id` and `destination_id`.

* `id` - (Optional) The ID of the activation.
* `group_id` - (Optional) The `group_id` of the activation's source connection. Requires `destination_id`.
* `destination_id` - (Optional) The numeric ID of the activation's destination connection (for example a `census_destination` ID). Requires `group_id`.

## Attribute Reference

* `id` - The ID of the activation.
* `group_id` - The `group_id` of the activation's source connection.
* `type` - The destination type (for example `salesforce`).
* `name` - The name of the activation.
* `last_test_succeeded` - Whether the last test of the destination connection succeeded. `false` if it has not been tested.
* `last_tested_at` - When the destination connection was last tested (RFC 3339), or empty if it has not been tested.
* `last_test_results` - The steps of the last test of the destination connection, each with `step`, `status` and `error_message`.
* `connection_details` - The destination connection's credential values that the API returns. Non-string values are JSON-encoded.
* `is_file_system_activation` - Whether the destination is a file system destination (for example S3, SFTP or Google Cloud Storage).
