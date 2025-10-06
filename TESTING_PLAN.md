# Census Terraform Provider - Testing Plan

**Created:** 2025-10-05
**Updated:** 2025-10-06
**Purpose:** Systematically add test coverage to match industry standards (Fivetran, HashiCorp providers)
**Target Coverage:** 70%+ code coverage across all resources

---

## Executive Summary

This plan provides a systematic approach to achieving comprehensive test coverage for the Census Terraform Provider, following HashiCorp's testing best practices and patterns used in production providers like Fivetran.

**Current Status:**

- ✅ **Unit tests**: 17 tests covering provider schema, client logic, and sync helper functions
- ✅ **Acceptance tests**: 19 tests (all resources covered, 1 sync test has simplified field mappings due to staging API limitations)
  - Workspace: 3 tests (basic, update, API key retrieval)
  - Source: 2 resource tests + 1 data source test (Redshift)
  - Destination: 2 resource tests + 1 data source test (Salesforce JWT OAuth)
  - Dataset: 3 tests (basic, update, dataset-as-source in sync)
  - Sync: 7 tests (basic CRUD, field mappings with constant values, schedules, alerts, run modes)
- ❌ **Missing data source tests for**: workspace, sync, dataset
- ❌ **No import tests**: Import functionality exists but lacks test coverage

**Test Coverage:** High acceptance test coverage across all resource types

**Goal:** Achieve 100% acceptance test pass rate and add missing test coverage for untested resources.

---

## Testing Strategy

We use a **2-phase testing approach** for simplicity and maintainability:

### Phase 1: Unit Tests (Fast, No External Dependencies)

**Purpose:** Test individual functions, validation logic, schema definitions

**What's Tested:**

- Schema validation (field types, required fields, defaults)
- Helper function logic (expand/flatten functions)
- Input validation
- Error message formatting
- Type conversions

**Runs:** Every commit, locally, in CI
**Speed:** < 1 second
**Dependencies:** None
**Command:** `make test` or `go test ./... -short -v`

**Environment Variables:** None required

---

### Phase 2: Integration Tests (Real API)

**Purpose:** Full Terraform lifecycle testing with real Census staging API

**What's Tested:**

- Resource create/read/update/delete with real API
- Import existing resources
- Plan/apply/refresh/destroy cycles
- Multi-resource dependencies (workspace → source → destination → sync)
- State management
- Field mappings, schedules, alerts, and all sync features

**Runs:** On-demand, before releases, in CI
**Speed:** Minutes (creates real resources)
**Dependencies:**

- Census staging environment (`app.staging.getcensus.com`)
- Real Redshift database credentials
- Real Salesforce sandbox credentials

**Command:** `make test-integration` (alias: `make test-acc`)

**Environment Variables Required:**

```bash
# Copy .env.test.example to .env.test and fill in:
CENSUS_BASE_URL=https://app.staging.getcensus.com/api/v1
CENSUS_PERSONAL_ACCESS_TOKEN=your-staging-personal-access-token

# Redshift credentials (5 variables)
CENSUS_TEST_REDSHIFT_HOST=your-cluster.region.redshift.amazonaws.com
CENSUS_TEST_REDSHIFT_PORT=5439
CENSUS_TEST_REDSHIFT_DATABASE=dev
CENSUS_TEST_REDSHIFT_USERNAME=testuser
CENSUS_TEST_REDSHIFT_PASSWORD=your-password

# Salesforce JWT OAuth credentials (5 variables)
# IMPORTANT: Use JWT-based auth, not password-based
CENSUS_TEST_SALESFORCE_USERNAME=test@example.com
CENSUS_TEST_SALESFORCE_INSTANCE_URL=https://your-sandbox.develop.my.salesforce.com
CENSUS_TEST_SALESFORCE_CLIENT_ID=3MVG9...your-consumer-key...
CENSUS_TEST_SALESFORCE_JWT_SIGNING_KEY='-----BEGIN RSA PRIVATE KEY-----\nMIIEpQIBAAK...\n-----END RSA PRIVATE KEY-----\n'
CENSUS_TEST_SALESFORCE_DOMAIN=test.salesforce.com
```

