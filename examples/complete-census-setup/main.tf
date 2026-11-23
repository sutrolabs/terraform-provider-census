terraform {
  required_providers {
    census = {
      source  = "sutrolabs/census"
      version = "~> 2.0"
    }
  }
}

# Provider configuration
provider "census" {
  personal_access_token = var.census_personal_token
}

# ==============================================================================
# PHASE 2: DATA CONNECTIONS
# ==============================================================================

# Create sources

resource "census_source" "marketing_prod_postgres" {

  name = "Marketing Production Postgres Warehouse"
  type = "postgres"

  connection_config   = var.postgres_warehouse_connection
  auto_refresh_tables = var.enable_auto_refresh
}

resource "census_source" "revops_prod_postgres" {

  name = "RevOps Production Postgres Warehouse"
  type = "postgres"

  connection_config   = var.postgres_warehouse_connection
  auto_refresh_tables = var.enable_auto_refresh
}

resource "census_source" "marketing_staging_postgres" {

  name = "Marketing Staging Postgres Warehouse"
  type = "postgres"

  connection_config   = var.postgres_warehouse_connection
  auto_refresh_tables = var.enable_auto_refresh
}

resource "census_source" "marketing_staging_warehouse" {

  name = "Marketing Staging Redshift Warehouse"
  type = "redshift"

  connection_config   = var.redshift_warehouse_connection
  auto_refresh_tables = var.enable_auto_refresh
}

resource "census_source" "marketing_prod_warehouse" {

  name = "Marketing Production Redshift Warehouse"
  type = "redshift"

  connection_config   = var.redshift_warehouse_connection
  auto_refresh_tables = var.enable_auto_refresh
}

resource "census_source" "revops_prod_warehouse" {

  name = "RevOps Production Redshift Warehouse"
  type = "redshift"

  connection_config   = var.redshift_warehouse_connection
  auto_refresh_tables = var.enable_auto_refresh
}

resource "census_source" "marketing_prod_snowflake" {

  name = "Marketing Production Snowflake Warehouse"
  type = "snowflake"

  connection_config   = var.snowflake_warehouse_connection
  auto_refresh_tables = var.enable_auto_refresh
}

# Create activations, each pairing a source with a destination
resource "census_activation" "marketing_prod_crm" {
  group_id = census_source.marketing_prod_warehouse.group_id

  name = "Marketing Production CRM"
  type = "salesforce"

  credentials = var.salesforce_prod_connection
}

resource "census_activation" "marketing_staging_crm" {
  group_id = census_source.marketing_staging_warehouse.group_id

  name = "Marketing Staging CRM"
  type = "salesforce"

  credentials = var.salesforce_staging_connection
}

# Example: Reference existing sources/activations
# Uncomment if you have existing sources/activations
# data "census_source" "existing_sales_warehouse" {
#   id = "98765"  # ID of existing source
# }
#
# data "census_activation" "existing_finance_system" {
#   id = "54321"  # ID of existing activation
# }

# ==============================================================================
# PHASE 2.5: SQL DATASETS (✅ Available!)
# ==============================================================================

# Create SQL datasets for data transformation and modeling
# Datasets allow you to define custom SQL queries against your sources

resource "census_dataset" "active_users" {
  name        = "Active Users"
  type        = "sql"
  description = "Filtered dataset of active users for marketing campaigns"

  group_id = census_source.marketing_prod_warehouse.group_id

  query = <<-SQL
    SELECT
      id,
      email,
      first_name,
      last_name,
      created_at,
      last_login_at
    FROM users
    WHERE active = true
      AND last_login_at > CURRENT_DATE - INTERVAL '30 days'
  SQL
}

resource "census_dataset" "high_value_customers" {
  name        = "High Value Customers"
  type        = "sql"
  description = "Customers with lifetime value > $1000"

  group_id = census_source.marketing_prod_warehouse.group_id

  query = <<-SQL
    SELECT
      u.id,
      u.email,
      u.first_name,
      u.last_name,
      SUM(o.amount) as lifetime_value,
      COUNT(o.id) as order_count
    FROM users u
    JOIN orders o ON u.id = o.user_id
    GROUP BY u.id, u.email, u.first_name, u.last_name
    HAVING SUM(o.amount) > 1000
  SQL
}

