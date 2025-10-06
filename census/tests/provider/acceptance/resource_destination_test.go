package acceptance

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	provider_test "github.com/sutrolabs/terraform-provider-census/census/tests/provider"
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
resource "census_workspace" "test" {
  name = "Test Workspace - Destination"
  notification_emails = ["test@example.com"]
}

resource "census_destination" "test" {
  workspace_id = census_workspace.test.id
  name = "Test Salesforce Destination"
  type = "salesforce"

  connection_config = {
    username        = "%s"
    instance_url    = "%s"
    client_id       = "%s"
    jwt_signing_key = "%s"
    domain          = "%s"
  }

  auto_refresh_objects = false
}
`,
		os.Getenv("CENSUS_TEST_SALESFORCE_USERNAME"),
		os.Getenv("CENSUS_TEST_SALESFORCE_INSTANCE_URL"),
		os.Getenv("CENSUS_TEST_SALESFORCE_CLIENT_ID"),
		os.Getenv("CENSUS_TEST_SALESFORCE_JWT_SIGNING_KEY"),
		os.Getenv("CENSUS_TEST_SALESFORCE_DOMAIN"),
	)
}

func testAccResourceDestinationConfig_salesforceUpdated() string {
	return fmt.Sprintf(`
resource "census_workspace" "test" {
  name = "Test Workspace - Destination"
  notification_emails = ["test@example.com"]
}

resource "census_destination" "test" {
  workspace_id = census_workspace.test.id
  name = "Updated Salesforce Destination"
  type = "salesforce"

  connection_config = {
    username        = "%s"
    instance_url    = "%s"
    client_id       = "%s"
    jwt_signing_key = "%s"
    domain          = "%s"
  }

  auto_refresh_objects = true
}
`,
		os.Getenv("CENSUS_TEST_SALESFORCE_USERNAME"),
		os.Getenv("CENSUS_TEST_SALESFORCE_INSTANCE_URL"),
		os.Getenv("CENSUS_TEST_SALESFORCE_CLIENT_ID"),
		os.Getenv("CENSUS_TEST_SALESFORCE_JWT_SIGNING_KEY"),
		os.Getenv("CENSUS_TEST_SALESFORCE_DOMAIN"),
	)
}