**Test Flow:**

1. Each test creates a complete stack from scratch:
   - Create workspace
   - Create Redshift source
   - Create Salesforce destination
   - Create sync with specific configuration to test
2. Verify all attributes in Terraform state
3. Clean up resources (CheckDestroy)

**Why This Approach:**

- Tests the actual user workflow (creating everything from scratch)
- Validates full resource creation lifecycle
- No dependency on pre-created resources that could drift or break
- Easy to run locally and in CI
- Self-contained tests that don't interfere with each other

---

## Current Test Coverage Audit

### ✅ Resources WITH Tests

| Resource             | File                      | Test File                           | Status     | Coverage |
| -------------------- | ------------------------- | ----------------------------------- | ---------- | -------- |
| `census_workspace`   | `resource_workspace.go`   | `resource_workspace_test.go`        | ✅ Partial | ~40%     |
| `census_source`      | `resource_source.go`      | `resource_source_test.go`           | ✅ Partial | ~30%     |
| `census_destination` | `resource_destination.go` | `resource_destination_test.go`      | ✅ Partial | ~30%     |
| `census_sync`        | `resource_sync.go`        | `resource_sync_integration_test.go` | ✅ Good    | ~60%     |

**What's tested:**

- Basic resource creation with real Redshift and Salesforce
- Resource updates
- Workspace management
- Source/destination connections
- Sync: field mappings (direct, constant, liquid template)
- Sync: run modes (manual, daily, hourly)
- Sync: alerts (failure, invalid record percent)

**What's NOT tested:**

- Import functionality
- Complex validation scenarios
- Error handling edge cases
- State migration

### ❌ Data Sources WITHOUT Tests

| Data Source          | File                         | Test File                          | Status      | Priority    |
| -------------------- | ---------------------------- | ---------------------------------- | ----------- | ----------- |
| `census_sync`        | `data_source_sync.go`        | ❌ `data_source_sync_test.go`      | **MISSING** | 🔴 CRITICAL |
| `census_dataset`     | `data_source_dataset.go`     | ❌ `data_source_dataset_test.go`   | **MISSING** | 🟡 HIGH     |
| `census_workspace`   | `data_source_workspace.go`   | ❌ `data_source_workspace_test.go` | **MISSING** | 🟡 MEDIUM   |
| `census_source`      | `data_source_source.go`      | `data_source_source_test.go`       | ✅ Complete | 🟢 DONE     |
| `census_destination` | `data_source_destination.go` | `data_source_destination_test.go`  | ✅ Complete | 🟢 DONE     |

### ❌ Client Methods WITHOUT Tests

| Client             | File                    | Test File  | Status         | Priority    |
| ------------------ | ----------------------- | ---------- | -------------- | ----------- |
| Sync Client        | `client/sync.go`        | ❌ Missing | **MISSING**    | 🔴 CRITICAL |
| Dataset Client     | `client/dataset.go`     | ❌ Missing | **MISSING**    | 🟡 HIGH     |
| Source Client      | `client/source.go`      | ⚠️ Minimal | **INCOMPLETE** | 🟡 MEDIUM   |
| Destination Client | `client/destination.go` | ⚠️ Minimal | **INCOMPLETE** | 🟡 MEDIUM   |

---

## Implementation Plan

### Phase 1: Complete Integration Test Coverage (Week 1-2) 🔴

**Goal:** Add integration tests for all missing resources and data sources

#### Task 1.1: `resource_dataset_test.go` ✅ Create File

**Priority:** HIGH
**Estimated Time:** 3-4 hours

**Test Scenarios:**