resource "census_dataset" "all_users" {
  name        = "All Users"
  type        = "sql"
  description = "Simple dataset with all user data for syncing"

  group_id = census_source.marketing_prod_warehouse.group_id

  query = <<-SQL
    SELECT * FROM users
  SQL
}

# Example: Use data source to reference an existing dataset
# data "census_dataset" "existing_dataset" {
#   id = "123"
# }

# ==============================================================================
# PHASE 3: DATA SYNCS (✅ Available!)
# ==============================================================================

# Create syncs that connect sources to destinations through an activation
# Each sync defines how data flows from a source to a destination

resource "census_sync" "marketing_contact_sync" {
  activation_id = census_activation.marketing_prod_crm.id
  label         = "REAL Marketing Contact Sync"

  # Source configuration - use a table from the warehouse
  source_attributes {
    object {
      type          = "table"
      table_name    = "users"
      table_schema  = "public"
      table_catalog = "dev" # Use database name as catalog
    }
  }

  # Destination configuration - specify Salesforce object
  destination_attributes {
    object = "Contact"
  }

  # Operation mode for the sync
  operation = "upsert"

  # Field mappings between source and destination
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
    from = "id"
    to   = "Census_ID__c"
  }

  field_mapping {
    from = "last_name"
    to   = "LastName"
  }

  # Run mode - triggered on daily schedule at 6 AM UTC
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

  # Alert configurations for production sync monitoring
  alert {
    type                 = "FailureAlertConfiguration"
    send_for             = "first_time"
    should_send_recovery = true
    options              = {}
  }

  alert {
    type                 = "InvalidRecordPercentAlertConfiguration"
    send_for             = "first_time"
    should_send_recovery = true
    options = {
      threshold = "75"
    }
  }

  # Start paused until ready to go live
  paused = true
}

resource "census_sync" "marketing_contact_sync_2" {
  activation_id = census_activation.marketing_prod_crm.id
  label         = "REAL Marketing Contact Sync 2"

  # Source configuration - use a table from the warehouse
  source_attributes {
    object {
      type          = "table"
      table_name    = "users"
      table_schema  = "public"
      table_catalog = "dev" # Use database name as catalog
    }
  }

  # Destination configuration - specify Salesforce object
  destination_attributes {
    object = "Contact"
  }

  # Operation mode for the sync
  operation = "upsert"

  alert {
    type                 = "FailureAlertConfiguration"
    send_for             = "first_time"
    should_send_recovery = true
    options              = {}
  }

  alert {
    type                 = "InvalidRecordPercentAlertConfiguration"
    send_for             = "first_time"
    should_send_recovery = true
    options = {
      threshold = "75"
    }
  }


  # Field mappings between source and destination
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
    from = "id"
    to   = "Census_ID__c"
  }

  field_mapping {
    from = "last_name"
    to   = "LastName"
  }

  # Run mode - triggered on daily schedule at 6 AM UTC
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

  # Start paused until ready to go live
  paused = true
}

resource "census_sync" "marketing_contact_sync_3" {
  activation_id = census_activation.marketing_prod_crm.id
  label         = "Marketing Contact Sync 3"

  # Source configuration - use a table from the warehouse
  source_attributes {
    object {
      type          = "table"
      table_name    = "users"
      table_schema  = "public"
      table_catalog = "dev" # Use database name as catalog
    }
  }

  # Destination configuration - specify Salesforce object
  destination_attributes {
    object = "Contact"
  }

  # Operation mode for the sync
  operation = "upsert"

  # Field mappings between source and destination
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
    from = "id"
    to   = "Census_ID__c"
  }


  field_mapping {
    from = "last_name"
    to   = "LastName"
  }
  alert {
    type                 = "FailureAlertConfiguration"
    send_for             = "first_time"
    should_send_recovery = true
    options              = {}
  }

  alert {
    type                 = "InvalidRecordPercentAlertConfiguration"
    send_for             = "first_time"
    should_send_recovery = true
    options = {
      threshold = "75"
    }
  }


  # Run mode - triggered on daily schedule at 6 AM UTC
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

  # Start paused until ready to go live
  paused = true
}

