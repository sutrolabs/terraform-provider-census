# census_destination Resource

Manages a Census destination connection. Destinations connect to business tools like Salesforce, HubSpot, and other SaaS applications where you want to sync your data.

## Example Usage

### Salesforce Destination

```hcl
resource "census_destination" "salesforce" {
  workspace_id = census_workspace.main.id
  name         = "Production Salesforce"
  type         = "salesforce"

  credentials = jsonencode({
    username      = "census@company.com"
    password      = var.salesforce_password
    security_token = var.salesforce_security_token
    instance_url  = "https://company.my.salesforce.com"
  })
}
```

### HubSpot Destination

```hcl
resource "census_destination" "hubspot" {
  workspace_id = census_workspace.main.id
  name         = "Marketing HubSpot"
  type         = "hubspot"

  credentials = jsonencode({
    access_token = var.hubspot_access_token
  })
}
```

### Intercom Destination

```hcl
resource "census_destination" "intercom" {
  workspace_id = census_workspace.main.id
  name         = "Customer Support"
  type         = "intercom"

  credentials = jsonencode({
    access_token = var.intercom_access_token
  })
}
```

## Argument Reference

* `workspace_id` - (Required, Forces new resource) The ID of the workspace this destination belongs to.
* `name` - (Required) The name of the destination.
* `type` - (Required, Forces new resource) The type of destination connector. Supported types include:
  - `salesforce`
  - `hubspot`
  - `intercom`
  - `segment`
  - `marketo`
  - `braze`
  - And many more... (validated against Census API)
* `credentials` - (Required, Sensitive) JSON-encoded credentials for connecting to the destination. The required fields vary by destination type and are validated against the Census API schema.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the destination.
* `connection_status` - The current connection status of the destination.

## Import

Destinations can be imported using the workspace ID and destination ID separated by a colon:

```shell
terraform import census_destination.salesforce "workspace_id:destination_id"
```

For example:

```shell
terraform import census_destination.salesforce "12345:67890"
```

## Notes

* The `credentials` field is marked as sensitive and will not be displayed in Terraform output.
* Destination types and required credential fields are validated against the Census API's `/connectors` endpoint.
* The provider automatically refreshes destination metadata after creation to discover available objects.