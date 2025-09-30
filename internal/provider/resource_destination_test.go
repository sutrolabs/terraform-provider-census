package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceDestination_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDestinationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("census_destination.test", "name", "Test Salesforce Destination"),
					resource.TestCheckResourceAttr("census_destination.test", "type", "salesforce"),
					resource.TestCheckResourceAttrSet("census_destination.test", "id"),
					resource.TestCheckResourceAttrSet("census_destination.test", "status"),
					resource.TestCheckResourceAttrSet("census_destination.test", "created_at"),
				),
			},
		},
	})
}

func TestAccResourceDestination_Update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDestinationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("census_destination.test", "name", "Test Salesforce Destination"),
				),
			},
			{
				Config: testAccResourceDestinationConfig_updated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("census_destination.test", "name", "Updated Salesforce Destination"),
				),
			},
		},
	})
}

func testAccResourceDestinationConfig_basic() string {
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

  auto_refresh_objects = false
}
`
}

func testAccResourceDestinationConfig_updated() string {
	return `
resource "census_destination" "test" {
  name = "Updated Salesforce Destination"
  type = "salesforce"
  
  connection_config = {
    username       = "updated@example.com"
    password       = "updated_password"
    security_token = "updated_token"
    sandbox        = "false"
  }

  auto_refresh_objects = true
}
`
}