1. **TestAccResourceDataset_basic**

   - Create SQL dataset
   - Verify query and source_id

2. **TestAccResourceDataset_update**

   - Update query
   - Update description

3. **TestAccResourceDataset_import**

   - Test import functionality

4. **TestAccResourceDataset_validation**
   - Test query validation
   - Test source_id validation

#### Task 1.2: `data_source_sync_test.go` ✅ Create File

**Priority:** CRITICAL
**Estimated Time:** 2-3 hours

**Test Scenarios:**

1. **TestAccDataSourceSync_basic**

   - Create sync first
   - Read sync by ID
   - Verify all attributes

2. **TestAccDataSourceSync_complex**
   - Verify field_mapping attributes
   - Verify run_mode structure
   - Verify alert configurations

#### Task 1.3: `data_source_dataset_test.go` ✅ Create File

**Priority:** HIGH
**Estimated Time:** 2 hours

**Test Scenarios:**

1. **TestAccDataSourceDataset_basic**
   - Create dataset first
   - Read dataset by ID
   - Verify query and columns

#### Task 1.4: `data_source_workspace_test.go` ✅ Create File

**Priority:** MEDIUM
**Estimated Time:** 1-2 hours

**Test Scenarios:**

1. **TestAccDataSourceWorkspace_basic**
   - Create workspace first
   - Read workspace by ID
   - Verify attributes

---

### Phase 2: Expand Existing Test Coverage (Week 3) 🟢

**Goal:** Add missing test scenarios to existing test files

#### Task 2.1: Enhance `resource_workspace_test.go`

**Estimated Time:** 2-3 hours

**Add Missing Scenarios:**

- [ ] TestAccResourceWorkspace_import
- [ ] TestAccResourceWorkspace_validation_errors
- [ ] TestAccResourceWorkspace_empty_notification_emails
- [ ] TestAccResourceWorkspace_concurrent_updates

#### Task 2.2: Enhance `resource_source_test.go`

**Estimated Time:** 3-4 hours

**Add Missing Scenarios:**

- [ ] TestAccResourceSource_all_types (snowflake, bigquery, postgres, redshift, etc.)
- [ ] TestAccResourceSource_connection_validation
- [ ] TestAccResourceSource_auto_refresh
- [ ] TestAccResourceSource_import
- [ ] TestAccResourceSource_update_credentials

#### Task 2.3: Enhance `resource_destination_test.go`

**Estimated Time:** 3-4 hours

**Add Missing Scenarios:**

- [ ] TestAccResourceDestination_all_types (salesforce, hubspot, etc.)
- [ ] TestAccResourceDestination_connection_validation
- [ ] TestAccResourceDestination_auto_refresh
- [ ] TestAccResourceDestination_import
- [ ] TestAccResourceDestination_update_credentials

#### Task 2.4: Enhance `resource_sync_integration_test.go`

**Estimated Time:** 3-4 hours

**Add Missing Scenarios:**

- [ ] TestAccResourceSync_dataset_source (using dataset instead of table)
- [ ] TestAccResourceSync_operations (append, mirror, mirror_strategy)
- [ ] TestAccResourceSync_run_mode_triggers (dbt_cloud, fivetran, sync_sequence)
- [ ] TestAccResourceSync_advanced_features (high_water_mark, historical_sync)
- [ ] TestAccResourceSync_import
- [ ] TestAccResourceSync_validation

---

### Phase 3: Unit Tests for Helper Functions (Week 4) 🔵

**Goal:** Add unit tests for all expand/flatten helper functions

#### Task 3.1: Sync Helper Functions

**File:** Create `resource_sync_unit_test.go`
**Estimated Time:** 4-5 hours

**Functions to Test:**

