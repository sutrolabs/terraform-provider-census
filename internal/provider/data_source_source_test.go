package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceSource_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.census_source.test", "name", "Test Postgres Source"),
					resource.TestCheckResourceAttr("data.census_source.test", "type", "postgres"),
					resource.TestCheckResourceAttrSet("data.census_source.test", "id"),
					resource.TestCheckResourceAttrSet("data.census_source.test", "status"),
					resource.TestCheckResourceAttrSet("data.census_source.test", "created_at"),
				),
			},
		},
	})
}

func testAccDataSourceSourceConfig_basic() string {
	return `
resource "census_source" "test" {
  name = "Test Postgres Source"
  type = "postgres"
  
  connection_config = {
    host     = "test-host.com"
    port     = "5432"
    username = "test_user"
    password = "test_password"
    database = "test_db"
  }
}

data "census_source" "test" {
  id = census_source.test.id
}
`
}