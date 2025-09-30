terraform {
  required_providers {
    census = {
      source = "sutrolabs/census"
      version = "~> 0.1.0"
    }
  }
}

# Configure the Census Provider with your personal access token
provider "census" {
  personal_access_token = var.census_personal_token
  region               = var.census_region
}

# Create a test workspace
resource "census_workspace" "test" {
  name = "Terraform Test Workspace"
  
  # Configure notification emails for alerts
  notification_emails = [
    "data-alerts@company.com"
  ]
  
  # Return API key during creation for testing
  return_workspace_api_key = true
}

# Read information about the created workspace using data source
data "census_workspace" "test_data" {
  id = census_workspace.test.id
  
  # This data source depends on the resource being created first
  depends_on = [census_workspace.test]
}