resource "census_dataset" "fixture" {
  workspace_id = census_workspace.fixture.id
  name         = "pre-migration-fixture-dataset"
  type         = "sql"
  description  = "Pre-migration fixture dataset, used as a sync source below."
  source_id    = census_source.fixture.id

  query = "select id, email, first_name, last_name from public.users"
}
