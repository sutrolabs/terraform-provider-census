output "source_id" {
  description = "ID of the created source"
  value       = census_source.warehouse.id
}

output "activation_id" {
  description = "ID of the created activation"
  value       = census_activation.crm.id
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
  description = "Activations staging URL being used"
  value       = var.census_base_url
}