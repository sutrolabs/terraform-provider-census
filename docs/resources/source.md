# census_source Resource

Manages a Census data source connection. Sources connect to data warehouses like Snowflake, BigQuery, Postgres, and others.

## Example Usage

### Snowflake Source (Password Authentication)

```hcl
resource "census_source" "warehouse" {
  workspace_id = census_workspace.main.id
  name         = "Production Warehouse"
  type         = "snowflake"

  connection_config = {
    account        = "abc12345.us-east-1"
    warehouse      = "COMPUTE_WH"
    database       = "PRODUCTION"
    username       = "census_user"
    password       = var.snowflake_password
    role           = "CENSUS_ROLE"
  }
}
```

### Snowflake Source (Explicit Basic Sync Engine)

```hcl
resource "census_source" "warehouse_basic" {
  workspace_id = census_workspace.main.id
  name         = "Warehouse Source (Basic Engine)"
  type         = "snowflake"
  sync_engine  = "basic"

  connection_config = {
    account   = "abc12345.us-east-1"
    warehouse = "COMPUTE_WH"
    database  = "PRODUCTION"
    username  = "census_user"
    password  = var.snowflake_password
    role      = "CENSUS_ROLE"
  }
}
```

### Snowflake Source (Warehouse Writeback Enabled)

```hcl
resource "census_source" "warehouse_with_writeback" {
  workspace_id = census_workspace.main.id
  name         = "Production Warehouse"
  type         = "snowflake"
  sync_engine  = "advanced"

  warehouse_writeback_retention_in_days = 30

  connection_config = {
    account   = "abc12345.us-east-1"
    warehouse = "COMPUTE_WH"
    database  = "PRODUCTION"
    username  = "census_user"
    password  = var.snowflake_password
    role      = "CENSUS_ROLE"
  }
}
```

### Snowflake Source (Keypair Authentication)

```hcl
resource "census_source" "warehouse_keypair" {
  workspace_id = census_workspace.main.id
  name         = "Production Warehouse (Keypair)"
  type         = "snowflake"

  connection_config = {
    account                = "abc12345.us-east-1"
    warehouse              = "COMPUTE_WH"
    database               = "PRODUCTION"
    username               = "census_user"
    role                   = "CENSUS_ROLE"
    use_keypair            = true  # Boolean works directly!
    private_key_pkcs8      = var.snowflake_private_key
    private_key_passphrase = var.snowflake_key_passphrase  # Optional, omit if key is not encrypted
  }
}
```

### Write-Only Secrets (any source type)

Supply secret credentials through the write-only `connection_config_wo` argument
so the secret payloads are never written to Terraform state or plan files. It
takes a `jsonencode`-d object of credential key/value pairs, and each key is merged into
`connection_config` (taking precedence over the same key there). This requires
Terraform 1.11 or newer. Keep only non-secret fields in `connection_config`, and
increment `connection_config_wo_version` whenever you rotate the secrets so
Terraform applies the new values.

```hcl
# Snowflake keypair example
resource "census_source" "warehouse_keypair_wo" {
  workspace_id = census_workspace.main.id
  name         = "Production Warehouse (Keypair, write-only)"
  type         = "snowflake"

  connection_config = {
    account     = "abc12345.us-east-1"
    warehouse   = "COMPUTE_WH"
    database    = "PRODUCTION"
    username    = "census_user"
    role        = "CENSUS_ROLE"
    use_keypair = true
  }

  # Secret payloads are write-only: present in config, never persisted in state.
  connection_config_wo = jsonencode({
    private_key_pkcs8      = var.snowflake_private_key
    private_key_passphrase = var.snowflake_key_passphrase # Optional
  })
  connection_config_wo_version = 1
}

# The same argument works for any connector, e.g. a password-based source:
resource "census_source" "postgres_wo" {
  workspace_id = census_workspace.main.id
  name         = "Production Database (write-only secret)"
  type         = "postgres"

  connection_config = {
    host     = "postgres.example.com"
    port     = 5432
    database = "production"
    username = "census"
  }

  connection_config_wo         = jsonencode({ password = var.postgres_password })
  connection_config_wo_version = 1
}
```

### BigQuery Source

```hcl
resource "census_source" "bigquery" {
  workspace_id = census_workspace.main.id
  name         = "Analytics BigQuery"
  type         = "big_query"

  connection_config = {
    project_id = "my-gcp-project"
    dataset_id = "analytics"
    service_account_key = var.gcp_service_account_key
  }
}
```

### Postgres Source

```hcl
resource "census_source" "postgres" {
  workspace_id = census_workspace.main.id
  name         = "Production Database"
  type         = "postgres"

  connection_config = {
    host     = "postgres.example.com"
    port     = 5432  # Numbers work directly
    database = "production"
    username = "census"
    password = var.postgres_password
  }
}
```

## Argument Reference

