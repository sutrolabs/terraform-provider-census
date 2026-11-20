# Fixtures

## `pre_migration.tfstate`

A real, redacted Terraform state snapshot (`connection_config` and `api_key`
values blanked) produced by applying `pre_migration/` against the current
(v1-backed) provider. Used by `census/tests/regression/` as a
before-and-after baseline: a provider build should be able to `plan` against
this state with no unexpected changes.

**All three `census_sync` resources in `pre_migration/` now apply successfully:**
- `fixture_table_all_mappings` — all five `field_mapping` types (direct,
  constant, sync_metadata, liquid_template, segment_membership) and two alert
  types (Failure, InvalidRecordPercent), with a `triggered`+`schedule`
  run_mode.
- `fixture_dataset_source` — dataset-type source object, `triggered`+`schedule`
  run_mode. No alerts: it originally also covered the FullSyncTrigger and
  RecordCountDeviation alert types, but both caused a
  `sync_alert_configurations is invalid` API error that wasn't worth chasing —
  alert configuration behavior is out of scope for this migration, and
  `fixture_table_all_mappings` already covers two other alert types. Those two
  alert blocks were removed from that sync's config permanently, not deferred.
- `fixture_sync_sequence_trigger` — `sync_sequence` trigger.

`fixture_dbt_cloud_trigger` and `fixture_fivetran_trigger` have been removed
from `pre_migration/` entirely (not just skipped) — dbt Cloud/Fivetran sync
triggers aren't changing as part of this migration, so there's no need to
cover them in this fixture.

There is no `run_mode.type = "live"` fixture — live mode syncs have been
completely sunset from the Census product since this provider was authored,
so there's nothing left to exercise. This also means Runtime and Status alert
types (which the removed live-mode fixture covered) aren't represented in
this fixture at all anymore.

**If this file doesn't yet reflect the state above**, it needs refreshing:
re-run `terraform apply` in `pre_migration/`, redact the result the same way
described in `pre_migration/README.md`, and overwrite this file.
