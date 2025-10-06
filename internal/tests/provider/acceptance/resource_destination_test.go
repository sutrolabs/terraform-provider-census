package acceptance

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	provider_test "github.com/sutrolabs/terraform-provider-census/internal/tests/provider"
)

func TestAccResourceDestination_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { provider_test.TestAccPreCheckIntegration(t) },
		Providers: provider_test.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDestinationConfig_salesforce(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("census_destination.test", "name", "Test Salesforce Destination"),
					resource.TestCheckResourceAttr("census_destination.test", "type", "salesforce"),
					resource.TestCheckResourceAttrSet("census_destination.test", "id"),
					resource.TestCheckResourceAttrSet("census_destination.test", "workspace_id"),
					resource.TestCheckResourceAttrSet("census_destination.test", "status"),
					resource.TestCheckResourceAttrSet("census_destination.test", "created_at"),
				),
			},
		},
	})
}

func TestAccResourceDestination_Update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { provider_test.TestAccPreCheckIntegration(t) },
		Providers: provider_test.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDestinationConfig_salesforce(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("census_destination.test", "name", "Test Salesforce Destination"),
					resource.TestCheckResourceAttr("census_destination.test", "auto_refresh_objects", "false"),
				),
			},
			{
				Config: testAccResourceDestinationConfig_salesforceUpdated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("census_destination.test", "name", "Updated Salesforce Destination"),
					resource.TestCheckResourceAttr("census_destination.test", "auto_refresh_objects", "true"),
				),
			},
		},
	})
}

func testAccResourceDestinationConfig_salesforce() string {
	return fmt.Sprintf(`
provider "census" {
  base_url = "%s"
}

resource "census_workspace" "test" {
  name = "Test Workspace - Destination"
  notification_emails = ["test@example.com"]
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

  auto_refresh_objects = false
}
`,
		os.Getenv("CENSUS_BASE_URL"),
		os.Getenv("CENSUS_TEST_SALESFORCE_USERNAME"),
		os.Getenv("CENSUS_TEST_SALESFORCE_PASSWORD"),
		os.Getenv("CENSUS_TEST_SALESFORCE_SECURITY_TOKEN"),
		getEnvOrDefault("CENSUS_TEST_SALESFORCE_SANDBOX", "true"),
	)
}

func testAccResourceDestinationConfig_salesforceUpdated() string {
	return fmt.Sprintf(`
provider "census" {
  base_url = "%s"
}

resource "census_workspace" "test" {
  name = "Test Workspace - Destination"
  notification_emails = ["test@example.com"]
}

resource "census_destination" "test" {
  workspace_id = census_workspace.test.id
  name = "Updated Salesforce Destination"
  type = "salesforce"

  connection_config = {
    username       = "%s"
    password       = "%s"
    security_token = "%s"
    sandbox        = "%s"
  }

  auto_refresh_objects = true
}
`,
		os.Getenv("CENSUS_BASE_URL"),
		os.Getenv("CENSUS_TEST_SALESFORCE_USERNAME"),
		os.Getenv("CENSUS_TEST_SALESFORCE_PASSWORD"),
		os.Getenv("CENSUS_TEST_SALESFORCE_SECURITY_TOKEN"),
		getEnvOrDefault("CENSUS_TEST_SALESFORCE_SANDBOX", "true"),
	)
}
