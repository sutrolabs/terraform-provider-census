package acceptance

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	provider_test "github.com/sutrolabs/terraform-provider-census/census/tests/provider"
)

// TestAccResourceSync_Basic tests basic sync creation with minimal configuration
func TestAccResourceSync_Basic(t *testing.T) {
	rName := acctest.RandString(6)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { provider_test.TestAccPreCheckIntegration(t) },
		Providers: provider_test.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSyncConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("census_sync.test", "label", "Test Basic Sync"),
					resource.TestCheckResourceAttr("census_sync.test", "operation", "upsert"),
					resource.TestCheckResourceAttrSet("census_sync.test", "id"),
					resource.TestCheckResourceAttrSet("census_sync.test", "workspace_id"),
					resource.TestCheckResourceAttr("census_sync.test", "paused", "true"),
				),
			},
		},
	})
}

// TestAccResourceSync_Update tests updating sync configuration
func TestAccResourceSync_Update(t *testing.T) {
	rName := acctest.RandString(6)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { provider_test.TestAccPreCheckIntegration(t) },
		Providers: provider_test.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSyncConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("census_sync.test", "label", "Test Basic Sync"),
					resource.TestCheckResourceAttr("census_sync.test", "paused", "true"),
				),
			},
			{
				Config: testAccResourceSyncConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("census_sync.test", "label", "Updated Sync Label"),
					resource.TestCheckResourceAttr("census_sync.test", "paused", "false"),
				),
			},
		},
	})
}

// TestAccResourceSync_FieldMappings tests various field mapping types
func TestAccResourceSync_FieldMappings(t *testing.T) {
	rName := acctest.RandString(6)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { provider_test.TestAccPreCheckIntegration(t) },
		Providers: provider_test.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSyncConfig_fieldMappings(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("census_sync.test", "label", "Test Field Mappings"),
					resource.TestCheckResourceAttr("census_sync.test", "field_mapping.#", "5"),
					// Direct mapping
					resource.TestCheckResourceAttr("census_sync.test", "field_mapping.0.from", "email"),
					resource.TestCheckResourceAttr("census_sync.test", "field_mapping.0.to", "Email"),
					resource.TestCheckResourceAttr("census_sync.test", "field_mapping.0.is_primary_identifier", "true"),
					// Constant mapping
					resource.TestCheckResourceAttr("census_sync.test", "field_mapping.3.type", "constant"),
					resource.TestCheckResourceAttr("census_sync.test", "field_mapping.3.to", "LeadSource"),
				),
			},
		},
	})
}

// TestAccResourceSync_RunMode_Daily tests daily schedule configuration
func TestAccResourceSync_RunMode_Daily(t *testing.T) {
	rName := acctest.RandString(6)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { provider_test.TestAccPreCheckIntegration(t) },
		Providers: provider_test.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSyncConfig_runModeDaily(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("census_sync.test", "label", "Test Daily Schedule"),
					resource.TestCheckResourceAttr("census_sync.test", "run_mode.0.type", "triggered"),
					resource.TestCheckResourceAttr("census_sync.test", "run_mode.0.triggers.0.schedule.0.frequency", "daily"),
					resource.TestCheckResourceAttr("census_sync.test", "run_mode.0.triggers.0.schedule.0.hour", "6"),
					resource.TestCheckResourceAttr("census_sync.test", "run_mode.0.triggers.0.schedule.0.minute", "0"),
				),
			},
		},
	})
}

// TestAccResourceSync_RunMode_Hourly tests hourly schedule configuration
func TestAccResourceSync_RunMode_Hourly(t *testing.T) {
	rName := acctest.RandString(6)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { provider_test.TestAccPreCheckIntegration(t) },
		Providers: provider_test.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSyncConfig_runModeHourly(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("census_sync.test", "label", "Test Hourly Schedule"),
					resource.TestCheckResourceAttr("census_sync.test", "run_mode.0.type", "triggered"),
					resource.TestCheckResourceAttr("census_sync.test", "run_mode.0.triggers.0.schedule.0.frequency", "hourly"),
					resource.TestCheckResourceAttr("census_sync.test", "run_mode.0.triggers.0.schedule.0.minute", "30"),
				),
			},
		},
	})
}

