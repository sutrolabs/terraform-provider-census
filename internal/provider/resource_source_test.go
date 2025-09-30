package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceSource_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("census_source.test", "name", "Test Postgres Source"),
					resource.TestCheckResourceAttr("census_source.test", "type", "postgres"),
					resource.TestCheckResourceAttrSet("census_source.test", "id"),
					resource.TestCheckResourceAttrSet("census_source.test", "status"),
					resource.TestCheckResourceAttrSet("census_source.test", "created_at"),
				),
			},
		},
	})
}

func TestAccResourceSource_Update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("census_source.test", "name", "Test Postgres Source"),
				),
			},
			{
				Config: testAccResourceSourceConfig_updated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("census_source.test", "name", "Updated Postgres Source"),
				),
			},
		},
	})
}

func testAccResourceSourceConfig_basic() string {
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

  auto_refresh_tables = false
}
`
}

func testAccResourceSourceConfig_updated() string {
	return `
resource "census_source" "test" {
  name = "Updated Postgres Source"
  type = "postgres"
  
  connection_config = {
    host     = "updated-host.com"
    port     = "5432"
    username = "updated_user"
    password = "updated_password"
    database = "updated_db"
  }

  auto_refresh_tables = true
}
`
}
