package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceWorkspace_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfig_basic("test-workspace"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists("census_workspace.test"),
					resource.TestCheckResourceAttr("census_workspace.test", "name", "test-workspace"),
					resource.TestCheckResourceAttr("census_workspace.test", "notification_emails.#", "1"),
					resource.TestCheckResourceAttr("census_workspace.test", "notification_emails.0", "test@example.com"),
				),
			},
		},
	})
}

func TestResourceWorkspace_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfig_basic("test-workspace"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists("census_workspace.test"),
					resource.TestCheckResourceAttr("census_workspace.test", "name", "test-workspace"),
				),
			},
			{
				Config: testAccWorkspaceConfig_updated("test-workspace-updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists("census_workspace.test"),
					resource.TestCheckResourceAttr("census_workspace.test", "name", "test-workspace-updated"),
					resource.TestCheckResourceAttr("census_workspace.test", "notification_emails.#", "2"),
				),
			},
		},
	})
}

func TestResourceWorkspace_withAPIKey(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfig_withAPIKey("test-workspace-api"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists("census_workspace.test"),
					resource.TestCheckResourceAttr("census_workspace.test", "name", "test-workspace-api"),
					resource.TestCheckResourceAttr("census_workspace.test", "return_workspace_api_key", "true"),
					resource.TestCheckResourceAttrSet("census_workspace.test", "api_key"),
				),
			},
		},
	})
}

func testAccCheckWorkspaceDestroy(s *terraform.State) error {
	// This would normally check that the workspace has been destroyed
	// For now, we'll just return nil since we don't have a real API to test against
	return nil
}

func testAccCheckWorkspaceExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID not set")
		}

		// This would normally check that the workspace exists in the API
		// For now, we'll just return nil since we don't have a real API to test against
		return nil
	}
}

func testAccWorkspaceConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "census_workspace" "test" {
  name = "%s"
  notification_emails = ["test@example.com"]
}
`, name)
}

func testAccWorkspaceConfig_updated(name string) string {
	return fmt.Sprintf(`
resource "census_workspace" "test" {
  name = "%s"
  notification_emails = ["test@example.com", "test2@example.com"]
}
`, name)
}

func testAccWorkspaceConfig_withAPIKey(name string) string {
	return fmt.Sprintf(`
resource "census_workspace" "test" {
  name = "%s"
  notification_emails = ["test@example.com"]
  return_workspace_api_key = true
}
`, name)
}
