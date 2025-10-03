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
      is_primary_identifier = true
    },
    {
      from = "first_name"
      to   = "FirstName"
    },
    {
      from = "last_name"
      to   = "LastName"
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
      from                  = "email"
      to                    = "email"
      is_primary_identifier = true
    },
    {
      from = "lifetime_value"
      to   = "lifetime_value"
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
      is_primary_identifier = true
    },
    {
      from = "email"
      to   = "email_hash"
      type = "hash"
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
      from                  = "email"
      to                    = "Email"
      is_primary_identifier = true
    },
    {
      type     = "constant"
      constant = "Terraform Managed"
      to       = "LeadSource"
    },
  ]

  operation = "upsert"
}
```

### Sync with Sync Metadata Mapping

```hcl
resource "census_sync" "metadata_sync" {
  workspace_id   = census_workspace.main.id
  name           = "Sync with Metadata Tracking"

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
      is_primary_identifier = true
    },
    {
      from = "first_name"
      to   = "FirstName"
    },
    {
      # Map Census sync_run_id to a custom field
      type             = "sync_metadata"
      sync_metadata_key = "sync_run_id"
      to               = "Last_Sync_Run_ID__c"
    },
  ]

  operation = "upsert"
}
```

### Sync with Segment Membership

```hcl
resource "census_sync" "segment_sync" {
  workspace_id   = census_workspace.main.id
  name           = "Sync with Segment Data"

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
      is_primary_identifier = true
    },
    {
      # Map segment membership information
      type              = "segment_membership"
      segment_identify_by = "name"
      to                = "Active_Segments__c"
    },
  ]

  operation = "upsert"
}
```

### Sync with Liquid Template Transformation

```hcl
resource "census_sync" "template_sync" {
  workspace_id   = census_workspace.main.id
  name           = "Sync with Field Transformations"

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
      is_primary_identifier = true
    },
    {
      from = "first_name"
      to   = "FirstName"
    },
    {
      # Use Liquid template to transform data
      type           = "liquid_template"
      liquid_template = "{{ record['status'] | upcase }}"
      to             = "Account_Status__c"
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
      is_primary_identifier = true
    },
    {
      # Map a constant value to user_list_id via lookup
      # This looks up the user_list record where id = "6600827417"
      type          = "constant"
      constant      = "6600827417"
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
  advanced_configuration = jsonencode({
    file_format        = "Parquet"
    csv_delimiter      = ","
    csv_include_header = true
  })

  schedule {
    frequency = "hourly"
    minute    = 35
  }
}
```

### Sync with Alert Configurations

```hcl
resource "census_sync" "monitored_sync" {
  workspace_id   = census_workspace.main.id
  name           = "High-Priority Customer Sync with Alerts"

  source_attributes = jsonencode({
    connection_id = census_source.warehouse.id
    object = {
      type       = "table"
      table_name = "customers"
    }
  })

  destination_object = "Contact"

  field_mapping = [
    {
      from                  = "email"
      to                    = "Email"
      is_primary_identifier = true
    },
    {
      from = "name"
      to   = "Name"
    },
  ]

  operation = "upsert"

  # Configure multiple alerts
  alert = [
    {
      # Alert when sync fails completely
      type                 = "FailureAlertConfiguration"
      send_for             = "first_time"
      should_send_recovery = true
      options              = {}
    },
    {
      # Alert when more than 50% of records are invalid
      type                 = "InvalidRecordPercentAlertConfiguration"
      send_for             = "every_time"
      should_send_recovery = true
      options = {
        threshold = "50"
      }
    },
    {
      # Alert when sync runtime exceeds 30 minutes
      type                 = "RuntimeAlertConfiguration"
      send_for             = "first_time"
      should_send_recovery = false
      options = {
        threshold  = "30"
        unit       = "minutes"
        start_type = "actual"
      }
    },
    {
      # Alert on sync completion
      type                 = "StatusAlertConfiguration"
      send_for             = "every_time"
      should_send_recovery = false
      options = {
        status_name = "completed"
      }
    },
  ]

  schedule {
    frequency = "hourly"
    minute    = 0
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
      is_primary_identifier = true
    },
    {
      from = "name"
      to   = "Name"
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

### Append Sync with High Water Mark

```hcl
resource "census_sync" "incremental_append" {
  workspace_id   = census_workspace.main.id
  name           = "Incremental Event Log Sync"

  source_attributes = jsonencode({
    connection_id = census_source.warehouse.id
    object = {
      type       = "table"
      table_name = "event_logs"
    }
  })

  destination_object = "Event__c"

  field_mapping = [
    {
      from                  = "event_id"
      to                    = "Event_ID__c"
      is_primary_identifier = true
    },
    {
      from = "event_name"
      to   = "Name"
    },
    {
      from = "updated_at"
      to   = "Updated_At__c"
    },
  ]

  operation = "append"

  # Use high water mark to only sync new records based on timestamp
  # This is more efficient than Census's default diff engine for append operations
  high_water_mark_attribute = "updated_at"

  schedule {
    frequency = "hourly"
    minute    = 15
  }
}
```

### Sync with Field Preservation and Null Value Control

```hcl
resource "census_sync" "preserve_example" {
  workspace_id   = census_workspace.main.id
  name           = "Customer Sync with Field Preservation"

  source_attributes = jsonencode({
    connection_id = census_source.warehouse.id
    object = {
      type       = "table"
      table_name = "customers"
    }
  })

  destination_object = "Contact"

  field_mapping = [
    {
      from                  = "email"
      to                    = "Email"
      is_primary_identifier = true
    },
    {
      from = "first_name"
      to   = "FirstName"
    },
    {
      # Don't overwrite existing phone numbers in destination
      from            = "phone"
      to              = "Phone"
      preserve_values = true
      sync_null_values = false  # Don't sync null phone values
    },
    {
      # Generate a custom field in the destination
      from           = "customer_tier"
      to             = "Customer_Tier__c"
      generate_field = true
    },
  ]

  operation = "upsert"

  schedule {
    frequency = "daily"
    hour      = 9
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
* `destination_attributes` - (Required) Destination configuration block:
  * `connection_id` - (Required) The destination connection ID
  * `object` - (Required) The destination object name (e.g., "Contact" for Salesforce, "contacts" for HubSpot)
  * `lead_union_insert_to` - (Optional) Where to insert a union object (for Salesforce connections only)
* `field_mapping` - (Optional) Set of field mappings between source and destination. Each mapping includes:
  * `from` - Source field name (required for `type="direct"` or `type="hash"`). Omit for `constant`, `sync_metadata`, `segment_membership`, and `liquid_template` mappings.
  * `to` - Destination field name (required)
  * `type` - Mapping type: `"direct"` (default), `"hash"`, `"constant"`, `"sync_metadata"`, `"segment_membership"`, or `"liquid_template"`.
  * `constant` - Constant value (must also set `type="constant"`)
  * `sync_metadata_key` - Sync metadata key (e.g., `"sync_run_id"`). Must also set `type="sync_metadata"`.
  * `segment_identify_by` - How to identify segments (e.g., `"name"`). Must also set `type="segment_membership"`.
  * `liquid_template` - Liquid template for data transformation (e.g., `"{{ record['field'] | upcase }}"`). Must also set `type="liquid_template"`.
  * `is_primary_identifier` - (Optional) Boolean indicating if this field is the primary identifier for matching records. Exactly one field_mapping must have this set to `true`. Defaults to `false`.
  * `lookup_object` - (Optional) Object to lookup for relationship mapping (e.g., `"user_list"`). Used with `lookup_field` for foreign key lookups.
  * `lookup_field` - (Optional) Field to use for lookup in the `lookup_object` (e.g., `"id"`). Used with `lookup_object` for foreign key lookups.
  * `preserve_values` - (Optional) If true, preserves existing values in the destination field and prevents Census from overwriting them. Defaults to `false`.
  * `generate_field` - (Optional) If true, Census will generate/create this field in the destination. Defaults to `false`.
  * `sync_null_values` - (Optional) If true (default), null values in the source will be synced to the destination. Set to false to skip syncing null values. Defaults to `true`.
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
* `advanced_configuration` - (Optional) Advanced configuration options specific to the destination type as JSON string. Use `jsonencode()` to specify values. Available options vary by destination (e.g., file format for file exports, bulk settings for APIs). Values can be strings, numbers, or booleans. Refer to destination-specific Census documentation for available options.
* `high_water_mark_attribute` - (Optional) The name of the timestamp column to use for high water mark diffing strategy. When set, append syncs will use this column to identify new records instead of the default Census diff engine (using primary keys). This is more efficient for append operations with timestamp-based data. Example: `"updated_at"`.
* `alert` - (Optional) Set of alert configurations for monitoring sync health. Multiple alerts can be configured. Each alert includes:
  * `type` - (Required) Type of alert. Valid values:
    * `"FailureAlertConfiguration"` - Alert when sync fails completely
    * `"InvalidRecordPercentAlertConfiguration"` - Alert when invalid/rejected records exceed threshold
    * `"FullSyncTriggerAlertConfiguration"` - Alert when a full sync is triggered
    * `"RecordCountDeviationAlertConfiguration"` - Alert when record counts deviate from expected
    * `"RuntimeAlertConfiguration"` - Alert when sync runtime exceeds threshold
    * `"StatusAlertConfiguration"` - Alert on sync status changes (started, completed)
  * `send_for` - (Optional) When to send alerts: `"first_time"` (default, only first violation) or `"every_time"` (every violation)
  * `should_send_recovery` - (Optional) Whether to send recovery notification when condition resolves. Defaults to `true`.
  * `options` - (Optional) Alert-specific configuration options (values as strings):
    * For `InvalidRecordPercentAlertConfiguration`:
      * `threshold` - Percentage (0-100) of invalid records that triggers alert
    * For `RecordCountDeviationAlertConfiguration`:
      * `threshold` - Percentage (0-100) deviation from expected count
      * `record_type` - Type to monitor: `source_record_count`, `records_updates`, `records_deletes`, `records_invalid`, `records_processed`, `records_updated`, or `records_failed`
    * For `RuntimeAlertConfiguration`:
      * `threshold` - Number of time units before alert
      * `unit` - Time unit: `"minutes"` or `"hours"`
      * `start_type` - When to start measuring: `"actual"` (when sync actually starts) or `"scheduled"` (from scheduled time)
    * For `StatusAlertConfiguration`:
      * `status_name` - Status to alert on: `"started"` or `"completed"`
  * `id` - (Computed) The alert configuration ID assigned by Census
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