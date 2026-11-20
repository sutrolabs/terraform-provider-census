package acceptance

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// census_destination is frozen: it stays in the provider's schema so existing
// state keeps working, but it can no longer be created or managed.
func TestAccResourceDestination_CreateFails(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { TestAccPreCheckIntegration(t) },
		Providers: TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
resource "census_destination" "test" {
  name = "Test Salesforce Destination"
  type = "salesforce"

  connection_config = {}
}
`,
				ExpectError: regexp.MustCompile(`census_destination can no longer be created`),
			},
		},
	})
}
