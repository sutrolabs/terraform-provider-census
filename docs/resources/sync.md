# census_sync Resource

Manages a sync that moves data from a source (table, dataset, model, etc.) to a destination (Salesforce, HubSpot, etc.) with configurable field mappings and scheduling. A sync belongs to an activation (see `census_activation`), which pairs the source with the destination.

## Example Usage

### Basic Sync with Field Mappings

```hcl
resource "census_sync" "user_sync" {
  activation_id = census_activation.salesforce.id
  label         = "Users to Salesforce"

  source_attributes {
    object {
      type          = "table"
      table_name    = "users"
      table_schema  = "demo"
      table_catalog = "dev"
    }
  }

  destination_attributes {
    object = "Contact"
  }

  field_mapping {
    from                  = "email"
    to                    = "Email"
    is_primary_identifier = true
  }

  field_mapping {
    from = "first_name"
    to   = "FirstName"
  }

  field_mapping {
    from = "last_name"
    to   = "LastName"
  }

  operation = "upsert"

  run_mode {
    type = "triggered"
    triggers {
      schedule {
        frequency = "hourly"
        minute    = 0
      }
    }
  }
}
```

### Sync with Dataset Source

```hcl
resource "census_sync" "high_value_sync" {
  activation_id = census_activation.hubspot.id
  label         = "High Value Customers to HubSpot"

  source_attributes {
    object {
      type = "dataset"
      id   = census_dataset.high_value_customers.id
    }
  }

  destination_attributes {
    object = "contacts"
  }

  field_mapping {
    from                  = "email"
    to                    = "email"
    is_primary_identifier = true
  }

  field_mapping {
    from = "lifetime_value"
    to   = "lifetime_value"
  }

  operation = "upsert"

  run_mode {
    type = "triggered"
    triggers {
      schedule {
        frequency = "daily"
        hour      = 8
        minute    = 0
      }
    }
  }
}
```

### Sync with Segment Source

```hcl
resource "census_sync" "vip_segment_sync" {
  activation_id = census_activation.salesforce.id
  label         = "VIP Users Segment to Salesforce"

  source_attributes {
    object {
      type       = "segment"
      id         = "3060" # The segment ID
      dataset_id = "5951" # The dataset ID that the segment belongs to
    }
  }

  destination_attributes {
    object = "Contact"
  }

  field_mapping {
    from                  = "email"
    to                    = "Email"
    is_primary_identifier = true
  }

  field_mapping {
    from = "first_name"
    to   = "FirstName"
  }

  field_mapping {
    from = "last_name"
    to   = "LastName"
  }

  operation = "upsert"

  run_mode {
    type = "triggered"
    triggers {
      schedule {
        frequency = "daily"
        hour      = 2
        minute    = 0
      }
    }
  }
}
```

### Sync with Constant Value

```hcl
resource "census_sync" "tagged_sync" {
  activation_id = census_activation.salesforce.id
  label         = "Tagged Contact Sync"

  source_attributes {
    object {
      type          = "table"
      table_name    = "users"
      table_schema  = "demo"
      table_catalog = "dev"
    }
  }

  destination_attributes {
    object = "Contact"
  }

  field_mapping {
    from                  = "email"
    to                    = "Email"
    is_primary_identifier = true
  }

  field_mapping {
    type     = "constant"
    constant = "Terraform Managed"
    to       = "LeadSource"
  }

  operation = "upsert"
}
```

### Sync with Sync Metadata Mapping

```hcl
resource "census_sync" "metadata_sync" {
  activation_id = census_activation.salesforce.id
  label         = "Sync with Metadata Tracking"

  source_attributes {
    object {
      type          = "table"
      table_name    = "users"
      table_schema  = "demo"
      table_catalog = "dev"
    }
  }

  destination_attributes {
    object = "Contact"
  }

  field_mapping {
    from                  = "email"
    to                    = "Email"
    is_primary_identifier = true
  }

  field_mapping {
    from = "first_name"
    to   = "FirstName"
  }

  field_mapping {
    # Map sync_run_id to a custom field
    type              = "sync_metadata"
    sync_metadata_key = "sync_run_id"
    to                = "Last_Sync_Run_ID__c"
  }

  operation = "upsert"
}
```

### Sync with Segment Membership

