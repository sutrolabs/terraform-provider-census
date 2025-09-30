package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"census": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ *schema.Provider = Provider()
}

func testAccPreCheck(t *testing.T) {
	// Check for required environment variables
	if v := os.Getenv("CENSUS_PERSONAL_ACCESS_TOKEN"); v == "" {
		if v := os.Getenv("CENSUS_WORKSPACE_ACCESS_TOKEN"); v == "" {
			t.Fatal("CENSUS_PERSONAL_ACCESS_TOKEN or CENSUS_WORKSPACE_ACCESS_TOKEN must be set for acceptance tests")
		}
	}

	// You can add additional pre-check logic here
	// For example, checking API connectivity, required permissions, etc.
}
