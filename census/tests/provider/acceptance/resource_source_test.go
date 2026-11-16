package acceptance

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	provider_test "github.com/sutrolabs/terraform-provider-census/census/tests/provider"
)

func TestAccResourceSource_Basic(t *testing.T) {
	rName := acctest.RandString(6)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { provider_test.TestAccPreCheckIntegration(t) },
		Providers: provider_test.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSourceConfig_redshift(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("census_source.test", "name", fmt.Sprintf("Test_Redshift_Source_%s", rName)),
					resource.TestCheckResourceAttr("census_source.test", "type", "redshift"),
					resource.TestCheckResourceAttrSet("census_source.test", "id"),
					resource.TestCheckResourceAttrSet("census_source.test", "group_id"),
					resource.TestCheckResourceAttrSet("census_source.test", "created_at"),
				),
			},
		},
	})
}

func TestAccResourceSource_Update(t *testing.T) {
	rName := acctest.RandString(6)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { provider_test.TestAccPreCheckIntegration(t) },
		Providers: provider_test.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSourceConfig_redshift(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("census_source.test", "name", fmt.Sprintf("Test_Redshift_Source_%s", rName)),
					resource.TestCheckResourceAttr("census_source.test", "auto_refresh_tables", "false"),
				),
			},
			{
				Config: testAccResourceSourceConfig_redshiftUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("census_source.test", "name", fmt.Sprintf("Updated_Redshift_Source_%s", rName)),
					resource.TestCheckResourceAttr("census_source.test", "auto_refresh_tables", "true"),
				),
			},
		},
	})
}

func testAccResourceSourceConfig_redshift(rName string) string {
	return fmt.Sprintf(`
resource "census_source" "test" {
  name = "Test_Redshift_Source_%s"
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
		rName,
		os.Getenv("CENSUS_TEST_REDSHIFT_HOST"),
		getEnvOrDefault("CENSUS_TEST_REDSHIFT_PORT", "5439"),
		os.Getenv("CENSUS_TEST_REDSHIFT_DATABASE"),
		os.Getenv("CENSUS_TEST_REDSHIFT_USERNAME"),
		os.Getenv("CENSUS_TEST_REDSHIFT_PASSWORD"),
	)
}

func testAccResourceSourceConfig_redshiftUpdated(rName string) string {
	return fmt.Sprintf(`
resource "census_source" "test" {
  name = "Updated_Redshift_Source_%s"
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
		rName,
		os.Getenv("CENSUS_TEST_REDSHIFT_HOST"),
		getEnvOrDefault("CENSUS_TEST_REDSHIFT_PORT", "5439"),
		os.Getenv("CENSUS_TEST_REDSHIFT_DATABASE"),
		os.Getenv("CENSUS_TEST_REDSHIFT_USERNAME"),
		os.Getenv("CENSUS_TEST_REDSHIFT_PASSWORD"),
	)
}

func TestAccResourceSource_Import(t *testing.T) {
	rName := acctest.RandString(6)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { provider_test.TestAccPreCheckIntegration(t) },
		Providers: provider_test.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSourceConfig_redshift(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("census_source.test", "name", fmt.Sprintf("Test_Redshift_Source_%s", rName)),
					resource.TestCheckResourceAttr("census_source.test", "type", "redshift"),
				),
			},
			{
				ResourceName:      "census_source.test",
				ImportState:       true,
				ImportStateVerify: true,
				// connection_config contains sensitive data that isn't returned by the API
				ImportStateVerifyIgnore: []string{"connection_config", "auto_refresh_tables"},
			},
		},
	})
}

// Helper function to get environment variable with default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
