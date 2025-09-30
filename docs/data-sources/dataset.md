# census_dataset Data Source

Retrieves information about an existing Census SQL dataset.

## Example Usage

```hcl
data "census_dataset" "active_users" {
  id           = "115721"
  workspace_id = census_workspace.main.id
}

output "dataset_columns" {
  value = data.census_dataset.active_users.columns
}

output "record_count" {
  value = data.census_dataset.active_users.cached_record_count
}
```

## Argument Reference

* `id` - (Required) The ID of the dataset.
* `workspace_id` - (Required) The ID of the workspace this dataset belongs to.

## Attribute Reference

* `name` - The name of the dataset.
* `type` - The type of dataset (typically "sql").
* `description` - The description of the dataset.
* `query` - The SQL query that defines the dataset.
* `source_id` - The ID of the source connection the query runs against.
* `resource_identifier` - A unique resource identifier for the dataset.
* `cached_record_count` - The cached count of records in the dataset.
* `columns` - A list of columns in the dataset:
  * `name` - The column name.
  * `data_type` - The column's data type.
* `created_at` - Timestamp when the dataset was created.
* `updated_at` - Timestamp when the dataset was last updated.