* `workspace_id` - (Required, Forces new resource) The ID of the workspace this source belongs to.
* `name` - (Required) The name of the source.
* `type` - (Required, Forces new resource) The type of data source connector. Supported types include:
  - `snowflake`
  - `big_query`
  - `postgres`
  - `redshift`
  - `databricks`
  - `mysql`
  - And many more... (validated against Census API)
* `sync_engine` - (Optional, Computed, Forces new resource) The sync engine to use when the source is created. New sources default to `advanced` to match the Census UI. Leave this unset to preserve the current engine on existing sources, or set it explicitly to `basic` only when you need the basic engine.
* `warehouse_writeback_retention_in_days` - (Optional, Computed) Enables Warehouse Writeback for the source and sets the sync log retention period (in days). Setting this attribute enables the feature; omit it to leave the value to whatever the Census API reports for this source. Only supported on the `advanced` sync engine and on source types that support sync logs (Snowflake, BigQuery, Databricks, Redshift). The Census API rejects requests that set this on unsupported source types or basic-engine sources.
* `connection_config` - (Required, Sensitive) Map of credentials for connecting to the source. Supports strings, numbers, and booleans. The required fields vary by source type and are validated against the Census API schema. Secret values may instead be supplied through the write-only `connection_config_wo` argument below so they are not persisted in Terraform state or plan files.
* `connection_config_wo` - (Optional, Write-only, Sensitive) A `jsonencode`-d object of secret connection credentials (e.g. `password`, Snowflake `private_key_pkcs8` / `private_key_passphrase`, BigQuery `service_account_key`, etc.). Requires Terraform 1.11+. Its value is never stored in Terraform state or plan. Each key is merged into `connection_config` on create and update, taking precedence over the same key set in `connection_config`. Must be used together with `connection_config_wo_version`.
* `connection_config_wo_version` - (Optional) Trigger for updating the write-only `connection_config_wo` argument. Because write-only values are never stored in state, Terraform cannot detect when they change. Increment this integer whenever you change `connection_config_wo` so the current write-only values are re-applied on the next apply.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the source.
* `status` - The current status of the source.
* `test_status` - The test status of the source connection.
* `sync_engine` - The sync engine configured for the source when it is returned by the Census API.

## Import

Sources can be imported using the workspace ID and source ID separated by a colon:

```shell
terraform import census_source.warehouse "workspace_id:source_id"
```

For example:

```shell
terraform import census_source.warehouse "12345:67890"
```

## Notes

* The `connection_config` field is marked as sensitive and will not be displayed in Terraform output.
* The write-only `connection_config_wo` argument requires Terraform 1.11 or newer. Its value is available to the provider during an apply but is never written to state or plan files. Because Terraform cannot diff a value it does not store, increment `connection_config_wo_version` whenever you change `connection_config_wo` so the provider re-sends it to Census on the next apply.
* `connection_config_wo` is fully optional and backwards compatible: existing configurations that keep secrets in `connection_config` continue to work unchanged. When a key is set in both `connection_config_wo` and `connection_config`, the `connection_config_wo` value takes precedence.
* Value typing in `connection_config_wo`: write each credential as its natural HCL type inside `jsonencode({ ... })` and that type is preserved (strings stay strings, numbers stay numbers, booleans stay booleans, objects and lists stay structured). Unlike `connection_config`, scalar string secrets are never re-typed — a `password` of `"123456"` is sent as the string `"123456"`, not the number `123456`. A structured credential that is supplied as a JSON *string* (for example `service_account_key = file("service-account.json")`, or a nested `jsonencode(...)`) is parsed back into a real object/array before being sent to Census, so you do not need to `jsondecode` it yourself. The only value that is re-typed is a string whose entire content is a JSON object or array; an ordinary secret string is left exactly as written.
* Recommended usage: keep non-secret settings in `connection_config`, put secret values in `connection_config_wo`, and bump `connection_config_wo_version` every time you rotate a secret.
* Source types and required credential fields are validated against the Census API's `/source_types` endpoint.
* After creation, the provider automatically triggers a table refresh to discover available tables.
* `sync_engine` is a create-time setting. The Census API rejects in-place sync engine changes with a 4xx response, so Terraform recreates the source when this field changes.
* Leaving `sync_engine` unset creates new sources with the advanced engine by default, but preserves the current engine for existing managed sources to avoid unexpected replacements during provider upgrades.
* If you already use `sync_engine`, set it explicitly in configuration so future provider upgrades cannot infer a different value from a changing default.
* `warehouse_writeback_retention_in_days` can be increased or decreased on an existing managed source by changing the value in configuration; the provider issues a PATCH on the next apply. Leaving the attribute unset preserves whatever value the Census API reports, including for sources where Warehouse Writeback was enabled outside Terraform. The Census API does not currently expose a way to disable Warehouse Writeback once enabled, so removing this attribute from configuration stops Terraform from managing the value but does not turn the feature off in Census. Use the Census UI if you need to disable it.
