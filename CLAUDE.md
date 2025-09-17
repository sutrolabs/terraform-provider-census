# Census Terraform Provider - Claude Session Context

## Project Overview
This is a Terraform provider for Census (https://getcensus.com), which enables infrastructure-as-code management of Census resources. Census has been acquired by Fivetran, and this provider is designed to feel similar to the Fivetran Terraform provider.

## Key API Information
- **Census API Documentation**: https://developers.getcensus.com/api-reference/introduction/overview
- **OpenAPI Specifications**: 
  - **Organization Management**: https://developers.getcensus.com/openapi/compiled/organization_management.yaml
  - **Workspace Management**: https://developers.getcensus.com/openapi/compiled/workspace_management.yaml
- **Census Connectors API**: https://app.getcensus.com/api/v1/connectors (requires /api/v1/ prefix)
- **Two API Levels**: 
  - Organization APIs: Use personal access tokens
  - Workspace APIs: Use workspace access tokens
- **Multi-region Support**: US (app.getcensus.com) and EU (app-eu.getcensus.com)
- **Authentication**: Bearer token based with different token types
- **API Endpoints**: Most endpoints require `/api/v1/` prefix (e.g., `/api/v1/connectors`, not `/connectors`)

## Project Structure
```
terraform-provider-census/
├── main.go                          # Main provider entry point
├── internal/
│   ├── provider/                    # Provider implementation
│   │   ├── provider.go             # Provider configuration and schema
│   │   ├── resource_workspace.go   # Workspace resource implementation
│   │   ├── data_source_workspace.go # Workspace data source
│   │   └── utils.go                # Helper utilities
│   └── client/                     # API client package
│       ├── client.go               # Base API client
│       └── workspace.go            # Workspace API operations
├── docs/                           # Documentation
├── examples/                       # Usage examples
└── scripts/                        # Build and utility scripts
```

## Current Implementation Status

### ✅ Completed
- **Project Structure**: Standard Go module with proper package organization
- **Provider Configuration**: 
  - Supports both personal and workspace access tokens
  - Region-aware (US/EU) endpoint configuration
  - Environment variable support
- **API Client**: 
  - Flexible HTTP client with authentication
  - Error handling with proper Census API error types
  - Pagination support for list operations
- **Workspace Resource**: Full CRUD operations
  - Create, Read, Update, Delete
  - Import support
  - Notification emails management
  - API key retrieval on creation

### 🚧 Next Priority Items
1. **Testing Framework**: Unit tests and acceptance tests
2. **Additional Resources**: Syncs, Destinations, Sources, Datasets
3. **Data Sources**: Read-only versions of all resources
4. **Advanced Features**: Sync runs, webhooks, bulk operations

## Key Implementation Details

### Authentication Pattern
The provider supports both token types through the configuration:
```hcl
provider "census" {
  personal_access_token  = "your-personal-token"  # For org-level operations
  workspace_access_token = "your-workspace-token" # For workspace operations
  region                 = "us"                   # or "eu"
}
```

### API Client Architecture
- Base client in `internal/client/client.go` handles HTTP operations
- Resource-specific clients (e.g., `workspace.go`) implement business logic
- Token type selection per operation (Personal vs Workspace)
- Automatic 404 handling for resource deletion

### Common Patterns
- Variable shadowing issue: Don't name variables `client` when importing the `client` package
- Use `apiClient` variable name for client instances
- Implement `IsNotFoundError()` utility for 404 handling
- Follow Terraform provider conventions for resource schemas

## Reference Examples
- **Fivetran Provider**: https://github.com/fivetran/terraform-provider-fivetran
- **DataDog Provider**: https://github.com/DataDog/terraform-provider-datadog
- **Provider Development Guide**: https://www.integralist.co.uk/posts/terraform-build-a-provider/

## Build and Test Commands
```bash
# Build the provider
go build .

# Run tests (once implemented)
go test ./...

# Install locally for testing
go install .
```

## Environment Variables
- `CENSUS_PERSONAL_ACCESS_TOKEN`: Personal access token
- `CENSUS_WORKSPACE_ACCESS_TOKEN`: Workspace access token
- `TF_LOG=TRACE`: Enable detailed Terraform logging for debugging

## Development Notes
- Provider builds successfully without errors
- Workspace resource implements full CRUD lifecycle
- Ready for extension with additional Census resources
- Uses terraform-plugin-sdk/v2 for modern Terraform compatibility