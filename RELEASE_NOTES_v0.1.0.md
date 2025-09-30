# Release v0.1.0 - Initial Release

## 🎉 First Release - Complete Census Workflow

This is the initial release of the Census Terraform Provider for Sutro Labs. The provider enables infrastructure-as-code management of the complete Census data sync pipeline.

## ✨ Features

### Resources

All resources support full CRUD operations, import, and state management:

- **`census_workspace`** - Manage Census workspaces
  - Notification email configuration
  - Multi-region support (US/EU)
  - API key retrieval on creation

- **`census_source`** - Data warehouse connections
  - Support for Snowflake, BigQuery, Postgres, Redshift, and more
  - Dynamic credential validation against Census API
  - Auto table refresh after creation

- **`census_destination`** - Business tool integrations
  - Support for Salesforce, HubSpot, Intercom, and 100+ destinations
  - Real-time connection testing
  - Metadata auto-refresh

- **`census_dataset`** - SQL data transformation
  - Multi-line SQL queries with heredoc syntax
  - Column schema discovery
  - Cached record count tracking

- **`census_sync`** - Data synchronization
  - Field mappings (direct, hash, constant operations)
  - Flexible scheduling (hourly, daily, weekly, manual)
  - Multiple sync modes (upsert, append, mirror)
  - Support for all source types

### Data Sources

Read-only data sources for all resources:
- `census_workspace`
- `census_source`
- `census_destination`
- `census_dataset`
- `census_sync`

### Authentication

- **PAT-only authentication** - No need for workspace tokens
- Automatic workspace token retrieval
- Multi-region support (US/EU)
- Environment variable support

### Technical Highlights

- Built with terraform-plugin-sdk/v2
- OpenAPI specification compliance
- TypeSet-based collections (prevents order drift)
- Comprehensive error handling
- Import support for all resources
- Go 1.21+ compatibility

## 📦 Installation

### For Sutro Labs Engineers

See [INTERNAL_INSTALLATION.md](INTERNAL_INSTALLATION.md) for detailed setup instructions.

**Quick install:**

1. Download the binary for your platform below
2. Create plugin directory: `mkdir -p ~/.terraform.d/plugins/registry.terraform.io/sutrolabs/census/0.1.0/darwin_arm64`
3. Move binary: `mv terraform-provider-census_darwin_arm64 ~/.terraform.d/plugins/registry.terraform.io/sutrolabs/census/0.1.0/darwin_arm64/terraform-provider-census_v0.1.0`
4. Make executable: `chmod +x ~/.terraform.d/plugins/registry.terraform.io/sutrolabs/census/0.1.0/darwin_arm64/terraform-provider-census_v0.1.0`

**Platform Paths:**
- macOS Apple Silicon: `darwin_arm64`
- macOS Intel: `darwin_amd64`
- Linux: `linux_amd64`

## 📚 Documentation

- **Resources**: See `docs/resources/` for detailed documentation
- **Data Sources**: See `docs/data-sources/` for read-only operations
- **Examples**: Check `examples/complete-census-setup/` for a full working example

## 🚀 Quick Start

```hcl
terraform {
  required_providers {
    census = {
      source  = "sutrolabs/census"
      version = "0.1.0"
    }
  }
}

provider "census" {
  personal_access_token = var.census_personal_token
  region               = "us"
}

resource "census_workspace" "prod" {
  name = "Production Workspace"
  notification_emails = ["data-team@sutrolabs.com"]
}

resource "census_source" "warehouse" {
  workspace_id = census_workspace.prod.id
  name         = "Snowflake Warehouse"
  type         = "snowflake"
  credentials  = jsonencode({
    account   = "abc12345.us-east-1"
    warehouse = "COMPUTE_WH"
    database  = "PROD"
    username  = "census_user"
    password  = var.snowflake_password
  })
}

resource "census_destination" "salesforce" {
  workspace_id = census_workspace.prod.id
  name         = "Production Salesforce"
  type         = "salesforce"
  credentials  = jsonencode({
    username       = "census@company.com"
    password       = var.salesforce_password
    security_token = var.salesforce_token
    instance_url   = "https://company.my.salesforce.com"
  })
}

resource "census_sync" "contacts" {
  workspace_id   = census_workspace.prod.id
  name           = "Contacts to Salesforce"
  source_id      = census_source.warehouse.id
  destination_id = census_destination.salesforce.id

  source_attributes = jsonencode({
    connection_id = census_source.warehouse.id
    object = {
      type       = "table"
      table_name = "customers"
    }
  })

  destination_object = "Contact"

  field_mappings = [
    {
      from      = "email"
      to        = "Email"
      operation = "direct"
    },
    {
      from      = "first_name"
      to        = "FirstName"
      operation = "direct"
    },
  ]

  operation = "upsert"

  trigger = jsonencode({
    schedule = {
      frequency = "daily"
      hour      = 8
      minute    = 0
    }
  })
}
```

## 🧪 Tested Integrations

- Snowflake sources
- Salesforce destinations
- Table-based syncs
- Field mappings
- Daily scheduling

## 📝 Known Limitations

- Integration tests require mock Census API server (planned)
- Some advanced sync features may need additional testing
- Documentation covers core use cases (advanced scenarios coming)

## 🔜 Future Roadmap

- Sync run operations (execute and monitor syncs)
- Webhook management
- Enhanced testing suite
- Public Terraform Registry publication

## 🐛 Bug Reports

Please report issues in the GitHub repository or reach out in #eng-data-platform Slack channel.

## 📄 License

MIT License - Copyright (c) 2025 Sutro Labs

---

**Built by**: Data Platform Team at Sutro Labs
**Release Date**: September 29, 2025