```hcl
resource "census_sync" "segment_sync" {
  activation_id = census_activation.salesforce.id
  label         = "Sync with Segment Data"

  source_attributes {
    object {
      type          = "table"
      table_name    = "users"
      table_schema  = "demo"
      table_catalog = "dev"
    }
  }

  destination_attributes {
    object = "Contact"
  }

  field_mapping {
    from                  = "email"
    to                    = "Email"
    is_primary_identifier = true
  }

  field_mapping {
    # Map segment membership information
    type                = "segment_membership"
    segment_identify_by = "name"
    to                  = "Active_Segments__c"
  }

  operation = "upsert"
}
```

### Sync with Liquid Template Transformation

```hcl
resource "census_sync" "template_sync" {
  activation_id = census_activation.salesforce.id
  label         = "Sync with Field Transformations"

  source_attributes {
    object {
      type          = "table"
      table_name    = "users"
      table_schema  = "demo"
      table_catalog = "dev"
    }
  }

  destination_attributes {
    object = "Contact"
  }

  field_mapping {
    from                  = "email"
    to                    = "Email"
    is_primary_identifier = true
  }

  field_mapping {
    from = "first_name"
    to   = "FirstName"
  }

  field_mapping {
    # Use Liquid template to transform data
    type            = "liquid_template"
    liquid_template = "{{ record['status'] | upcase }}"
    to              = "Account_Status__c"
  }

  operation = "upsert"
}
```

### Sync with Automatic Field Mapping (Sync All Properties)

```hcl
resource "census_sync" "auto_sync" {
  activation_id = census_activation.salesforce.id
  label         = "Auto-Mapped Users Sync"

  source_attributes {
    object {
      type          = "table"
      table_name    = "users"
      table_schema  = "demo"
      table_catalog = "dev"
    }
  }

  destination_attributes {
    object = "Contact"
  }

  # Automatically sync all properties from source to destination
  field_behavior      = "sync_all_properties"
  field_normalization = "snake_case" # Format field names in snake_case
  field_order         = "mapping_order"

  # Only need to define the primary identifier when using sync_all_properties
  field_mapping {
    from                  = "email"
    to                    = "Email"
    is_primary_identifier = true
  }

  operation = "upsert"

  run_mode {
    type = "triggered"
    triggers {
      schedule {
        frequency = "daily"
        hour      = 8
        minute    = 0
      }
    }
  }
}
```

### Sync to Google Sheets (Mirror with Auto-Mapping)

```hcl
resource "census_sync" "sheets_sync" {
  activation_id = census_activation.google_sheets.id
  label         = "Data Export to Google Sheets"

  source_attributes {
    object {
      type          = "table"
      table_name    = "users"
      table_schema  = "demo"
      table_catalog = "dev"
    }
  }

  destination_attributes {
    # Google Sheets object format: JSON with spreadsheet_id and sheet_id
    object = "{\"spreadsheet_id\":\"1r9HavWIo-CS14sblFl-Kxa2kVbMGlqWIBDU6Hb_W8sA\",\"sheet_id\":0}"
  }

  operation = "mirror"

  # Automatically sync all properties - no field mappings needed
  field_behavior      = "sync_all_properties"
  field_normalization = "match_source_names"

  run_mode {
    type = "triggered"
    triggers {
      schedule {
        frequency = "daily"
        hour      = 6
        minute    = 0
      }
    }
  }
}
```

### Sync with Lookup Field (Foreign Key Relationship)

```hcl
resource "census_sync" "user_list_sync" {
  activation_id = census_activation.google_ads.id
  label         = "Users to Google Ads Customer Match"

  source_attributes {
    object {
      type          = "table"
      table_name    = "users"
      table_schema  = "demo"
      table_catalog = "dev"
    }
  }

  destination_attributes {
    object = "user_data"
  }

  field_mapping {
    from                  = "email"
    to                    = "user_identifier.hashed_email"
    is_primary_identifier = true
  }

  field_mapping {
    # Map a constant value to user_list_id via lookup
    # This looks up the user_list record where id = "6600827417"
    type          = "constant"
    constant      = "6600827417"
    to            = "user_list_id"
    lookup_object = "user_list"
    lookup_field  = "id"
  }

  operation = "mirror"

  run_mode {
    type = "triggered"
    triggers {
      schedule {
        frequency = "hourly"
        minute    = 10
      }
    }
  }
}
```

### Sync with Advanced Configuration (File Export)

