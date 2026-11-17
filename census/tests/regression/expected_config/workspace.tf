resource "census_workspace" "fixture" {
  name                     = "pre-migration-fixture"
  notification_emails      = []
  return_workspace_api_key = true
}
