# Complete Activations Setup Example

This example demonstrates how to use the Activations Terraform provider to manage a complete Activations setup.

## Features

- ✅ **Mixed Management**: Combine Terraform-created and existing activations
- ✅ **Multi-Environment Support**: Production, staging, development, or any custom environments
- ✅ **Environment-Specific Configurations**: Different connections for staging vs production

## Quick Start

### 1. Prerequisites

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- Census account with appropriate permissions
- Census Personal Access Token

### 2. Get Your Census Token

#### Personal Access Token (PAT)
1. Go to Activations dashboard → Settings → Developer → Personal Access Tokens
2. Create a new token
3. Copy the token (starts with `census_pat_...`)

### 3. Configure Your Setup

```bash
# Clone or copy this example
cp terraform.tfvars.example terraform.tfvars

# Edit terraform.tfvars with your tokens and configuration
vim terraform.tfvars
```

### 4. Deploy

```bash
# Build the provider (if using development version)
cd ../../ && make dev

# Return to example directory
cd examples/complete-census-setup/

# Initialize and deploy
terraform init
terraform plan
terraform apply
```

## Configuration Options

### Basic Configuration

The minimal required configuration in `terraform.tfvars`:

```hcl
# Authentication
census_personal_token = "census_pat_your_personal_token_here"

# Data connections
source_type = "postgres"  # or snowflake, bigquery, etc.
source_connection = {
  host     = "your-database.amazonaws.com"
  port     = 5432  # Numbers work directly, no need for quotes
  username = "census_user"
  password = "your-password"
  database = "analytics"
}

destination_type = "salesforce"  # or hubspot, postgres, etc.
destination_connection = {
  username       = "your-sf-user@company.com"
  password       = "your-sf-password"
  security_token = "your-sf-security-token"
  sandbox        = false  # Booleans work directly!
}

# Auto-refresh metadata after creating/updating connections
enable_auto_refresh = true  # Set to false to manually refresh schemas in Activations UI
```

### Environment-Specific Connections

Configure different connections for staging environments:

```hcl
# Production database
source_connection = {
  host     = "prod-db.company.com"
  port     = 5432
  username = "census_prod"
  password = "prod-password"
  database = "analytics"
}

# Staging database (optional - falls back to production config if not set)
staging_source_connection = {
  host     = "staging-db.company.com"
  port     = 5432
  username = "census_staging"
  password = "staging-password"
  database = "staging_analytics"
}

# Production Salesforce
destination_connection = {
  username       = "census@company.com"
  password       = "prod-password"
  security_token = "prod-token"
  sandbox        = false
}

# Staging Salesforce (optional)
staging_destination_connection = {
  username       = "census@company.com.staging"
  password       = "staging-password"
  security_token = "staging-token"
  sandbox        = true
}
```

## Supported Source Types

- `snowflake`
- `bigquery`
- `postgres`
- `redshift`
- `databricks`
- `mysql`
- `sql_server`
- `oracle`
- `clickhouse`

## Supported Destination Types

- `salesforce`
- `hubspot`
- `postgres`
- `airtable`
- `intercom`
- `zendesk`
- `mailchimp`
- `segment`
- `mixpanel`
- `amplitude`

## Common Connection Examples

### Snowflake (Password Authentication)
```hcl
source_type = "snowflake"
source_connection = {
  account   = "your-account.snowflakecomputing.com"
  username  = "census_user"
  password  = "your-password"
  warehouse = "COMPUTE_WH"
  database  = "ANALYTICS"
  schema    = "PUBLIC"
}
```

### Snowflake (Keypair Authentication)
```hcl
source_type = "snowflake"
source_connection = {
  account                = "abc12345.us-east-1"
  warehouse              = "COMPUTE_WH"
  database               = "PRODUCTION"
  username               = "census_user"
  role                   = "CENSUS_ROLE"
  use_keypair            = true  # Boolean works directly!
  private_key_pkcs8      = var.snowflake_private_key
  private_key_passphrase = var.snowflake_key_passphrase  # Optional
}
```

### Snowflake (Keypair Authentication, Write-Only Secrets — Terraform 1.11+)
Keep secret payloads out of Terraform state/plan by passing them through the
write-only `connection_config_wo` argument on `census_source` — a `jsonencode`-d
object of secret credentials that works for any source type. Bump
`connection_config_wo_version` whenever you rotate a secret so Terraform
re-applies it.
```hcl
resource "census_source" "warehouse" {
  name         = "Production Warehouse (Keypair)"
  type         = "snowflake"

  connection_config = {
    account     = "abc12345.us-east-1"
    warehouse   = "COMPUTE_WH"
    database    = "PRODUCTION"
    username    = "census_user"
    role        = "CENSUS_ROLE"
    use_keypair = true
  }

  connection_config_wo = jsonencode({
    private_key_pkcs8      = var.snowflake_private_key
    private_key_passphrase = var.snowflake_key_passphrase # Optional
  })
  connection_config_wo_version = 1
}
```

### BigQuery
```hcl
source_type = "bigquery"
source_connection = {
  project_id          = "your-gcp-project-id"
  service_account_key = jsonencode(var.bigquery_service_account_key)
  client_email        = "census@your-project.iam.gserviceaccount.com"
}
```

### HubSpot
```hcl
destination_type = "hubspot"
destination_connection = {
  access_token = "your-hubspot-access-token"
}
```

## Troubleshooting

### Common Issues

**Provider not found**
```bash
# Build the development provider
cd ../../ && make dev
cd examples/complete-census-setup/
terraform init
```

**Token permissions errors**
- Ensure the Personal Access Token has appropriate permissions

**Connection test failures**
- Verify database/service credentials are correct
- Check network connectivity and firewall rules
- For staging environments, ensure sandbox settings are correct

### Validation

```bash
# Validate configuration syntax
terraform validate

# Plan without applying to check for issues
terraform plan
```

### Outputs

After successful deployment, you'll see outputs with:
- **sources_info**: Information about created data sources
- **activations_info**: Information about created activations  
- **setup_summary**: Overview of the complete setup with next steps

## Next Steps

1. **Verify Connections**: Check Activations dashboard to ensure all sources and activations are working
2. **Configure Syncs**: Set up data syncs between your sources and destinations
3. **Monitor**: Set up monitoring and alerting for your data pipelines
4. **Scale**: Add more sources and activations as needed

## Support

- **Provider Issues**: [Activations Support](mailto:activation-support@fivetran.com)
- **Activations documentation**: [Activations docs](https://fivetran.com/docs/activations)
- **Terraform Documentation**: [Terraform Docs](https://www.terraform.io/docs/)

## License

This example is provided as-is for demonstration purposes.