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

## Local API Reference
- **Current API Spec**: `census_api_20250926_131957.yaml` - Downloaded on 2025-09-26 17:19:57 UTC
- **CRITICAL**: ALWAYS refer to the local timestamped Census API YAML file for accurate API structure, required fields, and data types
- **Update Process**: Pull fresh API spec from https://developers.getcensus.com/openapi/compiled/workspace_management.yaml weekly and save with timestamp
- **Never assume API structure** - Always check the local YAML first before making changes

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

## Important Development Rules
- **ALWAYS UPDATE TODO.md**: After completing any significant work, immediately update TODO.md to reflect current progress and mark completed items
- **ALWAYS CONSULT OPENAPI SPECS**: Before making any API-related changes, ALWAYS read and carefully understand the OpenAPI specifications. Never make assumptions about API structure, required fields, or data types without first checking the official API documentation.
- **Keep TODO.md Current**: TODO.md should be the authoritative source of what's done vs what's next
- **Document Major Changes**: Add major architectural decisions and breakthroughs to TODO.md status sections

## Debugging Common Issues

### "Failed to load plugin schemas" / "Reattachment process not found"
This error occurs when Terraform tries to connect to a provider process that has crashed, timed out, or been killed. Here's the systematic solution:

**Root Cause**: Provider process crashes or exits, but Terraform still has the old PID in TF_REATTACH_PROVIDERS

**Solution Steps**:
1. **Kill all running provider processes**: `pkill -f "terraform-provider-census"`
2. **Rebuild the provider**: `go build -o bin/terraform-provider-census` 
3. **Start provider in debug mode**: `./terraform-provider-census -debug`
4. **Copy the new TF_REATTACH_PROVIDERS output** (contains new PID and socket path)
5. **Use the fresh TF_REATTACH_PROVIDERS** in your terraform command

**Example**:
```bash
# Kill old processes
pkill -f "terraform-provider-census"

# Build fresh binary
go build -o bin/terraform-provider-census
cd examples/complete-census-setup
cp ../../bin/terraform-provider-census ./terraform-provider-census-debug

# Start in debug mode and capture output
./terraform-provider-census-debug -debug
# Copy the TF_REATTACH_PROVIDERS='...' line from output

# Use in terraform command
TF_REATTACH_PROVIDERS='{"registry.terraform.io/your-org/census":{"Protocol":"grpc","ProtocolVersion":5,"Pid":12345,"Test":true,"Addr":{"Network":"unix","String":"/tmp/plugin123"}}}' terraform plan
```

**Key Points**:
- Always use a fresh build after code changes
- Provider must be running in background when terraform connects
- PID in TF_REATTACH_PROVIDERS must match actual running process
- Use background processes (`./provider -debug &`) or separate terminals