```hcl
resource "census_sync" "blob_storage_sync" {
  activation_id = census_activation.azure_blob.id
  label         = "Users to Azure Blob Storage"

  source_attributes {
    object {
      type = "model"
      id   = "21130"
    }
  }

  destination_attributes {
    object = "path_to_file/data_%m-%d-%y.parquet"
  }

  operation           = "mirror"
  field_behavior      = "sync_all_properties"
  field_normalization = "match_source_names"

  # Advanced configuration for file export
  advanced_configuration = jsonencode({
    file_format        = "Parquet"
    csv_delimiter      = ","
    csv_include_header = true
  })

  run_mode {
    type = "triggered"
    triggers {
      schedule {
        frequency = "hourly"
        minute    = 35
      }
    }
  }
}
```

### Mirror sync to file system destination (S3/SFTP/Azure Blob Storage/Google Drive) with sync all properties

Mirror syncs to file system destinations do not support primary identifier mappings, so use `sync_all_properties` with no `field_mapping` blocks.

```hcl
resource "census_sync" "s3_mirror_sync_all_properties" {
  activation_id = census_activation.s3.id
  label         = "Mirror to S3 - sync all properties"
  operation     = "mirror"

  source_attributes {
    object {
      type = "dataset"
      id   = census_dataset.events.id
    }
  }

  destination_attributes {
    object = "path/test_%m-%d-%y.csv"
  }

  # Automatically sync all properties from source to destination
  field_behavior      = "sync_all_properties"
  field_normalization = "snake_case" # Format field names in snake_case
  field_order         = "mapping_order"
}
```

### Upsert sync to file system destination (S3/SFTP/Azure Blob Storage/Google Drive) with sync specific properties

Upsert syncs to file system destinations require a primary identifier mapping to uniquely identify rows in the source between sync runs.

```hcl
resource "census_sync" "s3_upsert_sync_specific_properties" {
  activation_id  = census_activation.s3.id
  label          = "Upsert sync to S3 - sync specific properties"
  operation      = "upsert"
  field_behavior = "specific_properties"

  source_attributes {
    object {
      type = "dataset"
      id   = census_dataset.events.id
    }
  }

  destination_attributes {
    object = "path/test_%m-%d-%y.csv"
  }

  # Upsert syncs to file system destinations require
  # primary identifier mappings to the "unique_id" field.
  field_mapping {
    from                  = "id"
    to                    = "unique_id"
    is_primary_identifier = true
  }

  # Other mappings should set generate_field
  field_mapping {
    from           = "big_column"
    to             = "big_column"
    generate_field = true
  }
}
```

### Upsert sync to file system destination (S3/SFTP/Azure Blob Storage/Google Drive) with sync all properties

Upsert syncs to file system destinations require a primary identifier mapping to uniquely identify rows in the source between sync runs.

```hcl
resource "census_sync" "s3_upsert_sync_all_properties" {
  activation_id = census_activation.s3.id
  label         = "Upsert sync to S3 - sync all properties"
  operation     = "upsert"

  source_attributes {
    object {
      type = "dataset"
      id   = census_dataset.events.id
    }
  }

  destination_attributes {
    object = "path/test_%m-%d-%y.csv"
  }

  # Automatically sync all properties from source to destination
  field_behavior      = "sync_all_properties"
  field_normalization = "snake_case" # Format field names in snake_case
  field_order         = "mapping_order"

  # Upsert syncs to file system destinations require
  # primary identifier mappings to the "unique_id" field.
  field_mapping {
    from                  = "id"
    to                    = "unique_id"
    is_primary_identifier = true
  }
}
```

### Sync with Alert Configurations

