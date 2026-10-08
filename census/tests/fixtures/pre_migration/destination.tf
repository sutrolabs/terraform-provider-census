resource "census_destination" "fixture" {
  workspace_id = census_workspace.fixture.id

  name = "pre-migration-fixture-salesforce"
  type = "salesforce"

  connection_config = {
    username        = var.salesforce_username
    client_id       = var.salesforce_client_id
    instance_url    = var.salesforce_instance_url
    jwt_signing_key = var.salesforce_jwt_signing_key
    domain          = var.salesforce_domain
  }

  auto_refresh_objects = false
}
