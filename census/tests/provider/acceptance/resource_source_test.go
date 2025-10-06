package acceptance

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	provider_test "github.com/sutrolabs/terraform-provider-census/census/tests/provider"
)

func TestAccResourceSource_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { provider_test.TestAccPreCheckIntegration(t) },
		Providers: provider_test.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSourceConfig_redshift(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("census_source.test", "name", "Test Redshift Source"),
					resource.TestCheckResourceAttr("census_source.test", "type", "redshift"),
					resource.TestCheckResourceAttrSet("census_source.test", "id"),
					resource.TestCheckResourceAttrSet("census_source.test", "workspace_id"),
					resource.TestCheckResourceAttrSet("census_source.test", "created_at"),
				),
			},
		},
	})
}

func TestAccResourceSource_Update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { provider_test.TestAccPreCheckIntegration(t) },
		Providers: provider_test.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSourceConfig_redshift(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("census_source.test", "name", "Test Redshift Source"),
					resource.TestCheckResourceAttr("census_source.test", "auto_refresh_tables", "false"),
				),
			},
			{
				Config: testAccResourceSourceConfig_redshiftUpdated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("census_source.test", "name", "Updated Redshift Source"),
					resource.TestCheckResourceAttr("census_source.test", "auto_refresh_tables", "true"),
				),
			},
		},
	})
}

func testAccResourceSourceConfig_redshift() string {
	return fmt.Sprintf(`
resource "census_workspace" "test" {
  name = "Test Workspace - Source"
  notification_emails = ["test@example.com"]
}

resource "census_source" "test" {
  workspace_id = census_workspace.test.id
  name = "Test Redshift Source"
  type = "redshift"

  connection_config = {
    hostname = "%s"
    port     = "%s"
    database = "%s"
    user     = "%s"
    password = "%s"
  }

  auto_refresh_tables = false
}
`,
		os.Getenv("CENSUS_TEST_REDSHIFT_HOST"),
		getEnvOrDefault("CENSUS_TEST_REDSHIFT_PORT", "5439"),
		os.Getenv("CENSUS_TEST_REDSHIFT_DATABASE"),
		os.Getenv("CENSUS_TEST_REDSHIFT_USERNAME"),
		os.Getenv("CENSUS_TEST_REDSHIFT_PASSWORD"),
	)
}

func testAccResourceSourceConfig_redshiftUpdated() string {
	return fmt.Sprintf(`
resource "census_workspace" "test" {
  name = "Test Workspace - Source"
  notification_emails = ["test@example.com"]
}

resource "census_source" "test" {
  workspace_id = census_workspace.test.id
  name = "Updated Redshift Source"
  type = "redshift"

  connection_config = {
    hostname = "%s"
    port     = "%s"
    database = "%s"
    user     = "%s"
    password = "%s"
  }

  auto_refresh_tables = true
}
`,
		os.Getenv("CENSUS_TEST_REDSHIFT_HOST"),
		getEnvOrDefault("CENSUS_TEST_REDSHIFT_PORT", "5439"),
		os.Getenv("CENSUS_TEST_REDSHIFT_DATABASE"),
		os.Getenv("CENSUS_TEST_REDSHIFT_USERNAME"),
		os.Getenv("CENSUS_TEST_REDSHIFT_PASSWORD"),
	)
}

// Helper function to get environment variable with default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
