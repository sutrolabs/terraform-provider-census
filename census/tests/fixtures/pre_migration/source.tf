resource "census_source" "fixture" {
  workspace_id = census_workspace.fixture.id

  name = "pre_migration_fixture_redshift"
  type = "redshift"

  connection_config = {
    hostname = var.redshift_host
    port     = var.redshift_port
    database = var.redshift_database
    user     = var.redshift_username
    password = var.redshift_password
  }

  auto_refresh_tables = false
}
