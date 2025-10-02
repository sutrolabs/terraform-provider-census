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
  label        = var.source_label
  type         = var.source_type
  credentials  = var.source_credentials
}

# Create a destination (e.g., Salesforce sandbox)
resource "census_destination" "crm" {
  workspace_id = census_workspace.staging_test.id
  label        = var.destination_label
  type         = var.destination_type
  credentials  = var.destination_credentials
}

# Create a simple sync
resource "census_sync" "test_sync" {
  workspace_id   = census_workspace.staging_test.id
  label          = var.sync_label

  source_type = "table"
  source_attributes = {
    connection_id  = census_source.warehouse.id
    object         = var.source_table
    full_sync_mode = "replace"
  }

  destination_object = var.destination_object

  field_mapping = var.field_mapping

  schedule = {
    frequency = var.sync_frequency
    hour      = var.sync_hour
    minute    = var.sync_minute
  }

  paused = var.sync_paused
}