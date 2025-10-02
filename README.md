# Terraform Provider for Census

A Terraform provider for managing [Census](https://getcensus.com) resources. Census enables you to sync data from your warehouse to all your operational tools, and this provider allows you to manage Census infrastructure as code.

## Features

- **Multi-region support**: Works with US, EU, and AU Census regions
- **Complete Census workflow**: Manage workspaces, sources, datasets, destinations, and syncs
- **PAT-only authentication**: Uses personal access tokens with dynamic workspace token retrieval
- **Import support**: Import existing Census resources into Terraform state
- **Staging environment support**: Configure custom base URLs for testing

## Installation

```hcl
terraform {
  required_providers {
    census = {
      source  = "sutrolabs/census"
      version = "~> 0.1.1"
    }
  }
}

provider "census" {
  personal_access_token = var.census_personal_token
  region                = "us"  # or "eu"
}
```

## Usage

```hcl
resource "census_workspace" "data_team" {
  name = "Data Team Workspace"
  notification_emails = ["data-alerts@company.com"]
}

resource "census_source" "warehouse" {
  workspace_id = census_workspace.data_team.id
  name         = "Production Warehouse"
  type         = "snowflake"

  connection_config = {
    account   = "xy12345.us-east-1"
    database  = "ANALYTICS"
    warehouse = "COMPUTE_WH"
    role      = "CENSUS_ROLE"
    username  = "census_user"
    password  = var.snowflake_password
  }
}

resource "census_sync" "users_to_crm" {
  workspace_id = census_workspace.data_team.id
  label        = "Users to CRM"


  source_attributes {
    connection_id = census_source.warehouse.id
    object {
      type         = "table"
      table_name   = "users"
      table_schema = "public"
    }
  }

  destination_attributes {
    connection_id = census_destination.crm.id
    object        = "Contact"
  }

  operation = "upsert"
  sync_key  = ["email"]

  field_mapping {
    from      = "email"
    to        = "Email"
    operation = "direct"
  }

  schedule {
    frequency = "daily"
    hour      = 8
    timezone  = "UTC"
  }
}
```

## Resources

- `census_workspace` - Manage Census workspaces
- `census_source` - Data warehouse connections (Snowflake, BigQuery, Postgres, etc.)
- `census_destination` - Business tool integrations (Salesforce, HubSpot, etc.)
- `census_dataset` - SQL datasets for data transformation
- `census_sync` - Data syncs between sources and destinations

## Data Sources

All resources have corresponding data sources for read-only operations. See [documentation](docs/) for details.

## Documentation

- [Resource Documentation](docs/resources/) - Detailed documentation for each resource
- [Data Source Documentation](docs/data-sources/) - Read-only data source documentation
- [Examples](examples/) - Complete working examples
- [CHANGELOG](CHANGELOG.md) - Version history and changes
- [Census API Documentation](https://developers.getcensus.com/api-reference/introduction/overview)

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## Support

- [GitHub Issues](https://github.com/sutrolabs/terraform-provider-census/issues)
- [Census Documentation](https://docs.getcensus.com/)

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

---

**Note**: This provider is not officially maintained by Census. It is a community project designed to provide Terraform integration with Census services.
