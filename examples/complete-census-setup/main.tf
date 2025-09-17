terraform {
  required_providers {
    census = {
      source = "your-org/census"
      version = "~> 0.1.0"
    }
  }
}

# Provider configuration
provider "census" {
  personal_access_token = var.census_personal_token  # For workspace management
  region               = var.census_region
  
  # Note: No single workspace_access_token - workspace-specific tokens are handled
  # per resource using the workspace API keys from the workspaces configuration
}

# ==============================================================================
# PHASE 1: WORKSPACE MANAGEMENT (✅ Available)
# ==============================================================================

# Create workspaces as Terraform resources
resource "census_workspace" "marketing_prod" {
  name                     = "Marketing Production"
  notification_emails      = ["marketing-data@company.com", "data-alerts@company.com"]
  return_workspace_api_key = true
}

resource "census_workspace" "marketing_staging" {
  name                     = "Marketing Staging"
  notification_emails      = ["marketing-dev@company.com"]
  return_workspace_api_key = true
}

resource "census_workspace" "revops_prod" {
  name                     = "Revenue Operations Production"
  notification_emails      = ["revops@company.com", "finance-alerts@company.com"]
  return_workspace_api_key = true
}

# Example: Reference existing workspaces with data sources
# Uncomment and configure these if you have existing workspaces to reference
# data "census_workspace" "existing_sales" {
#   id = "12345"  # Your existing workspace ID
# }
# 
# data "census_workspace" "existing_finance" {
#   id = "67890"  # Another existing workspace ID
# }

# ==============================================================================
# PHASE 2: WORKSPACE-SCOPED DATA CONNECTIONS (🚧 Updated!)
# ==============================================================================

# Create sources explicitly tied to specific workspaces
# Each source belongs to exactly one workspace
resource "census_source" "marketing_prod_warehouse" {
  workspace_id = census_workspace.marketing_prod.id
  
  name = "Marketing Production Data Warehouse"
  type = "postgres"
  
  connection_config   = var.postgres_warehouse_connection
  auto_refresh_tables = var.enable_auto_refresh
}

resource "census_source" "revops_prod_warehouse" {
  workspace_id = census_workspace.revops_prod.id
  
  name = "RevOps Production Data Warehouse"
  type = "postgres"
  
  connection_config   = var.postgres_warehouse_connection
  auto_refresh_tables = var.enable_auto_refresh
}

# Create destinations explicitly tied to specific workspaces
resource "census_destination" "marketing_prod_crm" {
  workspace_id = census_workspace.marketing_prod.id
  
  name = "Marketing Production CRM"
  type = "salesforce"
  
  connection_config    = var.salesforce_prod_connection
  auto_refresh_objects = var.enable_auto_refresh
}

resource "census_destination" "marketing_staging_crm" {
  workspace_id = census_workspace.marketing_staging.id
  
  name = "Marketing Staging CRM"
  type = "salesforce"
  
  connection_config    = var.salesforce_staging_connection
  auto_refresh_objects = var.enable_auto_refresh
}

# Example: Reference sources/destinations from existing workspaces
# Uncomment if you have existing workspaces with existing sources/destinations
# data "census_source" "existing_sales_warehouse" {
#   id = "98765"  # ID of existing source in existing workspace
# }
# 
# data "census_destination" "existing_finance_system" {
#   id = "54321"  # ID of existing destination in existing workspace
# }

# ==============================================================================
# PHASE 3: DATA SYNCS (🔄 Coming next!)
# ==============================================================================

# Sync configuration - will be added in next iteration
# resource "census_sync" "user_sync" {
#   name        = "User Data Sync"
#   source_id   = census_source.warehouse.id
#   destination_id = census_destination.crm.id
#   
#   # Sync configuration
#   source_object      = "users_table"
#   destination_object = "Contact"
#   
#   # Field mapping
#   field_mappings = [
#     {
#       source_field      = "email"
#       destination_field = "Email"
#       is_primary_key   = true
#     },
#     {
#       source_field      = "first_name" 
#       destination_field = "FirstName"
#     },
#     {
#       source_field      = "last_name"
#       destination_field = "LastName"
#     }
#   ]
#
#   # Sync settings
#   sync_mode = "update_or_create"
#   schedule  = "0 */6 * * *"  # Every 6 hours
# }

# ==============================================================================
# PHASE 4: ADVANCED FEATURES (📈 Future updates)
# ==============================================================================

# Dataset definitions, sync runs, webhooks, etc. will be added here
# as we implement more resources

# ==============================================================================
# OUTPUTS - Information about created resources
# ==============================================================================

