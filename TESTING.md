# Testing Guide

This document describes how to test the Census Terraform Provider.

## Test Types

### Unit Tests

Tests individual components without external dependencies. Fast and require no credentials.

```bash
make test
```

**What's tested:**
- Client configuration and URL building
- Provider schema validation
- Field mapping logic
- Alert and schedule helpers
- Data transformations

### Integration Tests

Tests against the real Census staging API. Creates actual resources to verify full CRUD operations.

```bash
# Requires .env.test file with credentials
make test-integration
```

**What's tested:**
- Complete resource lifecycle (create, read, update, delete)
- Workspaces, sources, destinations, datasets, and syncs
- Field mappings, schedules, and alerts
- Resource import
- Multi-resource dependencies

## Running Tests

### Quick Start

```bash
# Unit tests (no setup required)
make test

# Integration tests (requires setup)
cp .env.test.example .env.test
# Edit .env.test with your credentials
make test-integration

# Generate coverage report
make test-coverage
```

### Run Specific Tests

```bash
# Run a specific test
go test ./census/tests/provider/acceptance -v -run TestAccResourceSync_Basic

# Run with verbose output
go test ./... -v -short
```

## Integration Test Setup

Integration tests require:

1. **Census Staging Account**
   - Sign up at https://app.staging.getcensus.com
   - Generate a personal access token from Settings → Developer

2. **Redshift Test Database**
   - Use a dev/test cluster (not production)
   - Create read-only test user
   - Allow connections from your IP

3. **Salesforce Sandbox**
   - Use a sandbox (not production)
   - Configure Connected App with JWT OAuth
   - Generate RSA key pair for JWT signing

### Salesforce JWT OAuth Setup

```bash
# Generate RSA key pair
openssl genrsa -out server.key 2048
openssl req -new -x509 -key server.key -out server.crt -days 365
```

In Salesforce (Setup → App Manager → New Connected App):
- Enable OAuth Settings
- Select scopes: `api`, `refresh_token`, `offline_access`
- Enable "Use digital signatures"
- Upload `server.crt` certificate
- Get Consumer Key (Client ID)

### Environment Variables

Create `.env.test` from `.env.test.example` and fill in:

```bash
# Census
CENSUS_BASE_URL=https://app.staging.getcensus.com/api/v1
CENSUS_PERSONAL_ACCESS_TOKEN=your-token

# Redshift
CENSUS_TEST_REDSHIFT_HOST=your-cluster.region.redshift.amazonaws.com
CENSUS_TEST_REDSHIFT_PORT=5439
CENSUS_TEST_REDSHIFT_DATABASE=dev
CENSUS_TEST_REDSHIFT_USERNAME=testuser
CENSUS_TEST_REDSHIFT_PASSWORD=your-password

# Salesforce
CENSUS_TEST_SALESFORCE_USERNAME=test@example.com
CENSUS_TEST_SALESFORCE_INSTANCE_URL=https://your-sandbox.my.salesforce.com
CENSUS_TEST_SALESFORCE_CLIENT_ID=3MVG9...
CENSUS_TEST_SALESFORCE_JWT_SIGNING_KEY='-----BEGIN RSA PRIVATE KEY-----\n...'
CENSUS_TEST_SALESFORCE_DOMAIN=test.salesforce.com
```

## Manual Testing

Use the examples directory to test manually:

```bash
cd examples/complete-census-setup/
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars
terraform init
terraform plan
terraform apply
```

## Troubleshooting

**Connection errors:**
- Verify network access (security groups, firewall)
- Check credentials are correct
- Ensure using staging environment, not production

**Salesforce errors:**
- Confirm using sandbox (domain: test.salesforce.com)
- Verify Connected App OAuth scopes
- Check JWT key format includes `\n` for newlines
- Ensure user has API access

**Test failures:**
- Verify `.env.test` exists and is formatted correctly
- Check `TF_ACC=1` is set
- Review test output for specific errors
- Increase timeout if needed: `-timeout 120m`

## CI/CD

The provider includes GitHub Actions workflows for automated testing. See `.github/workflows/test.yml` for configuration.
