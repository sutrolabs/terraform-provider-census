# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2025-01-XX

### Added

- Initial release of the Census Terraform provider
- Support for Census API authentication (personal and workspace tokens)
- Multi-region support (US and EU)
- Workspace resource (`census_workspace`) with full CRUD operations
- Workspace data source (`census_workspace`) for reading workspace information
- Basic testing framework with unit and acceptance tests
- Comprehensive documentation and examples
- Makefile for development tasks

### Features

#### Resources
- **census_workspace**: Create, read, update, and delete Census workspaces
  - Support for notification emails configuration
  - Optional API key retrieval during creation
  - Import support for existing workspaces

#### Data Sources
- **census_workspace**: Read Census workspace information by ID

#### Provider Configuration
- Support for both personal access tokens and workspace access tokens
- Region selection (US/EU) with automatic endpoint configuration
- Environment variable support for authentication
- Custom base URL configuration

#### Developer Experience
- Comprehensive test suite with unit and acceptance tests
- Local development support with Makefile
- Detailed documentation with examples
- Error handling with proper Census API error types

### Technical Details

- Built with terraform-plugin-sdk/v2 for modern Terraform compatibility
- Go 1.21+ support
- Structured logging and error handling
- Automatic pagination support for list operations
- HTTP client with proper timeout and retry handling

---

## Unreleased

### Planned Features

- **Sync resources**: Manage Census data syncs
- **Destination resources**: Configure data destinations
- **Source resources**: Manage data sources
- **Dataset resources**: Handle dataset configurations
- **Advanced operations**: Sync runs, webhooks, bulk operations
- **Enhanced testing**: Integration tests with mock Census API
- **Performance improvements**: Connection pooling and caching