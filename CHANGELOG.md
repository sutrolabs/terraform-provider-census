# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2025-09-29

### Added

- Initial release of the Census Terraform provider
- Complete Census data pipeline management from sources to syncs
- PAT-only authentication with dynamic workspace token retrieval
- Multi-region support (US and EU)
- Comprehensive resource and data source support

### Resources

#### Core Resources
- **census_workspace**: Create, read, update, and delete Census workspaces
  - Notification emails configuration
  - API key retrieval on creation
  - Import support for existing workspaces

- **census_source**: Manage data source connections
  - Support for all Census-supported databases (Snowflake, BigQuery, Postgres, Redshift, etc.)
  - Connection credential management with validation
  - Auto table refresh functionality
  - OpenAPI schema validation

- **census_destination**: Configure sync destinations
  - Support for all Census-supported destinations (Salesforce, HubSpot, etc.)
  - Dynamic connector type validation via Census API
  - Connection testing and credential management
  - Auto-refresh metadata after creation

- **census_dataset**: SQL dataset management for data transformation
  - Multi-line SQL query support with heredoc syntax
  - Column schema discovery (computed fields)
  - Source connection reference and validation
  - Resource identifier generation

- **census_sync**: Manage data syncs between sources and destinations
  - Field mapping configuration (direct, hash, constant operations)
  - Sync scheduling (hourly, daily, weekly, manual modes)
  - Sync mode support (upsert, append, mirror)
  - Support for all source types (table, dataset, model, topic, segment, cohort)
  - OpenAPI-compliant source attributes

#### Data Sources
- All resources have corresponding data sources for read-only operations
- **census_workspace**, **census_source**, **census_destination**, **census_dataset**, **census_sync**

### Provider Configuration

#### Authentication
- PAT-only authentication model
- Automatic workspace token retrieval for workspace-scoped operations
- Region selection (US/EU) with automatic endpoint configuration
- Environment variable support: `CENSUS_PERSONAL_ACCESS_TOKEN`
- Custom base URL configuration

### Technical Highlights

- Built with terraform-plugin-sdk/v2 for modern Terraform compatibility
- Go 1.21+ support
- Dynamic API schema validation against Census OpenAPI specifications
- Robust state management with workspace_id persistence
- Comprehensive error handling with helpful messages
- Pagination support for list operations
- TypeSet-based field mappings to prevent order-based drift
- Import support for all resources

---

## Unreleased

### Planned Features

- **Sync run operations**: Execute and monitor sync runs
- **Webhook management**: Event notifications and integrations
- **Advanced testing**: Comprehensive integration and acceptance tests
- **Performance improvements**: Request batching, caching strategies
- **Enhanced documentation**: Video tutorials, migration guides
- **Terraform Registry publication**: Official registry listing