// TestAccResourceSync_RunMode_Manual tests never/manual schedule configuration (triggered manually)
func TestAccResourceSync_RunMode_Manual(t *testing.T) {
	rName := acctest.RandString(6)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { provider_test.TestAccPreCheckIntegration(t) },
		Providers: provider_test.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSyncConfig_runModeManual(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("census_sync.test", "label", "Test Manual Schedule"),
					resource.TestCheckResourceAttr("census_sync.test", "run_mode.0.type", "triggered"),
					resource.TestCheckResourceAttr("census_sync.test", "run_mode.0.triggers.0.schedule.0.frequency", "never"),
				),
			},
		},
	})
}

// TestAccResourceSync_Alerts tests alert configurations
func TestAccResourceSync_Alerts(t *testing.T) {
	rName := acctest.RandString(6)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { provider_test.TestAccPreCheckIntegration(t) },
		Providers: provider_test.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSyncConfig_alerts(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("census_sync.test", "label", "Test Alerts"),
					resource.TestCheckResourceAttr("census_sync.test", "alert.#", "2"),
					resource.TestCheckResourceAttr("census_sync.test", "alert.0.type", "FailureAlertConfiguration"),
					resource.TestCheckResourceAttr("census_sync.test", "alert.1.type", "InvalidRecordPercentAlertConfiguration"),
				),
			},
		},
	})
}

// =============================================================================
// Configuration Helper Functions
// =============================================================================

// testAccIntegrationBaseConfig returns base configuration with workspace, source, and destination
func testAccIntegrationBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "census_workspace" "test" {
  name = "Test Workspace - Sync Integration"
}

resource "census_source" "test" {
  workspace_id = census_workspace.test.id
  name = "Test_Redshift_Source_%s"
  type = "redshift"

  connection_config = {
    hostname = "%s"
    port     = "%s"
    database = "%s"
    user     = "%s"
    password = "%s"
  }
}

resource "census_destination" "test" {
  workspace_id = census_workspace.test.id
  name = "Test Salesforce Destination"
  type = "salesforce"

  connection_config = {
    username        = "%s"
    instance_url    = "%s"
    client_id       = "%s"
    jwt_signing_key = "%s"
    domain          = "%s"
  }
}
`,
		rName,
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

func testAccResourceSyncConfig_basic(rName string) string {
	return testAccIntegrationBaseConfig(rName) + `
