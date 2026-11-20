terraform {
  required_providers {
    census = {
      source  = "sutrolabs/census"
      version = "~> 2.0"
    }
  }
}

provider "census" {
  personal_access_token = var.census_personal_token
  base_url              = var.census_base_url
}

# Create a source (warehouse connection)
resource "census_source" "warehouse" {
  name                = var.source_label
  type                = var.source_type
  notification_emails = var.notification_emails

  connection_config = var.source_credentials
}

# Create an activation pairing the source with a destination (e.g., Salesforce sandbox)
resource "census_activation" "crm" {
  group_id = census_source.warehouse.group_id
  name     = var.destination_label
  type     = var.destination_type

  credentials = var.destination_credentials
}

# Create a simple sync
resource "census_sync" "test_sync" {
  activation_id = census_activation.crm.id
  label         = var.sync_label

  source_attributes {
    object {
      type       = "table"
      table_name = var.source_table
    }
  }

  destination_attributes {
    object = var.destination_object
  }

  operation = "upsert"

  dynamic "field_mapping" {
    for_each = var.field_mapping
    content {
      from                  = field_mapping.value.from
      to                    = field_mapping.value.to
      is_primary_identifier = field_mapping.value.is_primary_identifier
    }
  }

  run_mode {
    type = "triggered"
    triggers {
      schedule {
        frequency = var.sync_frequency
        hour      = var.sync_hour
        minute    = var.sync_minute
      }
    }
  }

  paused = var.sync_paused
}