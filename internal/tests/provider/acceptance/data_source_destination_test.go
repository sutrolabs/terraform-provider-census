package acceptance

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	provider_test "github.com/sutrolabs/terraform-provider-census/internal/tests/provider"
)

func TestAccDataSourceDestination_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { provider_test.TestAccPreCheckIntegration(t) },
		Providers: provider_test.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceDestinationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.census_destination.test", "name", "Test Salesforce Destination"),
					resource.TestCheckResourceAttr("data.census_destination.test", "type", "salesforce"),
					resource.TestCheckResourceAttrSet("data.census_destination.test", "id"),
					resource.TestCheckResourceAttrSet("data.census_destination.test", "status"),
					resource.TestCheckResourceAttrSet("data.census_destination.test", "created_at"),
					resource.TestCheckResourceAttrPair("data.census_destination.test", "id", "census_destination.test", "id"),
					resource.TestCheckResourceAttrPair("data.census_destination.test", "workspace_id", "census_destination.test", "workspace_id"),
				),
			},
		},
	})
}

func testAccDataSourceDestinationConfig_basic() string {
	return fmt.Sprintf(`
provider "census" {
  base_url = "%s"
}

resource "census_workspace" "test" {
  name = "Test Workspace - Data Source"
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
}

data "census_destination" "test" {
  id = census_destination.test.id
  workspace_id = census_workspace.test.id
}
`,
		os.Getenv("CENSUS_BASE_URL"),
		os.Getenv("CENSUS_TEST_SALESFORCE_USERNAME"),
		os.Getenv("CENSUS_TEST_SALESFORCE_PASSWORD"),
		os.Getenv("CENSUS_TEST_SALESFORCE_SECURITY_TOKEN"),
		getEnvOrDefault("CENSUS_TEST_SALESFORCE_SANDBOX", "true"),
	)
}