- [ ] `expandSourceAttributes()` - all source types
- [ ] `flattenSourceAttributes()` - all source types
- [ ] `expandFieldMappings()` - all mapping types
- [ ] `flattenFieldMappings()` - all mapping types
- [ ] `expandRunMode()` - all trigger types
- [ ] `flattenRunMode()` - all trigger types
- [ ] `expandAlerts()` - all alert types
- [ ] `flattenAlerts()` - all alert types
- [ ] `expandAdvancedConfiguration()`
- [ ] `cleanEmptyStrings()` - edge cases

**Test Pattern:**

```go
func TestExpandFieldMappings(t *testing.T) {
    tests := []struct {
        name     string
        input    []interface{}
        expected []client.MappingAttribute
    }{
        {
            name: "direct mapping",
            input: []interface{}{
                map[string]interface{}{
                    "from": "email",
                    "to": "Email",
                    "is_primary_identifier": true,
                },
            },
            expected: []client.MappingAttribute{
                {
                    From: "email",
                    To: "Email",
                    IsPrimaryIdentifier: true,
                },
            },
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := expandFieldMappings(tt.input)
            if !reflect.DeepEqual(result, tt.expected) {
                t.Errorf("got %+v, want %+v", result, tt.expected)
            }
        })
    }
}
```

#### Task 3.2: Source/Destination Helper Functions

**Files:** Unit test files for each resource
**Estimated Time:** 2-3 hours

**Functions to Test:**

- [ ] `expandConnectionConfig()`
- [ ] `flattenConnectionConfig()`
- [ ] Connection validation helpers

#### Task 3.3: Dataset Helper Functions

**Estimated Time:** 1-2 hours

**Functions to Test:**

- [ ] SQL query validation
- [ ] Column flattening

---

### Phase 4: Integration Tests for Client Methods (Week 5) 🟣

**Goal:** Add integration tests using httptest mock server

#### Task 4.1: Sync Client Integration Tests

**File:** Create `client/sync_integration_test.go`
**Estimated Time:** 4-5 hours

**Methods to Test:**

- [ ] `CreateSync()` - test request formation, response parsing
- [ ] `GetSync()` - test 200 response, 404 handling
- [ ] `UpdateSync()` - test PATCH request, partial updates
- [ ] `DeleteSync()` - test DELETE request
- [ ] `ListSyncs()` - test pagination

**Test Pattern:**

```go
func TestSyncClient_CreateSync(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Verify request
        if r.Method != "POST" {
            t.Errorf("Expected POST, got %s", r.Method)
        }
        if r.URL.Path != "/api/v1/syncs" {
            t.Errorf("Expected /api/v1/syncs, got %s", r.URL.Path)
        }

        // Send mock response
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "id": "123",
            "label": "Test Sync",
        })
    }))
    defer server.Close()

    client := &Client{BaseURL: server.URL, HTTPClient: &http.Client{}}

    sync, err := client.CreateSync(context.Background(), &CreateSyncRequest{
        Label: "Test Sync",
    }, "token")

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if sync.ID != "123" {
        t.Errorf("expected ID 123, got %s", sync.ID)
    }
}
```

#### Task 4.2: Dataset Client Integration Tests

**File:** Create `client/dataset_integration_test.go`
**Estimated Time:** 2-3 hours

**Methods to Test:**

- [ ] `CreateDataset()`
- [ ] `GetDataset()`
- [ ] `UpdateDataset()`
- [ ] `DeleteDataset()`
- [ ] `ListDatasets()`

#### Task 4.3: Enhanced Client Base Tests

**File:** Enhance `client/client_test.go`
**Estimated Time:** 2-3 hours

**Additional Tests:**

- [ ] Error response parsing
- [ ] Retry logic (if implemented)
- [ ] Rate limiting handling
- [ ] Multi-region URL construction

---

## Test Patterns & Templates

### Standard Integration Test Structure

