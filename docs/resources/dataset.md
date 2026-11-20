# census_dataset Resource

Manages a SQL dataset for data transformation. Datasets allow you to write SQL queries against your source data to create derived datasets that can be used in syncs.

## Example Usage

### Basic SQL Dataset

```hcl
resource "census_dataset" "active_users" {
  name        = "Active Users"
  group_id    = census_source.warehouse.group_id

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
  name        = "High Value Customers"
  description = "Customers with lifetime value greater than $1000"
  group_id    = census_source.warehouse.group_id

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
  name        = "User Engagement Scores"
  description = "Calculated engagement metrics for active users"
  group_id    = census_source.warehouse.group_id

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

* `name` - (Required) The name of the dataset.
* `group_id` - (Required, Forces new resource) The group ID of the source connection to run the query against, usually `census_source.<name>.group_id`.
* `query` - (Required) The SQL query that defines the dataset. Use heredoc syntax for multi-line queries.
* `type` - (Optional, Forces new resource) The type of dataset. Defaults to `"sql"`. Currently only SQL datasets are supported.
* `description` - (Optional) A description of the dataset's purpose.
* `wait_for_metadata_refresh` - (Optional) Wait for Fivetran to fetch the dataset's columns before finishing the apply. Set this when a sync is created from the dataset in the same apply and the query takes a long time (over 60 seconds) to return column metadata. Changing it from `false` to `true` on an existing dataset refreshes the columns, unless they are already available. Defaults to `false`.

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
* `metadata_ready` - Whether Fivetran has fetched the dataset's columns.

## Import

Datasets can be imported using the dataset ID:

```shell
terraform import census_dataset.active_users 67890
```

## Notes

* Use heredoc syntax (`<<-SQL ... SQL`) for multi-line queries to maintain readability.
* The dataset query is validated by Fivetran when the dataset is created.
* Column information and record counts are computed by Fivetran after the query executes.
* Changes to the `query` will trigger an update to the dataset.
* Datasets can be used as sources in `census_sync` resources.