```hcl
resource "census_sync" "monitored_sync" {
  activation_id = census_activation.salesforce.id
  label         = "High-Priority Customer Sync with Alerts"

  source_attributes {
    object {
      type          = "table"
      table_name    = "users"
      table_schema  = "demo"
      table_catalog = "dev"
    }
  }

  destination_attributes {
    object = "Contact"
  }

  field_mapping {
    from                  = "email"
    to                    = "Email"
    is_primary_identifier = true
  }

  field_mapping {
    from = "name"
    to   = "Name"
  }

  operation = "upsert"

  # Configure multiple alerts
  alert {
    # Alert when sync fails completely
    type                 = "FailureAlertConfiguration"
    send_for             = "first_time"
    should_send_recovery = true
    options              = {}
  }

  alert {
    # Alert when more than 50% of records are invalid
    type                 = "InvalidRecordPercentAlertConfiguration"
    send_for             = "every_time"
    should_send_recovery = true
    options = {
      threshold = "50"
    }
  }

  alert {
    # Alert when sync runtime exceeds 30 minutes
    type                 = "RuntimeAlertConfiguration"
    send_for             = "first_time"
    should_send_recovery = false
    options = {
      threshold  = "30"
      unit       = "minutes"
      start_type = "actual"
    }
  }

  alert {
    # Alert on sync completion
    type                 = "StatusAlertConfiguration"
    send_for             = "every_time"
    should_send_recovery = false
    options = {
      status_name = "completed"
    }
  }

  run_mode {
    type = "triggered"
    triggers {
      schedule {
        frequency = "hourly"
        minute    = 0
      }
    }
  }
}
```

### Mirror Sync (Replace All)

```hcl
resource "census_sync" "mirror_sync" {
  activation_id = census_activation.salesforce.id
  label         = "Product Catalog Mirror"

  source_attributes {
    object {
      type          = "table"
      table_name    = "products"
      table_schema  = "demo"
      table_catalog = "dev"
    }
  }

  destination_attributes {
    object = "Product2"
  }

  field_mapping {
    from                  = "product_id"
    to                    = "ProductCode"
    is_primary_identifier = true
  }

  field_mapping {
    from = "name"
    to   = "Name"
  }

  operation = "mirror"

  run_mode {
    type = "triggered"
    triggers {
      schedule {
        frequency = "daily"
        hour      = 2
        minute    = 0
      }
    }
  }
}
```

### Append Sync with High Water Mark

```hcl
resource "census_sync" "incremental_append" {
  activation_id = census_activation.salesforce.id
  label         = "Incremental Event Log Sync"

  source_attributes {
    object {
      type          = "table"
      table_name    = "events"
      table_schema  = "demo"
      table_catalog = "dev"
    }
  }

  destination_attributes {
    object = "Event__c"
  }

  field_mapping {
    from                  = "event_id"
    to                    = "Event_ID__c"
    is_primary_identifier = true
  }

  field_mapping {
    from = "event_name"
    to   = "Name"
  }

  field_mapping {
    from = "updated_at"
    to   = "Updated_At__c"
  }

  operation = "append"

  # Use high water mark to only sync new records based on timestamp
  # This is more efficient than Fivetran's default diff engine for append operations
  high_water_mark_attribute = "updated_at"

  run_mode {
    type = "triggered"
    triggers {
      schedule {
        frequency = "hourly"
        minute    = 15
      }
    }
  }
}
```

### Sync with Field Preservation and Null Value Control

```hcl
resource "census_sync" "preserve_example" {
  activation_id = census_activation.salesforce.id
  label         = "Customer Sync with Field Preservation"

  source_attributes {
    object {
      type          = "table"
      table_name    = "users"
      table_schema  = "demo"
      table_catalog = "dev"
    }
  }

  destination_attributes {
    object = "Contact"
  }

  field_mapping {
    from                  = "email"
    to                    = "Email"
    is_primary_identifier = true
  }

  field_mapping {
    from = "first_name"
    to   = "FirstName"
  }

  field_mapping {
    # Don't overwrite existing phone numbers in destination
    from             = "phone"
    to               = "Phone"
    preserve_values  = true  # Don't overwrite existing data in the destination
    sync_null_values = false # Don't sync null phone values
  }

  field_mapping {
    # Generate a custom field in the destination
    from           = "customer_tier"
    to             = "Customer_Tier__c"
    generate_field = true
  }

  operation = "upsert"

  run_mode {
    type = "triggered"
    triggers {
      schedule {
        frequency = "daily"
        hour      = 9
        minute    = 0
      }
    }
  }
}
```

## Argument Reference

