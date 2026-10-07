# Fixtures

## `pre_migration.tfstate`

A real, redacted Terraform state snapshot (`connection_config` and `api_key`
values blanked) produced by applying `pre_migration/` against the current
(v1-backed) provider. Used by `census/tests/provider/regression/` as a
before-and-after baseline: a provider build should be able to `plan` against
this state with no unexpected changes.

**Current coverage is partial.** Only one `census_sync` resource
(`fixture_table_all_mappings`) actually applied successfully — it covers all
five `field_mapping` types (direct, constant, sync_metadata, liquid_template,
segment_membership) and two alert types (Failure, InvalidRecordPercent), with
a `triggered`+`schedule` run_mode.

Not yet represented, because the corresponding resources in `pre_migration/`
failed to apply due to a Census-side bug currently being worked with another
team:
- `fixture_dataset_source` — dataset-type source object; FullSyncTrigger and
  RecordCountDeviation alerts
- `fixture_live_mode` — `run_mode.type = "live"`; Runtime and Status alerts
- `fixture_sync_sequence_trigger` — `sync_sequence` trigger

Also not yet created (optional, `count`-gated on unset variables, not a
failure): `fixture_dbt_cloud_trigger`, `fixture_fivetran_trigger`.

**To refresh this fixture** once the upstream bug is fixed: re-run
`terraform apply` in `pre_migration/`, redact the result the same way
described in `pre_migration/README.md`, and overwrite this file.
