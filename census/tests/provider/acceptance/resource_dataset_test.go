package acceptance

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	provider_test "github.com/sutrolabs/terraform-provider-census/census/tests/provider"
)

func TestAccResourceDataset_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { provider_test.TestAccPreCheckIntegration(t) },
		Providers: provider_test.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDatasetConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("census_dataset.test", "name", "Test Active Users Dataset"),
					resource.TestCheckResourceAttr("census_dataset.test", "type", "sql"),
					resource.TestCheckResourceAttr("census_dataset.test", "description", "Dataset for testing purposes"),
					resource.TestCheckResourceAttrSet("census_dataset.test", "id"),
					resource.TestCheckResourceAttrSet("census_dataset.test", "workspace_id"),
					resource.TestCheckResourceAttrSet("census_dataset.test", "source_id"),
					resource.TestCheckResourceAttrSet("census_dataset.test", "query"),
					resource.TestCheckResourceAttrSet("census_dataset.test", "created_at"),
				),
			},
		},
	})
}

func TestAccResourceDataset_Update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { provider_test.TestAccPreCheckIntegration(t) },
		Providers: provider_test.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDatasetConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("census_dataset.test", "name", "Test Active Users Dataset"),
					resource.TestCheckResourceAttr("census_dataset.test", "description", "Dataset for testing purposes"),
				),
			},
			{
				Config: testAccResourceDatasetConfig_updated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("census_dataset.test", "name", "Updated Users Dataset"),
					resource.TestCheckResourceAttr("census_dataset.test", "description", "Updated dataset description"),
				),
			},
		},
	})
}

func TestAccResourceDataset_WithSync(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { provider_test.TestAccPreCheckIntegration(t) },
		Providers: provider_test.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDatasetConfig_withSync(),
				Check: resource.ComposeTestCheckFunc(
					// Check dataset
					resource.TestCheckResourceAttr("census_dataset.test", "name", "All Users Dataset"),
					resource.TestCheckResourceAttr("census_dataset.test", "type", "sql"),
					resource.TestCheckResourceAttrSet("census_dataset.test", "id"),

					// Check sync using dataset
					resource.TestCheckResourceAttr("census_sync.dataset_sync", "label", "Dataset to Contacts Sync"),
					resource.TestCheckResourceAttr("census_sync.dataset_sync", "operation", "upsert"),
					resource.TestCheckResourceAttrSet("census_sync.dataset_sync", "id"),
					resource.TestCheckResourceAttr("census_sync.dataset_sync", "paused", "true"),
				),
			},
		},
	})
}

func testAccResourceDatasetConfig_basic() string {
	return fmt.Sprintf(`
resource "census_workspace" "test" {
  name = "Test Workspace - Dataset"
  notification_emails = ["test@example.com"]
}

resource "census_source" "test" {
  workspace_id = census_workspace.test.id
  name = "Test Redshift Source"
  type = "redshift"

  connection_config = {
    hostname = "%s"
    port     = "%s"
    database = "%s"
    user     = "%s"
    password = "%s"
  }

  auto_refresh_tables = false
}

resource "census_dataset" "test" {
  workspace_id = census_workspace.test.id
  name         = "Test Active Users Dataset"
  type         = "sql"
  description  = "Dataset for testing purposes"
  source_id    = census_source.test.id

  query = <<-SQL
    SELECT
      id,
      email,
      first_name,
      last_name,
      created_at
    FROM users
    WHERE active = true
  SQL
}
`,
		os.Getenv("CENSUS_TEST_REDSHIFT_HOST"),
		getEnvOrDefault("CENSUS_TEST_REDSHIFT_PORT", "5439"),
		os.Getenv("CENSUS_TEST_REDSHIFT_DATABASE"),
		os.Getenv("CENSUS_TEST_REDSHIFT_USERNAME"),
		os.Getenv("CENSUS_TEST_REDSHIFT_PASSWORD"),
	)
}

func testAccResourceDatasetConfig_updated() string {
	return fmt.Sprintf(`
resource "census_workspace" "test" {
  name = "Test Workspace - Dataset"
  notification_emails = ["test@example.com"]
}

resource "census_source" "test" {
  workspace_id = census_workspace.test.id
  name = "Test Redshift Source"
  type = "redshift"

  connection_config = {
    hostname = "%s"
    port     = "%s"
    database = "%s"
    user     = "%s"
    password = "%s"
  }

  auto_refresh_tables = false
}

resource "census_dataset" "test" {
  workspace_id = census_workspace.test.id
  name         = "Updated Users Dataset"
  type         = "sql"
  description  = "Updated dataset description"
  source_id    = census_source.test.id

  query = <<-SQL
    SELECT
      id,
      email,
      first_name,
      last_name,
      created_at,
      updated_at
    FROM users
    WHERE active = true
  SQL
}
`,
		os.Getenv("CENSUS_TEST_REDSHIFT_HOST"),
		getEnvOrDefault("CENSUS_TEST_REDSHIFT_PORT", "5439"),
		os.Getenv("CENSUS_TEST_REDSHIFT_DATABASE"),
		os.Getenv("CENSUS_TEST_REDSHIFT_USERNAME"),
		os.Getenv("CENSUS_TEST_REDSHIFT_PASSWORD"),
	)
}

