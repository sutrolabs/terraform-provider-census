# census_sync Resource

Manages a Census sync that moves data from a source (table, dataset, model, etc.) to a destination (Salesforce, HubSpot, etc.) with configurable field mappings and scheduling.

## Example Usage

### Basic Sync with Field Mappings

```hcl
resource "census_sync" "user_sync" {
  workspace_id   = census_workspace.main.id
  name           = "Users to Salesforce"

  source_attributes = jsonencode({
    connection_id = census_source.warehouse.id
    object = {
      type       = "table"
      table_name = "users"
    }
  })

  destination_object = "Contact"

  field_mapping = [
    {
      from      = "email"
      to        = "Email"
      operation = "direct"
    },
    {
      from      = "first_name"
      to        = "FirstName"
      operation = "direct"
    },
    {
      from      = "last_name"
      to        = "LastName"
      operation = "direct"
    },
  ]

  operation = "upsert"

  schedule {
    frequency = "hourly"
    minute    = 0
  }
}
```

### Sync with Dataset Source

```hcl
resource "census_sync" "high_value_sync" {
  workspace_id   = census_workspace.main.id
  name           = "High Value Customers to HubSpot"

  source_attributes = jsonencode({
    connection_id = census_source.warehouse.id
    object = {
      type = "dataset"
      id   = census_dataset.high_value_customers.id
    }
  })

  destination_object = "contacts"

  field_mapping = [
    {
      from      = "email"
      to        = "email"
      operation = "direct"
    },
    {
      from      = "lifetime_value"
      to        = "lifetime_value"
      operation = "direct"
    },
  ]

  operation = "upsert"

  schedule {
    frequency = "daily"
    hour      = 8
    minute    = 0
  }
}
```

### Sync with Hash Operation

```hcl
resource "census_sync" "secure_sync" {
  workspace_id   = census_workspace.main.id
  name           = "Hashed Email Sync"

  source_attributes = jsonencode({
    connection_id = census_source.warehouse.id
    object = {
      type       = "table"
      table_name = "users"
    }
  })

  destination_object = "user"

  field_mapping = [
    {
      from      = "id"
      to        = "userId"
      operation = "direct"
    },
    {
      from      = "email"
      to        = "email_hash"
      operation = "hash"
    },
  ]

  operation = "upsert"
}
```

### Sync with Constant Value

```hcl
resource "census_sync" "tagged_sync" {
  workspace_id   = census_workspace.main.id
  name           = "Tagged Contact Sync"

  source_attributes = jsonencode({
    connection_id = census_source.warehouse.id
    object = {
      type       = "table"
      table_name = "users"
    }
  })

  destination_object = "Contact"

  field_mapping = [
    {
      from      = "email"
      to        = "Email"
      operation = "direct"
    },
    {
      to        = "LeadSource"
      operation = "constant"
      constant  = "Terraform Managed"
    },
  ]

  operation = "upsert"
}
```

### Mirror Sync (Replace All)

```hcl
resource "census_sync" "mirror_sync" {
  workspace_id   = census_workspace.main.id
  name           = "Product Catalog Mirror"

  source_attributes = jsonencode({
    connection_id = census_source.warehouse.id
    object = {
      type       = "table"
      table_name = "products"
    }
  })

  destination_object = "Product2"

  field_mapping = [
    {
      from      = "product_id"
      to        = "ProductCode"
      operation = "direct"
    },
    {
      from      = "name"
      to        = "Name"
      operation = "direct"
    },
  ]

  operation = "mirror"

  schedule {
    frequency = "daily"
    hour      = 2
    minute    = 0
  }
}
```

## Argument Reference

* `workspace_id` - (Required, Forces new resource) The ID of the workspace this sync belongs to.
* `name` - (Required) The name of the sync.
* `source_attributes` - (Required) JSON-encoded configuration for the source. Must include:
  * `connection_id` - The source connection ID
  * `object` - Object configuration with:
    * `type` - Source type: `"table"`, `"dataset"`, `"model"`, `"topic"`, `"segment"`, or `"cohort"`
    * For table sources: `table_name`, optionally `table_schema` and `table_catalog`
    * For other sources: `id` of the dataset/model/etc.
* `destination_object` - (Required) The destination object name (e.g., "Contact" for Salesforce, "contacts" for HubSpot).
* `field_mapping` - (Optional) Set of field mappings between source and destination. Each mapping includes:
  * `from` - Source field name (required for non-constant operations)
  * `to` - Destination field name (required)
  * `operation` - Mapping operation: `"direct"`, `"hash"`, or `"constant"`. Defaults to `"direct"`.
  * `constant` - Constant value (required when operation is `"constant"`)
* `operation` - (Optional) Sync mode: `"upsert"`, `"append"`, or `"mirror"`. Defaults to `"upsert"`.
* `schedule` - (Optional) Scheduling configuration block:
  * `frequency` - (Required) `"hourly"`, `"daily"`, `"weekly"`, or `"manual"`
  * `minute` - (Optional) Minute of hour to run (0-59)
  * `hour` - (Optional) Hour of day to run (0-23) for daily/weekly syncs
  * `day_of_week` - (Optional) Day of week (0-6, Sunday=0) for weekly syncs
  * `timezone` - (Optional) Timezone for scheduling. Defaults to "UTC"

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the sync.
* `paused` - Whether the sync is currently paused.
* `status` - The current status of the sync.

## Import

Syncs can be imported using the workspace ID and sync ID separated by a colon:

```shell
terraform import census_sync.user_sync "workspace_id:sync_id"
```

For example:

```shell
terraform import census_sync.user_sync "12345:67890"
```

## Notes

* Field mappings use TypeSet to prevent drift from ordering changes returned by the API.
* The `source_attributes` structure must be OpenAPI compliant with proper table source format.
* Sync operations:
  * `upsert` - Insert new records and update existing ones
  * `append` - Only insert new records, never update
  * `mirror` - Replace all destination records with source data
* Manual syncs (frequency="manual") must be triggered externally.
* Source types determine which fields are required in `source_attributes.object`.