resource "census_sync" "test" {
  workspace_id = census_workspace.test.id
  label        = "Test Basic Sync"

  source_attributes {
    connection_id = census_source.test.id
    object {
      type          = "table"
      table_name    = "users"
      table_schema  = "public"
      table_catalog = "dev"
    }
  }

  destination_attributes {
    connection_id = census_destination.test.id
    object        = "Contact"
  }

  operation = "upsert"

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
    from = "id"
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
`
}

func testAccResourceSyncConfig_updated(rName string) string {
	return testAccIntegrationBaseConfig(rName) + `
resource "census_sync" "test" {
  workspace_id = census_workspace.test.id
  label        = "Updated Sync Label"

  source_attributes {
    connection_id = census_source.test.id
    object {
      type          = "table"
      table_name    = "users"
      table_schema  = "public"
      table_catalog = "dev"
    }
  }

  destination_attributes {
    connection_id = census_destination.test.id
    object        = "Contact"
  }

  operation = "upsert"

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
    from = "id"
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

  paused = false
}
`
}

func testAccResourceSyncConfig_fieldMappings(rName string) string {
	return testAccIntegrationBaseConfig(rName) + `
resource "census_sync" "test" {
  workspace_id = census_workspace.test.id
  label        = "Test Field Mappings"

  source_attributes {
    connection_id = census_source.test.id
    object {
      type          = "table"
      table_name    = "users"
      table_schema  = "public"
      table_catalog = "dev"
    }
  }

  destination_attributes {
    connection_id = census_destination.test.id
    object        = "Contact"
  }

  operation = "upsert"

  # Direct mapping with primary identifier
  field_mapping {
    from                  = "email"
    to                    = "Email"
    is_primary_identifier = true
  }

  # Regular direct mappings
  field_mapping {
    from = "first_name"
    to   = "FirstName"
  }

  field_mapping {
    from = "last_name"
    to   = "LastName"
  }

  # Constant mapping
  field_mapping {
    type     = "constant"
    constant = "Terraform Test"
    to       = "LeadSource"
  }

  field_mapping {
    from = "id"
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
`
}

func testAccResourceSyncConfig_runModeDaily(rName string) string {
	return testAccIntegrationBaseConfig(rName) + `
resource "census_sync" "test" {
  workspace_id = census_workspace.test.id
  label        = "Test Daily Schedule"

  source_attributes {
    connection_id = census_source.test.id
    object {
      type          = "table"
      table_name    = "users"
      table_schema  = "public"
      table_catalog = "dev"
    }
  }

  destination_attributes {
    connection_id = census_destination.test.id
    object        = "Contact"
  }

  operation = "upsert"

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
    from = "id"
    to   = "Census_ID__c"
  }

  run_mode {
    type = "triggered"
    triggers {
      schedule {
        frequency = "daily"
        hour      = 6
        minute    = 0
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
`
}

func testAccResourceSyncConfig_runModeHourly(rName string) string {
	return testAccIntegrationBaseConfig(rName) + `
resource "census_sync" "test" {
  workspace_id = census_workspace.test.id
  label        = "Test Hourly Schedule"

  source_attributes {
    connection_id = census_source.test.id
    object {
      type          = "table"
      table_name    = "users"
      table_schema  = "public"
      table_catalog = "dev"
    }
  }

  destination_attributes {
    connection_id = census_destination.test.id
    object        = "Contact"
  }

  operation = "upsert"

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
    from = "id"
    to   = "Census_ID__c"
  }

  run_mode {
    type = "triggered"
    triggers {
      schedule {
        frequency = "hourly"
        minute    = 30
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
`
}

func testAccResourceSyncConfig_runModeManual(rName string) string {
	return testAccIntegrationBaseConfig(rName) + `
resource "census_sync" "test" {
  workspace_id = census_workspace.test.id
  label        = "Test Manual Schedule"

  source_attributes {
    connection_id = census_source.test.id
    object {
      type          = "table"
      table_name    = "users"
      table_schema  = "public"
      table_catalog = "dev"
    }
  }

  destination_attributes {
    connection_id = census_destination.test.id
    object        = "Contact"
  }

  operation = "upsert"

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
    from = "id"
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
`
}

func testAccResourceSyncConfig_alerts(rName string) string {
	return testAccIntegrationBaseConfig(rName) + `
resource "census_sync" "test" {
  workspace_id = census_workspace.test.id
  label        = "Test Alerts"

  source_attributes {
    connection_id = census_source.test.id
    object {
      type          = "table"
      table_name    = "users"
      table_schema  = "public"
      table_catalog = "dev"
    }
  }

  destination_attributes {
    connection_id = census_destination.test.id
    object        = "Contact"
  }

  operation = "upsert"

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
    from = "id"
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
    options              = {}
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
`
}

func TestAccResourceSync_Import(t *testing.T) {
	rName := acctest.RandString(6)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { provider_test.TestAccPreCheckIntegration(t) },
		Providers: provider_test.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSyncConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("census_sync.test", "label", "Test Basic Sync"),
					resource.TestCheckResourceAttr("census_sync.test", "operation", "upsert"),
				),
			},
			{
				ResourceName:      "census_sync.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccSyncImportStateIdFunc("census_sync.test"),
			},
		},
	})
}

// Helper to construct composite ID for import (workspace_id:sync_id)
func testAccSyncImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
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