func testAccResourceDatasetConfig_withSync() string {
	return fmt.Sprintf(`
resource "census_workspace" "test" {
  name = "Test Workspace - Dataset Sync"
  notification_emails = ["test@example.com"]
}

resource "census_source" "test" {
  workspace_id = census_workspace.test.id
  name = "Test Redshift Source"
  type = "redshift"

  connection_config = {
    hostname = "%s"
    port     = "%s"
    database = "%s"
    user     = "%s"
    password = "%s"
  }

  auto_refresh_tables = false
}

resource "census_destination" "test" {
  workspace_id = census_workspace.test.id
  name         = "Test Salesforce Destination"
  type         = "salesforce"

  connection_config = {
    username        = "%s"
    instance_url    = "%s"
    client_id       = "%s"
    jwt_signing_key = "%s"
    domain          = "%s"
  }
}

resource "census_dataset" "test" {
  workspace_id = census_workspace.test.id
  name         = "All Users Dataset"
  type         = "sql"
  description  = "Simple dataset with all user data for syncing"
  source_id    = census_source.test.id

  query = <<-SQL
    SELECT
      user_id,
      email,
      first_name,
      last_name
    FROM users
  SQL
}

resource "census_sync" "dataset_sync" {
  workspace_id = census_workspace.test.id
  label        = "Dataset to Contacts Sync"

  # Source configuration - use a dataset instead of table
  source_attributes {
    connection_id = census_source.test.id
    object {
      type = "dataset"
      id   = census_dataset.test.id
    }
  }

  # Destination configuration - Salesforce Contacts
  destination_attributes {
    connection_id = census_destination.test.id
    object        = "Contact"
  }

  operation = "upsert"

  # Field mappings using dataset columns
  field_mapping {
    from                  = "email"
    to                    = "Email"
    is_primary_identifier = true
  }

  field_mapping {
    from = "first_name"
    to   = "FirstName"
  }

  field_mapping {
    from = "last_name"
    to   = "LastName"
  }

  field_mapping {
    from = "user_id"
    to   = "Census_ID__c"
  }

  run_mode {
    type = "triggered"
    triggers {
      schedule {
        frequency = "never"
      }
    }
  }

  alert {
    type                 = "FailureAlertConfiguration"
    send_for             = "first_time"
    should_send_recovery = true
  }

  alert {
    type                 = "InvalidRecordPercentAlertConfiguration"
    send_for             = "first_time"
    should_send_recovery = true
    options = {
      threshold = "75"
    }
  }

  paused = true
}
`,
		os.Getenv("CENSUS_TEST_REDSHIFT_HOST"),
		getEnvOrDefault("CENSUS_TEST_REDSHIFT_PORT", "5439"),
		os.Getenv("CENSUS_TEST_REDSHIFT_DATABASE"),
		os.Getenv("CENSUS_TEST_REDSHIFT_USERNAME"),
		os.Getenv("CENSUS_TEST_REDSHIFT_PASSWORD"),
		os.Getenv("CENSUS_TEST_SALESFORCE_USERNAME"),
		os.Getenv("CENSUS_TEST_SALESFORCE_INSTANCE_URL"),
		os.Getenv("CENSUS_TEST_SALESFORCE_CLIENT_ID"),
		os.Getenv("CENSUS_TEST_SALESFORCE_JWT_SIGNING_KEY"),
		os.Getenv("CENSUS_TEST_SALESFORCE_DOMAIN"),
	)
}

func TestAccResourceDataset_Import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { provider_test.TestAccPreCheckIntegration(t) },
		Providers: provider_test.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDatasetConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("census_dataset.test", "name", "Test Active Users Dataset"),
					resource.TestCheckResourceAttr("census_dataset.test", "type", "sql"),
				),
			},
			{
				ResourceName:      "census_dataset.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccDatasetImportStateIdFunc("census_dataset.test"),
			},
		},
	})
}

// Helper to construct composite ID for import (workspace_id:dataset_id)
func testAccDatasetImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s:%s",
			rs.Primary.Attributes["workspace_id"],
			rs.Primary.ID), nil
	}
}
