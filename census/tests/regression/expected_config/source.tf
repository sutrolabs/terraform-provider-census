resource "census_source" "fixture" {
  workspace_id = census_workspace.fixture.id

  name = "pre_migration_fixture_redshift"
  type = "redshift"

  connection_config   = {}
  auto_refresh_tables = false
}
