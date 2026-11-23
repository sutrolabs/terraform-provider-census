---
page_title: "Fivetran Activations Provider"
subcategory: ""
description: |-
  Terraform provider for managing Activations resources
---

# Fivetran Activations Provider

The Activations provider allows you to manage [Fivetran Activations](https://www.fivetran.com/data-movement/activations) resources using Terraform. Activations enables you to sync data from your warehouse to all your operational tools, and this provider allows you to manage Activations infrastructure as code.

## Example Usage

```terraform
terraform {
  required_providers {
    census = {
      source  = "sutrolabs/census"
      version = "~> 2.0.0"
    }
  }
}

provider "census" {
  personal_access_token = var.census_personal_token
}
```

## Authentication

The Activations provider uses Personal Access Tokens (PAT) for authentication.

To obtain a Personal Access Token:

1. Log in to Fivetran
2. Navigate to Activations > Settings > Organization Settings > User Settings > Personal Access Tokens
3. Generate a new Personal Access Token
4. Store it securely (e.g., in environment variables or a secret manager)

## Schema

### Required

- `personal_access_token` (String, Sensitive) Personal Access Token for Activations API authentication. Can also be set via the `CENSUS_PERSONAL_ACCESS_TOKEN` environment variable.

### Optional

- `base_url` (String) Custom base URL for the Activations API. Primarily used for testing against staging environments. Can also be set via the `CENSUS_BASE_URL` environment variable.

## Resources

The Activations provider supports the following resources:

- `census_source` - Data warehouse connections (Snowflake, BigQuery, Postgres, Redshift, etc.)
- `census_activation` - Business tool integrations (Salesforce, HubSpot, etc.)
- `census_dataset` - SQL datasets for data transformation
- `census_sync` - Data syncs between sources and destinations

## Data Sources

All resources have corresponding data sources for read-only operations:

- `census_source`
- `census_activation`
- `census_dataset`
- `census_sync`

For detailed documentation on each resource and data source, see the navigation menu.
