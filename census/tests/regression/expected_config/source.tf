resource "census_source" "fixture" {
  name = "pre_migration_fixture_redshift"
  type = "redshift"

  connection_config   = {}
  auto_refresh_tables = false
}