* `activation_id` - (Required, Forces new resource) The ID of the activation this sync belongs to (see `census_activation`). The activation determines the sync's source and destination connections.
* `label` - (Required) The label of the sync.
* `operation` - (Required) How records are synced to the destination: `"upsert"`, `"append"`, `"mirror"`, `"update"`, or `"insert"`. Which operations are available depends on the destination object.
* `source_attributes` - (Required) Configuration block for the source. Block contains:
  * `object` - (Required) Object configuration block:
    * `type` - (Required) Source type: `"table"`, `"dataset"`, `"model"`, `"topic"`, `"segment"`, or `"cohort"`
    * For table sources: `table_name`, `table_schema`, and `table_catalog`
    * For dataset/model sources: `id` of the dataset/model
    * For segment sources: use `type="segment"`, provide the segment `id`, and specify `dataset_id` for the dataset the segment belongs to
    * For cohort sources: use `type="cohort"`, provide the cohort `id`, and specify `dataset_id` for the dataset the cohort belongs to
  * `cohort_id` - (Optional) The ID of the cohort, for cohort sources.
* `destination_attributes` - (Required) Destination configuration block:
  * `object` - (Required) The destination object name (e.g., "Contact" for Salesforce, "contacts" for HubSpot). For file system destinations, the file path.
  * `lead_union_insert_to` - (Optional) Where to insert a union object (for Salesforce connections only)
* `field_mapping` - (Optional) Field mappings between source and destination. Define multiple `field_mapping` blocks for multiple mappings. If any `field_mapping` blocks are defined, one of them must have `is_primary_identifier = true`. Each mapping block includes:
  * `from` - Source field name (required for `type="direct"`). Omit for `constant`, `sync_metadata`, `segment_membership`, and `liquid_template` mappings.
  * `to` - (Required) Destination field name.
  * `type` - (Optional) Mapping type: `"direct"` (default), `"constant"`, `"sync_metadata"`, `"segment_membership"`, or `"liquid_template"`.
  * `constant` - (Optional) Constant value. Must also set `type="constant"`.
  * `sync_metadata_key` - (Optional) Sync metadata key (e.g., `"sync_run_id"`). Must also set `type="sync_metadata"`.
  * `segment_identify_by` - (Optional) How to identify segments (e.g., `"name"`). Must also set `type="segment_membership"`.
  * `liquid_template` - (Optional) Liquid template for data transformation (e.g., `"{{ record['field'] | upcase }}"`). Must also set `type="liquid_template"`.
  * `is_primary_identifier` - (Optional) Whether this field is the primary identifier (sync key) for matching records. Defaults to `false`.
  * `lookup_object` - (Optional) Object to lookup for relationship mapping (e.g., `"user_list"`). Used with `lookup_field` for foreign key lookups.
  * `lookup_field` - (Optional) Field to use for lookup in the `lookup_object` (e.g., `"id"`). Used with `lookup_object` for foreign key lookups.
  * `preserve_values` - (Optional) If true, preserves existing values in the destination field and prevents Fivetran from overwriting them. Defaults to `false`.
  * `generate_field` - (Optional) If true, Fivetran will generate/create this field in the destination. Defaults to `false`.
  * `sync_null_values` - (Optional) If true, null values in the source will be synced to the destination. Set to false to skip syncing null values. Defaults to `true`.
  * `array_field` - (Optional) Whether the destination field is an array type. Only applicable when `generate_field` is true (for user-defined fields). Defaults to `false`.
  * `field_type` - (Optional) The type of the destination field. Only applicable when `generate_field` is true (for user-defined fields). Available types depend on the destination.
  * `follow_source_type` - (Optional) Whether the destination field type should automatically follow changes to the source column type. Defaults to `false`.
  * `operation` - (Optional) For array fields, how new values are merged with existing ones: `"append"` (add to the existing array) or `"set"` (replace it). Only applicable when `array_field` is true.
* `paused` - (Optional) Whether the sync is paused. Defaults to `false`.
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
  * `"source_column_order"` - Destination fields are ordered by the position of the columns returned from the source
* `sync_behavior_family` - (Optional, Deprecated) Ignored.
* `advanced_configuration` - (Optional) Advanced configuration options specific to the destination type as a JSON string. Use `jsonencode()` to specify values. Available options vary by destination (e.g., file format for file exports, bulk settings for APIs). Values can be strings, numbers, or booleans. Refer to destination-specific Fivetran Activations documentation for available options.
* `high_water_mark_attribute` - (Optional) The name of the timestamp column to use for high water mark diffing strategy. When set, append syncs will use this column to identify new records instead of the default Fivetran diff engine (using primary keys). Example: `"updated_at"`.
* `historical_sync_operation` - (Optional) Specifies how the first sync should handle historical records when using append operation. Only applicable for append syncs:
  * `"skip_current_records"` - Skip existing records on first sync, only sync new records going forward
  * `"backfill_all_records"` - Include all existing records on first sync (full backfill)
