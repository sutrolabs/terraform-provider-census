# census_sync Data Source

Retrieves information about an existing Fivetran Activations sync.

## Example Usage

```hcl
data "census_sync" "user_sync" {
  id = "98765"
}

output "sync_activation" {
  value = data.census_sync.user_sync.activation_id
}

output "sync_paused" {
  value = data.census_sync.user_sync.paused
}

output "sync_source_table" {
  value = jsondecode(data.census_sync.user_sync.source_object).table_name
}
```

## Argument Reference

* `id` - (Required) The ID of the sync.

## Attribute Reference

* `activation_id` - The ID of the activation this sync belongs to.
* `label` - The name/label of the sync.
* `source_object` - The sync's source object as a JSON string, e.g. `{"type":"table","table_catalog":"prod","table_schema":"public","table_name":"users"}`.
* `destination_object` - The destination object name.
* `paused` - Whether the sync is paused.
* `run_mode` - Run mode configuration:
  * `type` - Mode type: `triggered`.
  * `triggers` - Trigger configurations:
    * `schedule` - Schedule-based trigger:
      * `frequency` - How often to run: `never`, `continuous`, `quarter_hourly`, `hourly`, `daily`, `weekly`, or `expression`.
      * `day` - Day of week (for weekly schedules).
      * `hour` - Hour to run (0-24).
      * `minute` - Minute to run (0-59).
      * `cron_expression` - Cron expression (when `frequency` is `expression`).
    * `dbt_cloud` - dbt Cloud job trigger: `project_id`, `job_id`.
    * `fivetran` - Fivetran connector trigger: `job_id`, `job_name`.
    * `sync_sequence` - Runs after another sync completes: `sync_id`.
* `status` - The current status of the sync.
* `created_at` - When the sync was created.
* `updated_at` - When the sync was last updated.
