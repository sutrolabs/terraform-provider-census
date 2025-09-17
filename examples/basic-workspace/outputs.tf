output "workspace_id" {
  description = "ID of the created workspace"
  value       = census_workspace.test.id
}

output "workspace_name" {
  description = "Name of the created workspace"
  value       = census_workspace.test.name
}

output "workspace_organization_id" {
  description = "Organization ID that owns this workspace"
  value       = census_workspace.test.organization_id
}

output "workspace_created_at" {
  description = "Timestamp when the workspace was created"
  value       = census_workspace.test.created_at
}

output "workspace_notification_emails" {
  description = "Email addresses configured for notifications"
  value       = census_workspace.test.notification_emails
}

output "workspace_api_key" {
  description = "API key for the workspace (sensitive - only shown if return_api_key is true)"
  value       = census_workspace.test.api_key
  sensitive   = true
}

# Outputs from data source (should match resource outputs)
output "data_source_workspace_name" {
  description = "Workspace name from data source (should match resource)"
  value       = data.census_workspace.test_data.name
}

output "data_source_organization_id" {
  description = "Organization ID from data source (should match resource)"
  value       = data.census_workspace.test_data.organization_id
}

# Validation - these should match
output "names_match" {
  description = "Whether resource and data source names match"
  value       = census_workspace.test.name == data.census_workspace.test_data.name
}

output "org_ids_match" {
  description = "Whether resource and data source organization IDs match"
  value       = census_workspace.test.organization_id == data.census_workspace.test_data.organization_id
}