```go
package acceptance

import (
    "fmt"
    "os"
    "testing"

    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
    provider_test "github.com/sutrolabs/terraform-provider-census/census/tests/provider"
)

// Basic test - create and verify
func TestAccResource{Name}_basic(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:  func() { provider_test.TestAccPreCheckIntegration(t) },
        Providers: provider_test.TestAccProviders,
        Steps: []resource.TestStep{
            {
                Config: testAcc{Name}Config_basic(),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttr("census_{resource}.test", "field", "value"),
                    resource.TestCheckResourceAttrSet("census_{resource}.test", "id"),
                ),
            },
        },
    })
}

// Update test - create, update, verify
func TestAccResource{Name}_update(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:  func() { provider_test.TestAccPreCheckIntegration(t) },
        Providers: provider_test.TestAccProviders,
        Steps: []resource.TestStep{
            {
                Config: testAcc{Name}Config_basic(),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttr("census_{resource}.test", "field", "original"),
                ),
            },
            {
                Config: testAcc{Name}Config_updated(),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttr("census_{resource}.test", "field", "updated"),
                ),
            },
        },
    })
}

// Configuration helper - creates full stack
func testAccIntegrationBaseConfig() string {
    return fmt.Sprintf(`
provider "census" {
  base_url = "%s"
}

resource "census_workspace" "test" {
  name = "Test Workspace"
  notification_emails = ["test@example.com"]
}

resource "census_source" "test" {
  workspace_id = census_workspace.test.id
  name = "Test Redshift Source"
  type = "redshift"
  connection_config = {
    host     = "%s"
    port     = "%s"
    database = "%s"
    username = "%s"
    password = "%s"
  }
}

resource "census_destination" "test" {
  workspace_id = census_workspace.test.id
  name = "Test Salesforce Destination"
  type = "salesforce"
  connection_config = {
    username       = "%s"
    password       = "%s"
    security_token = "%s"
    sandbox        = "%s"
  }
}
`,
        os.Getenv("CENSUS_BASE_URL"),
        os.Getenv("CENSUS_TEST_REDSHIFT_HOST"),
        getEnvOrDefault("CENSUS_TEST_REDSHIFT_PORT", "5439"),
        os.Getenv("CENSUS_TEST_REDSHIFT_DATABASE"),
        os.Getenv("CENSUS_TEST_REDSHIFT_USERNAME"),
        os.Getenv("CENSUS_TEST_REDSHIFT_PASSWORD"),
        os.Getenv("CENSUS_TEST_SALESFORCE_USERNAME"),
        os.Getenv("CENSUS_TEST_SALESFORCE_PASSWORD"),
        os.Getenv("CENSUS_TEST_SALESFORCE_SECURITY_TOKEN"),
        getEnvOrDefault("CENSUS_TEST_SALESFORCE_SANDBOX", "true"),
    )
}

func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
```

### Standard Unit Test Structure

```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   inputType
        want    outputType
        wantErr bool
    }{
        {
            name:  "valid input",
            input: validInput,
            want:  expectedOutput,
        },
        {
            name:    "invalid input",
            input:   invalidInput,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FunctionUnderTest(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("got = %v, want %v", got, tt.want)
            }
        })
    }
}
```

---

## Running Tests

### Quick Commands

```bash
# Unit tests only (fast, no credentials needed)
make test
# or
go test ./... -short -v

# Integration tests (requires Census staging credentials)
make test-integration
# or (alias)
make test-acc

# Specific resource tests
go test ./census/tests/provider/acceptance -run TestAccResourceSync -v

# Coverage report
make test-coverage
```

### Makefile Targets

```makefile
test: ## Run unit tests (no credentials needed)
	go test ./... -short -v

test-integration: ## Run integration tests (creates all resources in staging)
	@echo "Running integration tests against Census staging API..."
	@echo "This will create workspaces, sources, destinations, and syncs in staging"
	@echo "Requires: .env.test with staging credentials (see .env.test.example)"
	@if [ ! -f .env.test ]; then \
		echo "Error: .env.test not found. Copy .env.test.example and fill in your credentials."; \
		exit 1; \
	fi
	@set -a && . ./.env.test && set +a && TF_ACC=1 go test -v ./census/tests/provider/acceptance -timeout 60m

test-acc: test-integration ## Alias for test-integration (Terraform convention)

test-coverage: ## Generate test coverage report
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
```