# Sync using a dataset as the source
resource "census_sync" "dataset_contact_sync" {
  activation_id = census_activation.marketing_prod_crm.id
  label         = "Dataset to Contacts Syncer"
  paused        = true

  # Source configuration - use a dataset instead of table
  source_attributes {
    object {
      type = "dataset"
      id   = census_dataset.all_users.id
    }
  }

  # Destination configuration - Salesforce Contacts
  destination_attributes {
    object = "Contact"
  }

  operation = "upsert"

  advanced_configuration = jsonencode({
    bulk_id_lookup = true
  })

  field_mapping {
    type              = "sync_metadata"
    sync_metadata_key = "sync_run_id"
    to                = "another__c"
  }

  field_mapping {
    from = "first_name"
    to   = "FirstName"
  }

  field_mapping {
    type            = "liquid_template"
    liquid_template = "{{ record['_fivetran_synced'] | upcase }}"
    to              = "OtherPostalCode"
  }

  field_mapping {
    from = "last_name"
    to   = "LastName"
  }

  field_mapping {
    type                = "segment_membership"
    segment_identify_by = "name"
    to                  = "MailingStreet"
  }

  field_mapping {
    from                  = "email"
    to                    = "Email"
    is_primary_identifier = true
  }

  field_mapping {
    from = "user_id"
    to   = "Census_ID__c"
  }

  field_mapping {
    type     = "constant"
    constant = "HERE is my constant value"
    to       = "AssistantName"
  }

  field_mapping {
    type     = "constant"
    constant = "HERE is my constant value 2"
    to       = "Title"
  }

  # Run mode - triggered hourly at minute 10
  run_mode {
    type = "triggered"
    triggers {
      schedule {
        frequency = "hourly"
        minute    = 10
      }
    }
  }

  # Comprehensive alerting for high-frequency sync
  alert {
    type                 = "FailureAlertConfiguration"
    send_for             = "first_time"
    should_send_recovery = true
    options              = {}
  }

  alert {
    type                 = "RuntimeAlertConfiguration"
    send_for             = "first_time"
    should_send_recovery = false
    options = {
      threshold  = "45"
      unit       = "minutes"
      start_type = "actual"
    }
  }

  alert {
    type                 = "InvalidRecordPercentAlertConfiguration"
    send_for             = "first_time"
    should_send_recovery = true
    options = {
      threshold = "75"
    }
  }

  alert {
    type                 = "StatusAlertConfiguration"
    send_for             = "every_time"
    should_send_recovery = false
    options = {
      status_name = "completed"
    }
  }
}

# resource "census_sync" "marketing_staging_test_sync" {
#   activation_id = census_activation.marketing_staging_crm.id
#   label         = "Marketing Staging Test Sync"


#   # Test sync with limited data  
#   source_attributes {
#     object {
#       type          = "table"
#       table_name    = "users"
#       table_schema  = "public"
#       table_catalog = "dev" # Use database name as catalog
#     }
#   }

#   destination_attributes {
#     object = "Contact"
#   }

#   # Operation mode for the sync
#   operation = "upsert"

#   field_mapping {
#     from                  = "email"
#     to                    = "Email"
#     is_primary_identifier = true
#   }

#   field_mapping {
#     from = "first_name"
#     to   = "FirstName"
#   }

#   field_mapping {
#     from = "last_name"
#     to   = "LastName"
#   }


#   # Manual sync for testing
#   run_mode {
#     type = "triggered"
#     triggers {
#       schedule {
#         frequency = "manual"
#       }
#     }
#   }

#   paused = false # Ready for testing
# }

# # RevOps activation: the production CRM synced from the RevOps warehouse
# resource "census_activation" "revops_prod_crm" {
#   group_id = census_source.revops_prod_warehouse.group_id
#
#   name = "RevOps Production CRM"
#   type = "salesforce"
#
#   credentials = var.salesforce_prod_connection
# }
#
# # RevOps sync for account data
# resource "census_sync" "revops_account_sync" {
#   activation_id = census_activation.revops_prod_crm.id
#   label         = "RevOps Account Sync"


#   # Source configuration - account enrichment data
#   source_attributes {
#     object {
#       type          = "table"
#       table_name    = "users"
#       table_schema  = "public"
#       table_catalog = "dev" # Use database name as catalog
#     }
#   }

#   # Destination configuration
#   destination_attributes {
#     object = "Account"
#   }

#   # Operation mode for the sync
#   operation = "upsert"

