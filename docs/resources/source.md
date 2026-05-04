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

### BigQuery Source

```hcl
resource "census_source" "bigquery" {
  workspace_id = census_workspace.main.id
  name         = "Analytics BigQuery"
  type         = "big_query"

  connection_config = {
    project_id = "my-gcp-project"
    dataset_id = "analytics"
    private_key = var.gcp_service_account_key
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
* `connection_config` - (Required, Sensitive) Map of credentials for connecting to the source. Supports strings, numbers, and booleans. The required fields vary by source type and are validated against the Census API schema.

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
* Source types and required credential fields are validated against the Census API's `/source_types` endpoint.
* After creation, the provider automatically triggers a table refresh to discover available tables.
* `sync_engine` is a create-time setting. The Census API rejects in-place sync engine changes with a 4xx response, so Terraform recreates the source when this field changes.
* Leaving `sync_engine` unset creates new sources with the advanced engine by default, but preserves the current engine for existing managed sources to avoid unexpected replacements during provider upgrades.
* If you already use `sync_engine`, set it explicitly in configuration so future provider upgrades cannot infer a different value from a changing default.
* `warehouse_writeback_retention_in_days` can be increased or decreased on an existing managed source by changing the value in configuration; the provider issues a PATCH on the next apply. Leaving the attribute unset preserves whatever value the Census API reports, including for sources where Warehouse Writeback was enabled outside Terraform. The Census API does not currently expose a way to disable Warehouse Writeback once enabled, so removing this attribute from configuration stops Terraform from managing the value but does not turn the feature off in Census. Use the Census UI if you need to disable it.