* `mirror_strategy` - (Optional) Specifies the strategy for mirror syncs. Only applicable when `operation` is set to `"mirror"`:
  * `"sync_updates_and_deletes"` - Incrementally syncs changes by inserting new records, updating modified records, and deleting records that no longer exist in the source. This is the most common strategy.
  * `"sync_updates_and_nulls"` - Updates existing records and sets fields to null when the source contains null values, without performing deletes.
  * `"upload_and_swap"` - Replaces the entire destination table with the current source snapshot.
* `alert` - (Optional) Alert configurations for monitoring sync health. Define multiple `alert` blocks to configure multiple alerts. Each alert block includes:
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
  * `id` - (Computed) The alert configuration ID assigned by Fivetran
* `run_mode` - (Optional) Run mode configuration block for controlling when the sync runs. If not specified, Fivetran uses a default schedule (manually triggered); specify `run_mode` explicitly to avoid drift.
  * `type` - (Required) Mode type. Only `"triggered"` is accepted.
  * `triggers` - (Optional) Trigger configurations. Multiple triggers can be configured simultaneously:
    * `schedule` - (Optional) Schedule-based trigger configuration block:
      * `frequency` - (Required) How often to run: `"never"`, `"continuous"`, `"quarter_hourly"`, `"hourly"`, `"daily"`, `"weekly"`, or `"expression"` (for cron)
      * `day` - (Optional) Day of week for weekly schedules: `"Sunday"`, `"Monday"`, `"Tuesday"`, `"Wednesday"`, `"Thursday"`, `"Friday"`, or `"Saturday"`
      * `hour` - (Optional) Hour to run (0-24) for daily/weekly schedules
      * `minute` - (Optional) Minute to run (0-59)
      * `cron_expression` - (Optional) Cron expression when `frequency` is `"expression"`
    * `dbt_cloud` - (Optional) dbt Cloud job trigger configuration block:
      * `project_id` - (Required) dbt Cloud project ID
      * `job_id` - (Required) dbt Cloud job ID
    * `fivetran` - (Optional) Fivetran connector trigger configuration block:
      * `job_id` - (Required) Fivetran job ID
      * `job_name` - (Required) Fivetran job name
    * `sync_sequence` - (Optional) Sync dependency trigger configuration block (triggers after another sync completes):
      * `sync_id` - (Required) ID of the sync to trigger after
* `sync_mode` - (Optional, Deprecated) Ignored. Use `operation` instead.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the sync.
* `status` - The current status of the sync.
* `created_at` - Timestamp when the sync was created.
* `updated_at` - Timestamp when the sync was last updated.
* `last_run_at` - (Deprecated) Timestamp when the sync was last executed. Unsupported and will be removed in a future version.
* `next_run_at` - (Deprecated) Timestamp when the sync is scheduled to run next. Unsupported and will be removed in a future version.
* `last_run_id` - (Deprecated) ID of the last sync run. Unsupported and will be removed in a future version.

## Import

Syncs can be imported using the sync ID:

```shell
terraform import census_sync.user_sync 67890
```

The `workspace_id:sync_id` format used by earlier provider versions is no longer accepted.

## Notes

* Field mappings are defined as multiple `field_mapping` blocks. The order of field mappings is preserved.
* The `source_attributes` and `destination_attributes` are configuration blocks, not JSON-encoded strings.
* Sync operations:
  * `upsert` - Insert new records and update existing ones
  * `append` - Only insert new records, never update
  * `mirror` - Replace all destination records with source data
  * `update` - Only update existing records, never insert
  * `insert` - Only insert records that don't exist yet
* Manual syncs (frequency="never") must be triggered externally.
* Source types determine which fields are required in `source_attributes.object`.
  * For segment sources, use `type="segment"` with `id` (segment ID) and `dataset_id` (parent dataset ID)
  * For cohort sources, use `type="cohort"` with `id` (cohort ID) and `dataset_id` (parent dataset ID)
* **Compound Key Mappings (Google Ads, etc.)**: Some destinations (like Google Ads Click Conversions) use compound keys for primary identifiers. Fivetran generates these mappings automatically; don't include them in your Terraform configuration. The provider filters out Fivetran-managed `compound_key` mappings on read to prevent drift.