#   # Field mappings for account data
#   field_mapping {
#     from                  = "company_domain"
#     to                    = "Website"
#     is_primary_identifier = true
#   }

#   field_mapping {
#     from = "company_name"
#     to   = "Name"
#   }

#   field_mapping {
#     from = "industry"
#     to   = "Industry"
#   }

#   field_mapping {
#     from = "employee_count"
#     to   = "NumberOfEmployees"
#   }

#   field_mapping {
#     from = "annual_revenue"
#     to   = "AnnualRevenue"
#   }

#   field_mapping {
#     from = "country"
#     to   = "BillingCountry"
#   }


#   # Run weekly on Sundays at 2 AM
#   run_mode {
#     type = "triggered"
#     triggers {
#       schedule {
#         frequency = "weekly"
#         day       = "Sunday"
#         hour      = 2
#         minute    = 0
#       }
#     }
#   }

#   paused = true # Start paused for review
# }

# Example of using data source to reference an existing sync
# data "census_sync" "existing_sync" {
#   id = "123456"
# }

# ==============================================================================
# PHASE 4: ADVANCED FEATURES (📈 Future updates)
# ==============================================================================

# Dataset definitions, sync runs, webhooks, etc. will be added here
# as we implement more resources

# ==============================================================================
# OUTPUTS - Information about created resources
# ==============================================================================

output "sources_info" {
  description = "Information about all created sources"
  value = {
    marketing_prod_warehouse = {
      id         = census_source.marketing_prod_warehouse.id
      name       = census_source.marketing_prod_warehouse.name
      type       = census_source.marketing_prod_warehouse.type
      group_id   = census_source.marketing_prod_warehouse.group_id
      created_at = census_source.marketing_prod_warehouse.created_at
    }
    revops_prod_warehouse = {
      id         = census_source.revops_prod_warehouse.id
      name       = census_source.revops_prod_warehouse.name
      type       = census_source.revops_prod_warehouse.type
      group_id   = census_source.revops_prod_warehouse.group_id
      created_at = census_source.revops_prod_warehouse.created_at
    }
  }
}

output "activations_info" {
  description = "Information about all created activations"
  value = {
    marketing_prod_crm = {
      id                  = census_activation.marketing_prod_crm.id
      name                = census_activation.marketing_prod_crm.name
      type                = census_activation.marketing_prod_crm.type
      group_id            = census_activation.marketing_prod_crm.group_id
      last_test_succeeded = census_activation.marketing_prod_crm.last_test_succeeded
    }
    marketing_staging_crm = {
      id                  = census_activation.marketing_staging_crm.id
      name                = census_activation.marketing_staging_crm.name
      type                = census_activation.marketing_staging_crm.type
      group_id            = census_activation.marketing_staging_crm.group_id
      last_test_succeeded = census_activation.marketing_staging_crm.last_test_succeeded
    }
  }
}

output "setup_summary" {
  description = "Summary of the complete Activations setup"
  sensitive   = true
  value = {
    sources_ready = {
      marketing_prod_warehouse = census_source.marketing_prod_warehouse.id != ""
      revops_prod_warehouse    = census_source.revops_prod_warehouse.id != ""
    }
    activations_ready = {
      marketing_prod_crm    = census_activation.marketing_prod_crm.id != ""
      marketing_staging_crm = census_activation.marketing_staging_crm.id != ""
    }

    total_sources     = 2 # marketing_prod_warehouse + revops_prod_warehouse
    total_activations = 2 # marketing_prod_crm + marketing_staging_crm
    total_resources   = 4 # 2 sources + 2 activations

    resource_relationships = [
      "✅ Each activation pairs one source with one destination",
      "✅ Each sync belongs to one activation",
      "✅ Proper Terraform resource relationships"
    ]

    authentication_model = [
      "Personal Access Token (PAT): Used for all operations",
      "Use data sources to reference existing sources/activations"
    ]

    terraform_patterns = [
      "✅ Resources for managing Activations objects with Terraform",
      "✅ Data sources for referencing existing Activations objects",
      "✅ Explicit group_id and activation_id parameters for clear relationships",
      "✅ Standard Terraform import patterns for existing resources"
    ]

    next_steps = [
      "Verify all connections in the Activations dashboard",
      "Use 'terraform import' to bring existing activations under Terraform management",
      "Add data sources to reference existing sources/activations if needed"
    ]
  }
}