---

## Progress Tracking

### ✅ Completed Integration Tests

**Resources:**

- ✅ resource_workspace_test.go - Basic CRUD
- ✅ resource_source_test.go - Redshift connection
- ✅ resource_destination_test.go - Salesforce connection
- ✅ resource_sync_integration_test.go - 7 comprehensive test scenarios
  - Basic sync creation
  - Sync updates
  - Field mappings (direct, constant, liquid template)
  - Run modes (manual, daily, hourly)
  - Alerts (failure, invalid record percent)

**Data Sources:**

- ✅ data_source_source_test.go - Complete
- ✅ data_source_destination_test.go - Complete

### ❌ Missing Integration Tests

**Resources:**

- [ ] resource_dataset_test.go - MISSING

**Data Sources:**

- [ ] data_source_sync_test.go - MISSING
- [ ] data_source_dataset_test.go - MISSING
- [ ] data_source_workspace_test.go - MISSING

### ❌ Missing Unit Tests

- [ ] All expand/flatten helper functions (Phase 3)

### ❌ Missing Client Integration Tests

- [ ] All client methods with httptest mocks (Phase 4)

---

## Coverage Goals

### Target Coverage by Component

| Component               | Current  | Target   | Priority    |
| ----------------------- | -------- | -------- | ----------- |
| resource_sync.go        | ~60%     | 80%      | 🟡 Medium   |
| resource_dataset.go     | 0%       | 70%      | 🟡 High     |
| resource_workspace.go   | ~40%     | 80%      | 🟢 Medium   |
| resource_source.go      | ~30%     | 75%      | 🟢 Medium   |
| resource_destination.go | ~30%     | 75%      | 🟢 Medium   |
| data*source*\*.go       | ~30%     | 70%      | 🟡 High     |
| client/sync.go          | 0%       | 70%      | 🔴 Critical |
| client/dataset.go       | 0%       | 70%      | 🟡 High     |
| **Overall Provider**    | **~35%** | **70%+** | 🔴 Critical |

### Success Metrics

- ✅ All resources have integration tests
- ⏳ All data sources have integration tests (2/5 done)
- ❌ All client methods have integration tests (0% done)
- ❌ All helper functions have unit tests (0% done)
- ⏳ Import functionality tested for all resources (0/4 done)
- ⏳ Validation tested for all resources (partial)
- ✅ Update scenarios tested for key resources
- ❌ Coverage reports generated and tracked
- ❌ Tests run in CI/CD pipeline

---

## CI/CD Integration

### GitHub Actions Workflow (Future)

```yaml
name: Tests

on: [push, pull_request]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: "1.21"
      - name: Run unit tests
        run: make test

  integration-tests:
    if: github.event_name == 'workflow_dispatch'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: "1.21"
      - name: Run integration tests
        env:
          CENSUS_BASE_URL: ${{ secrets.CENSUS_STAGING_URL }}
          CENSUS_PERSONAL_ACCESS_TOKEN: ${{ secrets.CENSUS_TEST_TOKEN }}
          CENSUS_TEST_REDSHIFT_HOST: ${{ secrets.TEST_REDSHIFT_HOST }}
          CENSUS_TEST_REDSHIFT_DATABASE: ${{ secrets.TEST_REDSHIFT_DATABASE }}
          CENSUS_TEST_REDSHIFT_USERNAME: ${{ secrets.TEST_REDSHIFT_USERNAME }}
          CENSUS_TEST_REDSHIFT_PASSWORD: ${{ secrets.TEST_REDSHIFT_PASSWORD }}
          CENSUS_TEST_SALESFORCE_USERNAME: ${{ secrets.TEST_SALESFORCE_USERNAME }}
          CENSUS_TEST_SALESFORCE_PASSWORD: ${{ secrets.TEST_SALESFORCE_PASSWORD }}
          CENSUS_TEST_SALESFORCE_SECURITY_TOKEN: ${{ secrets.TEST_SALESFORCE_TOKEN }}
        run: make test-integration

  coverage:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: "1.21"
      - name: Generate coverage
        run: make test-coverage
      - name: Upload coverage
        uses: actions/upload-artifact@v3
        with:
          name: coverage-report
          path: coverage.html
```