output "workspaces_info" {
  description = "Information about all created workspaces"
  value = {
    marketing_prod = {
      id              = census_workspace.marketing_prod.id
      name            = census_workspace.marketing_prod.name
      organization_id = census_workspace.marketing_prod.organization_id
      created_at      = census_workspace.marketing_prod.created_at
      api_key         = census_workspace.marketing_prod.api_key
    }
    marketing_staging = {
      id              = census_workspace.marketing_staging.id
      name            = census_workspace.marketing_staging.name
      organization_id = census_workspace.marketing_staging.organization_id
      created_at      = census_workspace.marketing_staging.created_at
      api_key         = census_workspace.marketing_staging.api_key
    }
    revops_prod = {
      id              = census_workspace.revops_prod.id
      name            = census_workspace.revops_prod.name
      organization_id = census_workspace.revops_prod.organization_id
      created_at      = census_workspace.revops_prod.created_at
      api_key         = census_workspace.revops_prod.api_key
    }
  }
  sensitive = true
}

output "sources_info" {
  description = "Information about all created sources"
  value = {
    marketing_prod_warehouse = {
      id           = census_source.marketing_prod_warehouse.id
      name         = census_source.marketing_prod_warehouse.name
      type         = census_source.marketing_prod_warehouse.type
      status       = census_source.marketing_prod_warehouse.status
      created_at   = census_source.marketing_prod_warehouse.created_at
      workspace_id = census_source.marketing_prod_warehouse.workspace_id
    }
    revops_prod_warehouse = {
      id           = census_source.revops_prod_warehouse.id
      name         = census_source.revops_prod_warehouse.name
      type         = census_source.revops_prod_warehouse.type
      status       = census_source.revops_prod_warehouse.status
      created_at   = census_source.revops_prod_warehouse.created_at
      workspace_id = census_source.revops_prod_warehouse.workspace_id
    }
  }
}

output "destinations_info" {
  description = "Information about all created destinations"  
  value = {
    marketing_prod_crm = {
      id           = census_destination.marketing_prod_crm.id
      name         = census_destination.marketing_prod_crm.name
      type         = census_destination.marketing_prod_crm.type
      status       = census_destination.marketing_prod_crm.status
      created_at   = census_destination.marketing_prod_crm.created_at
      workspace_id = census_destination.marketing_prod_crm.workspace_id
    }
    marketing_staging_crm = {
      id           = census_destination.marketing_staging_crm.id
      name         = census_destination.marketing_staging_crm.name
      type         = census_destination.marketing_staging_crm.type
      status       = census_destination.marketing_staging_crm.status
      created_at   = census_destination.marketing_staging_crm.created_at
      workspace_id = census_destination.marketing_staging_crm.workspace_id
    }
  }
}

output "setup_summary" {
  description = "Summary of the complete workspace-scoped Census setup"
  sensitive   = true
  value = {
    workspaces_ready = {
      marketing_prod    = census_workspace.marketing_prod.id != ""
      marketing_staging = census_workspace.marketing_staging.id != ""
      revops_prod       = census_workspace.revops_prod.id != ""
    }
    sources_ready = {
      marketing_prod_warehouse = census_source.marketing_prod_warehouse.id != ""
      revops_prod_warehouse    = census_source.revops_prod_warehouse.id != ""
    }
    destinations_ready = {
      marketing_prod_crm    = census_destination.marketing_prod_crm.id != ""
      marketing_staging_crm = census_destination.marketing_staging_crm.id != ""
    }
    
    total_workspaces   = 3  # marketing_prod + marketing_staging + revops_prod
    total_sources      = 2  # marketing_prod_warehouse + revops_prod_warehouse
    total_destinations = 2  # marketing_prod_crm + marketing_staging_crm
    total_resources    = 7  # 3 workspaces + 2 sources + 2 destinations
    
    workspace_structure = [
      "marketing_prod: ${census_workspace.marketing_prod.name} (ID: ${census_workspace.marketing_prod.id})",
      "marketing_staging: ${census_workspace.marketing_staging.name} (ID: ${census_workspace.marketing_staging.id})",
      "revops_prod: ${census_workspace.revops_prod.name} (ID: ${census_workspace.revops_prod.id})"
    ]
    
    resource_relationships = [
      "✅ Each source/destination explicitly belongs to one workspace",
      "✅ No ambiguous global configurations",
      "✅ Clear workspace isolation and security boundaries",
      "✅ Proper Terraform resource relationships"
    ]
    
    authentication_model = [
      "Personal Access Token (PAT): Used for workspace management operations",
      "Each workspace has its own API key stored in Terraform state",
      "Sources and destinations are scoped to specific workspaces",
      "Use data sources to reference existing workspaces/sources/destinations"
    ]
    
    terraform_patterns = [
      "✅ Resources for managing Census objects with Terraform",
      "✅ Data sources for referencing existing Census objects",
      "✅ Explicit workspace_id parameters for clear relationships",
      "✅ Standard Terraform import patterns for existing resources"
    ]
    
    next_steps = [
      "Verify all workspace connections in Census dashboard",
      "Use 'terraform import' to bring existing workspaces under Terraform management",
      "Add data sources to reference existing sources/destinations if needed",
      "Configure workspace-specific sync rules (coming in next update)"
    ]
  }
}