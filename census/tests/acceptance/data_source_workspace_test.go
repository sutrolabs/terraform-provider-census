package acceptance

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// census_workspace is frozen: the data source makes no API call and returns
// only the id it was given (plus a warning).
func TestAccDataSourceWorkspace_Frozen(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { TestAccPreCheckIntegration(t) },
		Providers: TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
data "census_workspace" "test" {
  id = "12345"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.census_workspace.test", "id", "12345"),
					resource.TestCheckResourceAttr("data.census_workspace.test", "name", ""),
				),
			},
		},
	})
}
