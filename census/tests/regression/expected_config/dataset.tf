resource "census_dataset" "fixture" {
  name        = "pre-migration-fixture-dataset"
  type        = "sql"
  description = "Pre-migration fixture dataset, used as a sync source below."
  group_id    = "pre_migration_fixture_group"

  query = "select id, email, first_name, last_name from public.users"
}
