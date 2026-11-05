package acceptance

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	provider_test "github.com/sutrolabs/terraform-provider-census/census/tests/provider"
)

// census_destination is frozen: it stays in the provider's schema so existing
// state keeps working, but it can no longer be created or managed.
func TestAccResourceDestination_CreateFails(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { provider_test.TestAccPreCheckIntegration(t) },
		Providers: provider_test.TestAccProviders,
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
