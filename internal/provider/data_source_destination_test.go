package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceDestination_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceDestinationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.census_destination.test", "name", "Test Salesforce Destination"),
					resource.TestCheckResourceAttr("data.census_destination.test", "type", "salesforce"),
					resource.TestCheckResourceAttrSet("data.census_destination.test", "id"),
					resource.TestCheckResourceAttrSet("data.census_destination.test", "status"),
					resource.TestCheckResourceAttrSet("data.census_destination.test", "created_at"),
				),
			},
		},
	})
}

func testAccDataSourceDestinationConfig_basic() string {
	return `
resource "census_destination" "test" {
  name = "Test Salesforce Destination"
  type = "salesforce"
  
  connection_config = {
    username       = "test@example.com"
    password       = "test_password"
    security_token = "test_token"
    sandbox        = "true"
  }
}

data "census_destination" "test" {
  id = census_destination.test.id
}
`
}
