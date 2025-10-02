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
      from                  = "email"
      to                    = "Email"
      operation             = "direct"
      is_primary_identifier = true
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
      is_primary_identifier = true
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
      from                  = "id"
      to                    = "userId"
      operation             = "direct"
      is_primary_identifier = true
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
      is_primary_identifier = true
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

### Sync with Automatic Field Mapping (Sync All Properties)

```hcl
resource "census_sync" "auto_sync" {
  workspace_id   = census_workspace.main.id
  name           = "Auto-Mapped Users Sync"

  source_attributes = jsonencode({
    connection_id = census_source.warehouse.id
    object = {
      type       = "table"
      table_name = "users"
    }
  })

  destination_object = "Contact"

  # Automatically sync all properties from source to destination
  field_behavior      = "sync_all_properties"
  field_normalization = "snake_case"  # Format field names in snake_case
  field_order         = "mapping_order"

  # Only need to define the primary identifier when using sync_all_properties
  field_mapping = [
    {
      from                  = "email"
      to                    = "Email"
      operation             = "direct"
      is_primary_identifier = true
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

### Sync with Lookup Field (Foreign Key Relationship)

```hcl
resource "census_sync" "user_list_sync" {
  workspace_id   = census_workspace.main.id
  name           = "Users to Google Ads Customer Match"

  source_attributes = jsonencode({
    connection_id = census_source.warehouse.id
    object = {
      type       = "table"
      table_name = "users"
    }
  })

  destination_object = "user_data"

  field_mapping = [
    {
      from                  = "email"
      to                    = "user_identifier.hashed_email"
      operation             = "direct"
      is_primary_identifier = true
    },
    {
      # Map a constant value to user_list_id via lookup
      # This looks up the user_list record where id = "6600827417"
      constant      = "6600827417"
      operation     = "constant"
      to            = "user_list_id"
      lookup_object = "user_list"
      lookup_field  = "id"
    },
  ]

  operation = "mirror"

  schedule {
    frequency = "hourly"
    minute    = 10
  }
}
```

### Sync with Advanced Configuration (File Export)

```hcl
resource "census_sync" "blob_storage_sync" {
  workspace_id   = census_workspace.main.id
  name           = "Users to Azure Blob Storage"

  source_attributes = jsonencode({
    connection_id = census_source.warehouse.id
    object = {
      type  = "model"
      id    = "21130"
    }
  })

  destination_object = "path_to_file/data_%m-%d-%y.parquet"

  field_mapping = [
    {
      from = "email"
      to   = "EMAIL"
    },
  ]

  operation     = "mirror"
  field_behavior = "sync_all_properties"
  field_normalization = "match_source_names"

  # Advanced configuration for file export
  advanced_configuration = {
    file_format        = "Parquet"
    csv_delimiter      = ","
    csv_include_header = "true"
  }

  schedule {
    frequency = "hourly"
    minute    = 35
  }
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
      from                  = "product_id"
      to                    = "ProductCode"
      operation             = "direct"
      is_primary_identifier = true
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
  * `is_primary_identifier` - (Optional) Boolean indicating if this field is the primary identifier for matching records. Exactly one field_mapping must have this set to `true`. Defaults to `false`.
  * `lookup_object` - (Optional) Object to lookup for relationship mapping (e.g., `"user_list"`). Used with `lookup_field` for foreign key lookups.
  * `lookup_field` - (Optional) Field to use for lookup in the `lookup_object` (e.g., `"id"`). Used with `lookup_object` for foreign key lookups.
* `operation` - (Optional) Sync mode: `"upsert"`, `"append"`, or `"mirror"`. Defaults to `"upsert"`.
* `field_behavior` - (Optional) Controls how fields are synced:
  * `"specific_properties"` (default) - Use only the field mappings defined in `field_mapping`
  * `"sync_all_properties"` - Automatically sync all properties from source to destination
* `field_normalization` - (Optional) When `field_behavior` is `"sync_all_properties"`, specifies how automatic field names should be normalized:
  * `"start_case"` - Start Case (e.g., "First Name")
  * `"lower_case"` - lower case (e.g., "first name")
  * `"upper_case"` - UPPER CASE (e.g., "FIRST NAME")
  * `"camel_case"` - camelCase (e.g., "firstName")
  * `"snake_case"` - snake_case (e.g., "first_name")
  * `"match_source_names"` - Use exact source field names
* `field_order` - (Optional) Specifies how destination fields should be ordered. Only applicable for destinations that support field ordering:
  * `"alphabetical_column_name"` (default) - Sort fields alphabetically
  * `"mapping_order"` - Use the order fields are defined in `field_mapping`
* `advanced_configuration` - (Optional) Map of advanced configuration options specific to the destination type. Available options vary by destination (e.g., file format for file exports, bulk settings for APIs). Values must be strings. Refer to destination-specific Census documentation for available options.
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