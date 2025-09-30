# Testing Guide

This document describes how to test the Census Terraform Provider at different levels.

## Test Types Available

### 1. Unit Tests ✅ (Currently Working)

Tests individual components without external dependencies.

```bash
# Run all unit tests
go test ./... -short -v

# Run specific package tests
go test ./internal/client -v
go test ./internal/provider -v
```

**What's tested:**
- Client configuration validation
- URL building and parameter handling
- Error message formatting
- Provider schema validation
- HTTP request creation and headers

### 2. Integration Tests 🧪 (Requires Census API)

Tests the full client flow against the real Census API. These tests are skipped by default in CI.

```bash
# Integration tests are skipped with -short flag
go test ./... -short -v

# To run integration tests (requires Census API access)
export CENSUS_PERSONAL_ACCESS_TOKEN="your-token"
go test ./internal/client -v -run TestWorkspaceIntegration
```

**What's tested:**
- Complete workspace CRUD operations
- API request/response handling
- Error scenarios (404, authentication)
- JSON serialization/deserialization

### 3. Acceptance Tests 🔒 (Requires Real Census API)

Tests against the actual Census API (requires valid tokens).

```bash
# Set up environment
export CENSUS_PERSONAL_ACCESS_TOKEN="your-real-token"
export TF_ACC=1

# Run acceptance tests
go test ./internal/provider -v -run TestResourceWorkspace
```

**What's tested:**
- Real API authentication
- Actual resource creation/modification
- Terraform state management
- End-to-end workflows

## Testing Scenarios

### ✅ Currently Testable (No External Dependencies)

1. **Provider Configuration**
   ```bash
   go test ./internal/provider -v -run TestProvider
   ```

2. **Client Creation & Configuration**
   ```bash
   go test ./internal/client -v -run TestNewClient
   ```

3. **URL Building & Parameter Handling**
   ```bash
   go test ./internal/client -v -run TestClient_buildURL
   go test ./internal/client -v -run TestListOptions_ToParams
   ```

4. **Error Handling**
   ```bash
   go test ./internal/client -v -run TestAPIError_Error
   ```

5. **HTTP Request Formation**
   ```bash
   go test ./internal/client -v -run TestClient_makeRequest
   ```

### 🧪 Testable with Mock Server

1. **Full Workspace CRUD Cycle**
   - Create workspace with notification emails
   - Retrieve workspace by ID
   - Update workspace name and emails
   - List all workspaces
   - Delete workspace
   - Verify 404 after deletion

2. **Authentication Flow**
   - Test with valid Bearer token
   - Test with missing/invalid token

3. **API Response Handling**
   - Success responses
   - Error responses (400, 404, 500)
   - Pagination handling

### 🔒 Testable with Real API (Requires Census Account)

1. **Real Authentication**
   - Personal access token validation
   - Workspace access token validation
   - Multi-region support (US/EU)

2. **Actual Resource Management**
   - Create real workspaces
   - Modify real workspace settings
   - Delete real workspaces

3. **Terraform Integration**
   - Full Terraform lifecycle
   - State management
   - Import existing resources
   - Plan/Apply/Destroy cycles

## Quick Test Commands

### Run All Available Tests
```bash
# Unit tests only (fast)
make test

# With integration tests (requires mock server)
make test-integration

# With acceptance tests (requires real API)
make test-acc
```

### Manual Testing with Examples
```bash
# Use the complete example to test all resources
cd examples/complete-census-setup/
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with your credentials
terraform init
terraform plan
```

### Specific Test Scenarios
```bash
# Test provider builds
go build .

# Test provider schema validation
go test ./internal/provider -run TestProvider -v

# Test client functionality
go test ./internal/client -v

# Test integration (requires Census API token)
export CENSUS_PERSONAL_ACCESS_TOKEN="your-token"
go test ./internal/client -run TestWorkspaceIntegration -v
```

## Test Data & Fixtures

### Test Configurations
Example test configurations in `examples/`:
- `complete-census-setup/` - Full workflow with all 5 resources
- `basic-workspace/` - Simple workspace creation and management

## Coverage Analysis

```bash
# Generate test coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# View coverage by package
go test ./... -cover
```

## CI/CD Testing

The provider includes GitHub Actions workflows for:
- **Unit tests** (no external dependencies) - runs on every push
- **Code quality** (go vet, go fmt) - runs on every push
- **Integration tests** - skipped in CI using `-short` flag
- **Acceptance tests** - requires Census API credentials (not run in CI)

## What's NOT Tested Yet

1. **Advanced Operations**: Sync runs, webhooks (not yet implemented)
2. **Import Functionality**: Requires real Census resources to test imports
3. **Error Edge Cases**: Some API-specific error scenarios
4. **Multi-region Testing**: EU region testing with real API

## Adding New Tests

### For New Resources
1. Add unit tests in `internal/provider/*_test.go`
2. Add client tests in `internal/client/*_test.go`
3. Add acceptance tests with `TF_ACC` flag
4. Update example configurations in `examples/`

### For New Features
1. Start with unit tests for business logic
2. Add integration tests for API interactions
3. Add acceptance tests for end-to-end validation

## Test Environment Setup

### Prerequisites
- Go 1.21+
- Valid Census account (for acceptance tests)
- Network access for integration tests

### Environment Variables
```bash
# For acceptance tests
export CENSUS_PERSONAL_ACCESS_TOKEN="your-token"
export CENSUS_WORKSPACE_ACCESS_TOKEN="your-workspace-token" 
export TF_ACC=1

# For debugging
export TF_LOG=TRACE
```

This testing strategy ensures reliability at multiple levels while providing flexibility for different testing scenarios.