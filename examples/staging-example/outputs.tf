output "workspace_id" {
  description = "ID of the created workspace"
  value       = census_workspace.staging_test.id
}

output "workspace_name" {
  description = "Name of the created workspace"
  value       = census_workspace.staging_test.name
}

output "source_id" {
  description = "ID of the created data source"
  value       = census_source.warehouse.id
}

output "destination_id" {
  description = "ID of the created destination"
  value       = census_destination.crm.id
}

output "sync_id" {
  description = "ID of the created sync"
  value       = census_sync.test_sync.id
}

output "sync_status" {
  description = "Current status of the sync"
  value       = census_sync.test_sync.paused ? "paused" : "active"
}

output "staging_url" {
  description = "Census staging URL being used"
  value       = var.census_base_url
}