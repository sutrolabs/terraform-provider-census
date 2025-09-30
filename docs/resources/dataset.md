# census_dataset Resource

Manages a Census SQL dataset for data transformation. Datasets allow you to write SQL queries against your source data to create derived datasets that can be used in syncs.

## Example Usage

### Basic SQL Dataset

```hcl
resource "census_dataset" "active_users" {
  workspace_id = census_workspace.main.id
  name         = "Active Users"
  source_id    = census_source.warehouse.id

  query = <<-SQL
    SELECT
      id,
      email,
      first_name,
      last_name,
      created_at
    FROM users
    WHERE active = true
      AND last_login_at > CURRENT_DATE - INTERVAL '30 days'
  SQL
}
```

### Dataset with Description

```hcl
resource "census_dataset" "high_value_customers" {
  workspace_id = census_workspace.main.id
  name         = "High Value Customers"
  description  = "Customers with lifetime value greater than $1000"
  source_id    = census_source.warehouse.id

  query = <<-SQL
    SELECT
      u.id,
      u.email,
      u.first_name,
      u.last_name,
      SUM(o.amount) as lifetime_value,
      COUNT(o.id) as order_count
    FROM users u
    JOIN orders o ON u.id = o.user_id
    GROUP BY u.id, u.email, u.first_name, u.last_name
    HAVING SUM(o.amount) > 1000
  SQL
}
```

### Complex Transformation

```hcl
resource "census_dataset" "user_engagement_score" {
  workspace_id = census_workspace.main.id
  name         = "User Engagement Scores"
  description  = "Calculated engagement metrics for active users"
  source_id    = census_source.warehouse.id

  query = <<-SQL
    WITH user_activity AS (
      SELECT
        user_id,
        COUNT(DISTINCT DATE(event_timestamp)) as days_active,
        COUNT(*) as total_events
      FROM events
      WHERE event_timestamp > CURRENT_DATE - INTERVAL '90 days'
      GROUP BY user_id
    )
    SELECT
      u.id,
      u.email,
      u.name,
      ua.days_active,
      ua.total_events,
      ROUND((ua.days_active::FLOAT / 90) * 100, 2) as engagement_score
    FROM users u
    JOIN user_activity ua ON u.id = ua.user_id
    WHERE ua.days_active >= 5
  SQL
}
```

## Argument Reference

* `workspace_id` - (Required, Forces new resource) The ID of the workspace this dataset belongs to.
* `name` - (Required) The name of the dataset.
* `source_id` - (Required, Forces new resource) The ID of the source connection to run the query against.
* `query` - (Required) The SQL query that defines the dataset. Use heredoc syntax for multi-line queries.
* `type` - (Optional, Forces new resource) The type of dataset. Defaults to `"sql"`. Currently only SQL datasets are supported.
* `description` - (Optional) A description of the dataset's purpose.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the dataset.
* `resource_identifier` - A unique resource identifier for the dataset.
* `cached_record_count` - The cached count of records in the dataset (computed after query execution).
* `columns` - A list of columns in the dataset with their data types:
  * `name` - The column name.
  * `data_type` - The column's data type.
* `created_at` - Timestamp when the dataset was created.
* `updated_at` - Timestamp when the dataset was last updated.

## Import

Datasets can be imported using the workspace ID and dataset ID separated by a colon:

```shell
terraform import census_dataset.active_users "workspace_id:dataset_id"
```

For example:

```shell
terraform import census_dataset.active_users "12345:67890"
```

## Notes

* Use heredoc syntax (`<<-SQL ... SQL`) for multi-line queries to maintain readability.
* The dataset query is validated by Census when the dataset is created.
* Column information and record counts are computed by Census after the query executes.
* Changes to the `query` will trigger an update to the dataset.
* Datasets can be used as sources in `census_sync` resources.