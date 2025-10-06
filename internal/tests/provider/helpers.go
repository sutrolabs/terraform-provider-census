package provider_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/sutrolabs/terraform-provider-census/internal/provider"
)

// TestAccProviders is a shared map of providers used across all acceptance tests
var TestAccProviders map[string]*schema.Provider

// TestAccProvider is a single provider instance for tests
var TestAccProvider *schema.Provider

func init() {
	TestAccProvider = provider.Provider()
	TestAccProviders = map[string]*schema.Provider{
		"census": TestAccProvider,
	}
}

// TestAccPreCheck checks for required environment variables for all acceptance tests
func TestAccPreCheck(t *testing.T) {
	// Check for required environment variables
	if v := os.Getenv("CENSUS_PERSONAL_ACCESS_TOKEN"); v == "" {
		if v := os.Getenv("CENSUS_WORKSPACE_ACCESS_TOKEN"); v == "" {
			t.Fatal("CENSUS_PERSONAL_ACCESS_TOKEN or CENSUS_WORKSPACE_ACCESS_TOKEN must be set for acceptance tests")
		}
	}

	// You can add additional pre-check logic here
	// For example, checking API connectivity, required permissions, etc.
}

// TestAccPreCheckIntegration checks for credentials needed for integration tests
// Integration tests create real resources (workspace, source, destination, sync) in staging
func TestAccPreCheckIntegration(t *testing.T) {
	TestAccPreCheck(t)

	// Check for Redshift credentials
	requiredRedshift := []string{
		"CENSUS_TEST_REDSHIFT_HOST",
		"CENSUS_TEST_REDSHIFT_DATABASE",
		"CENSUS_TEST_REDSHIFT_USERNAME",
		"CENSUS_TEST_REDSHIFT_PASSWORD",
	}

	for _, envVar := range requiredRedshift {
		if v := os.Getenv(envVar); v == "" {
			t.Fatalf("%s must be set for integration tests. See .env.test.example for setup instructions.", envVar)
		}
	}

	// Check for Salesforce JWT OAuth credentials
	requiredSalesforce := []string{
		"CENSUS_TEST_SALESFORCE_USERNAME",
		"CENSUS_TEST_SALESFORCE_INSTANCE_URL",
		"CENSUS_TEST_SALESFORCE_CLIENT_ID",
		"CENSUS_TEST_SALESFORCE_JWT_SIGNING_KEY",
		"CENSUS_TEST_SALESFORCE_DOMAIN",
	}

	for _, envVar := range requiredSalesforce {
		if v := os.Getenv(envVar); v == "" {
			t.Fatalf("%s must be set for integration tests. See .env.test.example for JWT OAuth setup instructions.", envVar)
		}
	}
}
