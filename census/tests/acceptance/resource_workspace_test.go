package acceptance

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// census_workspace is frozen: it stays in the provider's schema so existing
// state keeps working, but it can no longer be created or managed.
func TestAccResourceWorkspace_CreateFails(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { TestAccPreCheckIntegration(t) },
		Providers: TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
resource "census_workspace" "test" {
  name = "test-workspace"
}
`,
				ExpectError: regexp.MustCompile(`census_workspace can no longer be created`),
			},
		},
	})
}
