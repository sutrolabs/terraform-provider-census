# Testing Guide

This document describes how to test the Census Terraform Provider at different levels.

## Testing Strategy

The Census Terraform Provider uses a **2-phase testing approach**:

### Phase 1: Unit Tests ✅ (Fast, No Dependencies)

Tests individual components without external dependencies. These tests validate business logic, data transformations, and helper functions.

```bash
# Run all unit tests
make test

# Or use go test directly
go test ./... -short -v

# Run specific package tests
go test ./census/tests/client -v
go test ./census/tests/provider/unit -v
```

**What's tested:**
- Client configuration validation
- URL building and parameter handling
- Error message formatting
- Provider schema validation
- Field mapping expand/flatten logic
- Alert and schedule configuration helpers
- Data type conversions

**Speed:** < 1 second
**Requirements:** None (no API access needed)

---

### Phase 2: Acceptance Tests 🔒 (Real API, Full Lifecycle)

Tests the complete Terraform lifecycle against the real Census staging API. These create actual resources and verify full CRUD operations.

```bash
# IMPORTANT: Requires .env.test file with credentials
# Copy .env.test.example and fill in your staging credentials
cp .env.test.example .env.test
# Edit .env.test with your credentials

# Run all acceptance tests
make test-integration

# Run specific test
TF_ACC=1 go test ./census/tests/provider/acceptance -v -run TestAccResourceSync_Basic
```

**What's tested:**
- Full Terraform resource lifecycle (create → read → update → delete)
- Real API authentication with Census staging environment
- Workspace creation and management
- Source connections (Redshift)
- Destination connections (Salesforce with JWT OAuth)
- Sync configuration (field mappings, schedules, alerts, run modes)
- Data source lookups
- Resource import
- State management
- Multi-resource dependencies

**Speed:** 5-10 minutes (creates real resources)
**Requirements:**
- Census staging account and personal access token
- Redshift test database credentials
- Salesforce sandbox with JWT OAuth configured
- See `.env.test.example` for complete setup guide

**Current Status:** 15 of 16 acceptance tests passing (93.75%)

## Test Coverage

### Unit Tests (17 tests)

Located in `census/tests/provider/unit/` and `census/tests/client/`:

**Provider Tests:**
- Provider schema validation
- Provider implementation interface check

**Client Tests:**
- Client creation & configuration (`TestNewClient`)
- API error handling (`TestAPIError_Error`)
- URL parameter building (`TestListOptions_ToParams`)

**Sync Resource Helper Tests:**
- Field mapping expand/flatten (6 tests)
- Alert configuration (2 tests)
- Schedule configuration (3 tests)
- String list/map helpers (3 tests)

### Acceptance Tests (16 tests, 15 passing)

Located in `census/tests/provider/acceptance/`:

**Workspace Tests (3 tests):**
- Basic workspace creation
- Workspace updates
- Workspace with API key retrieval

**Source Tests (2 tests + 1 data source):**
- Redshift source creation
- Redshift source updates
- Source data source lookup

**Destination Tests (2 tests + 1 data source):**
- Salesforce destination creation with JWT OAuth
- Salesforce destination updates
- Destination data source lookup

**Sync Tests (7 tests):**
- Basic sync creation with field mappings
- Sync updates (label, paused state)
- Field mappings (constant, liquid template) ⚠️ 1 failing - API issue
- RunMode - Daily schedule
- RunMode - Hourly schedule
- RunMode - Manual (never) schedule
- Alert configurations

## Quick Test Commands

### Run Tests
```bash
# Unit tests only (fast, no credentials needed)
make test

# Acceptance tests (requires .env.test with credentials)
make test-integration
# Or: make test-acc (alias)

# Run with coverage report
make test-coverage

# Run specific acceptance test
TF_ACC=1 go test ./census/tests/provider/acceptance -v -run TestAccResourceSync_Alerts
```

### Manual Testing with Examples
```bash
# Use the complete example to test all resource types
cd examples/complete-census-setup/
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with your credentials
terraform init
terraform plan
terraform apply

# Test basic workspace only
cd examples/basic-workspace/
terraform init
terraform apply
```

## Test Data & Fixtures

### Test Configurations
Example test configurations in `examples/`:
- `complete-census-setup/` - Full workflow demonstrating all 5 resource types (workspace, source, destination, sync, dataset)
- `basic-workspace/` - Simple workspace creation and management
- `staging-example/` - Configuration for Census staging environment testing

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

1. **Dataset Resource**: No acceptance tests for dataset resource (code exists but untested)
2. **Data Sources**: Missing tests for workspace, sync, and dataset data sources
3. **Import Functionality**: Import is implemented but lacks dedicated test coverage
4. **Advanced Field Mappings**: Constant and liquid template mappings (failing due to API issue in staging)
5. **Multi-region Testing**: EU region endpoint testing
6. **Error Edge Cases**: Some API-specific error scenarios and validation edge cases

## Adding New Tests

### For New Resources
1. Add unit tests in `census/provider/*_test.go`
2. Add client tests in `census/client/*_test.go`
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

**For Unit Tests:**
- None required

**For Acceptance Tests:**
Create a `.env.test` file based on `.env.test.example`:
```bash
# Required
CENSUS_BASE_URL=https://app.staging.getcensus.com/api/v1
CENSUS_PERSONAL_ACCESS_TOKEN=your-staging-token

# Redshift credentials (5 variables)
CENSUS_TEST_REDSHIFT_HOST=your-cluster.region.redshift.amazonaws.com
CENSUS_TEST_REDSHIFT_PORT=5439
CENSUS_TEST_REDSHIFT_DATABASE=dev
CENSUS_TEST_REDSHIFT_USERNAME=testuser
CENSUS_TEST_REDSHIFT_PASSWORD=your-password

# Salesforce JWT OAuth (5 variables)
CENSUS_TEST_SALESFORCE_USERNAME=test@example.com
CENSUS_TEST_SALESFORCE_INSTANCE_URL=https://your-sandbox.develop.my.salesforce.com
CENSUS_TEST_SALESFORCE_CLIENT_ID=3MVG9...
CENSUS_TEST_SALESFORCE_JWT_SIGNING_KEY='-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----\n'
CENSUS_TEST_SALESFORCE_DOMAIN=test.salesforce.com

# For debugging
TF_LOG=TRACE
```

See `.env.test.example` for detailed setup instructions.

---

This 2-phase testing strategy ensures reliability while providing flexibility for different testing scenarios.