---

## Notes & Best Practices

### Testing Anti-Patterns to Avoid

❌ **Don't:** Write tests that depend on pre-created external resources
❌ **Don't:** Write tests that require manual setup
❌ **Don't:** Hard-code IDs or credentials in tests
❌ **Don't:** Skip cleanup (always use CheckDestroy when possible)
❌ **Don't:** Test multiple things in one test case

### Testing Best Practices

✅ **Do:** Create all resources from scratch in each test
✅ **Do:** Use table-driven tests for unit tests
✅ **Do:** Test both success and error cases
✅ **Do:** Use meaningful test names (TestAcc{Resource}\_{scenario})
✅ **Do:** Use test fixtures and helpers (testAccIntegrationBaseConfig)
✅ **Do:** Run tests in parallel when possible
✅ **Do:** Document complex test scenarios
✅ **Do:** Use environment variables for credentials

### Resource Naming Convention

- Integration tests: `TestAcc{Resource}_{scenario}` (in `acceptance/` directory)
- Unit tests: `Test{Function}_{scenario}` (in same package as function)
- Integration tests: `Test{Client}_{method}_{scenario}` (in `client/` package)

### Test Data Management

- Use unique names with timestamps when needed: `test-sync-${timestamp}`
- Tests create dedicated workspace, source, and destination for each run
- Clean up resources after tests (automatically handled by Terraform)
- Don't rely on specific resource IDs

---

## Timeline & Milestones

### Week 1-2: Complete Integration Test Coverage

- ✅ Complete resource_sync_integration_test.go (7 test scenarios)
- [ ] Complete resource_dataset_test.go (4 test scenarios)
- [ ] Complete all missing data source tests (3 files)
- **Deliverable:** All resources and data sources have integration tests

### Week 3: Expand Coverage

- [ ] Add import tests for all resources
- [ ] Add validation tests
- [ ] Add update scenarios
- [ ] Test additional sync features (dataset source, operations, triggers)
- **Deliverable:** 70%+ coverage on existing resources

### Week 4: Unit Tests

- [ ] Test all expand/flatten helpers
- [ ] Test validation logic
- [ ] Test edge cases
- **Deliverable:** 80%+ coverage on helper functions

### Week 5: Client Integration Tests

- [ ] Mock HTTP tests for all client methods
- [ ] Error handling tests
- **Deliverable:** All client methods tested with mocks

### Week 6: Polish & Documentation

- [ ] Review coverage reports
- [ ] Add missing tests
- [ ] Update documentation
- **Deliverable:** Production-ready test suite

---

## References

- [HashiCorp Provider Testing Guide](https://developer.hashicorp.com/terraform/plugin/sdkv2/testing)
- [Fivetran Provider Tests](https://github.com/fivetran/terraform-provider-fivetran)
- [Terraform Plugin SDK Testing](https://github.com/hashicorp/terraform-plugin-testing)
- Existing Census Provider Tests (workspace, source, destination, sync)

---

**Last Updated:** 2025-10-06
**Status:** Phase 1 Integration Tests In Progress
**Current:** Completed sync integration tests (7 scenarios), 2/5 data sources complete
**Next Step:** Complete resource_dataset_test.go and remaining data source tests
