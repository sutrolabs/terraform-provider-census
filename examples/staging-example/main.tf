terraform {
  required_providers {
    census = {
      source  = "sutrolabs/census"
      version = "~> 0.1.0"
    }
  }
}

provider "census" {
  personal_access_token = var.census_personal_token
  base_url              = var.census_base_url
}

# Create a workspace for staging testing
resource "census_workspace" "staging_test" {
  name                = var.workspace_name
  notification_emails = var.notification_emails
}

# Create a data source (warehouse connection)
resource "census_source" "warehouse" {
  workspace_id = census_workspace.staging_test.id
  name         = var.source_label
  type         = var.source_type

  connection_config = var.source_credentials
}

# Create a destination (e.g., Salesforce sandbox)
resource "census_destination" "crm" {
  workspace_id = census_workspace.staging_test.id
  name         = var.destination_label
  type         = var.destination_type

  connection_config = var.destination_credentials
}

# Create a simple sync
resource "census_sync" "test_sync" {
  workspace_id = census_workspace.staging_test.id
  label        = var.sync_label

  source_attributes {
    connection_id = census_source.warehouse.id
    object {
      type       = "table"
      table_name = var.source_table
    }
  }

  destination_attributes {
    connection_id = census_destination.crm.id
    object        = var.destination_object
  }

  operation = "upsert"

  dynamic "field_mapping" {
    for_each = var.field_mapping
    content {
      from                  = field_mapping.value.from
      to                    = field_mapping.value.to
      is_primary_identifier = lookup(field_mapping.value, "is_primary_identifier", false)
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