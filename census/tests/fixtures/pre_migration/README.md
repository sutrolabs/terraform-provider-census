# Pre-migration state fixture

Produces a regression baseline: one real resource of every type (workspace,
source, destination, sync, dataset), with the sync resources collectively
covering every `field_mapping` type, every `alert` type, and every
`run_mode`/trigger variant the current (v1-backed) provider schema supports.

Run this **before any provider code change lands** — it has to reflect
genuinely untouched behavior.

## Setup

```bash
cp terraform.tfvars.example terraform.tfvars
# fill in terraform.tfvars with real staging credentials

cd ../../../..  # repo root
make build      # or: make install, if you want a local dev-override build

cd census/tests/fixtures/pre_migration
terraform init
terraform plan
terraform apply
```

If `fixture_dbt_cloud_trigger` or `fixture_fivetran_trigger` fail because you
don't have those integrations configured in your staging org, leave their
variables unset — `count` makes both resources skip creation entirely rather
than fail. Same for `fixture_live_mode` if your source type doesn't support
live/continuous mode.

## ⚠️ Before committing the resulting state anywhere

Terraform state stores `Sensitive`-marked attribute values — `connection_config`
on the source and destination, and the workspace's `api_key` — **in
plaintext**. The `Sensitive` flag only affects CLI output, not what's written
to the state file. Do not commit the raw state file.

Before copying it to `census/tests/fixtures/pre_migration.tfstate`, redact
those fields, e.g.:

```bash
jq '
  .resources |= map(
    if .type == "census_destination" or .type == "census_source" then
      .instances |= map(.attributes.connection_config = {})
    elif .type == "census_workspace" then
      .instances |= map(.attributes.api_key = "")
    else
      .
    end
  )
' terraform.tfstate > ../pre_migration.tfstate
```

This has to be scoped by resource `type` rather than applied blanket across
every resource — a blanket `.resources[].instances[].attributes.connection_config`
assignment creates that key (as `null`) on resource types that don't actually
have it (e.g. `census_dataset`, `census_sync`, `census_workspace`), since a
jq path assignment creates missing keys rather than skipping them. An earlier
version of this command had exactly that bug; if you redacted a state file
with it before this note was added, re-redact from the original rather than
trusting the result.

Then diff the result against the original and confirm no credential values
remain anywhere else (e.g. inside a `query` string, if you ever add one that
embeds a literal secret — this fixture's dataset query doesn't).

## Teardown

The regression harness (`census/tests/provider/regression/`) runs `terraform
plan -refresh=false` against the frozen, redacted state file — it never
contacts the real API. Once you've captured and redacted the state, the real
staging resources created here don't need to stay alive.

**Copy/redact the state before you destroy** — `terraform destroy` empties
your local state file as part of tearing down, so do the copy step above
first, then:

```bash
terraform destroy
```
