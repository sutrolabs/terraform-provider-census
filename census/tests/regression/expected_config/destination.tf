resource "census_destination" "fixture" {
  workspace_id = census_workspace.fixture.id

  name = "pre-migration-fixture-salesforce"
  type = "salesforce"

  connection_config    = {}
  auto_refresh_objects = false
}
