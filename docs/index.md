---
page_title: "Census Provider"
subcategory: ""
description: |-
  Terraform provider for managing Census resources
---

# Census Provider

The Census provider allows you to manage [Census](https://getcensus.com) resources using Terraform. Census enables you to sync data from your warehouse to all your operational tools, and this provider allows you to manage Census infrastructure as code.

## Example Usage

```terraform
terraform {
  required_providers {
    census = {
      source  = "sutrolabs/census"
      version = "~> 0.2.0"
    }
  }
}

provider "census" {
  personal_access_token = var.census_personal_token
  region                = "us"  # or "eu"
}
```

## Authentication

The Census provider uses Personal Access Tokens (PAT) for authentication. The provider automatically retrieves workspace-scoped tokens as needed for workspace-level operations.

To obtain a Personal Access Token:

1. Log in to your Census account
2. Navigate to Organization Settings > User Settings > Personal Access Tokens
3. Generate a new Personal Access Token
4. Store it securely (e.g., in environment variables or a secret manager)

## Multi-Region Support

Census operates in multiple regions. Specify your region when configuring the provider:

- `us` - United States (default)
- `eu` - European Union
- `au` - Australia

The provider automatically configures the appropriate API endpoints for your region.

## Schema

### Required

- `personal_access_token` (String, Sensitive) Personal Access Token for Census API authentication. Can also be set via the `CENSUS_PERSONAL_ACCESS_TOKEN` environment variable.

### Optional

- `region` (String) Census region: `us`, `eu`, or `au`. Defaults to `us`. Can also be set via the `CENSUS_REGION` environment variable.
- `base_url` (String) Custom base URL for the Census API. Primarily used for testing against staging environments. Can also be set via the `CENSUS_BASE_URL` environment variable.

## Resources

The Census provider supports the following resources:

- `census_workspace` - Manage Census workspaces
- `census_source` - Data warehouse connections (Snowflake, BigQuery, Postgres, Redshift, etc.)
- `census_destination` - Business tool integrations (Salesforce, HubSpot, etc.)
- `census_dataset` - SQL datasets for data transformation
- `census_sync` - Data syncs between sources and destinations

## Data Sources

All resources have corresponding data sources for read-only operations:

- `census_workspace`
- `census_source`
- `census_destination`
- `census_dataset`
- `census_sync`

For detailed documentation on each resource and data source, see the navigation menu.
