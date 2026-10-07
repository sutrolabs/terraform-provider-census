# ==============================================================================
# SOURCE CONNECTION (Redshift) — matches the TESTING.md staging setup.
# host/username/password/table_name have no default on purpose: if .env.test
# isn't sourced (see README), terraform should fail asking for them rather
# than silently applying against a placeholder.
# ==============================================================================

variable "redshift_host" {
  description = "Redshift test cluster hostname. Reused from .env.test's CENSUS_TEST_REDSHIFT_HOST."
  type        = string
}

variable "redshift_port" {
  description = "Redshift test cluster port. Reused from .env.test's CENSUS_TEST_REDSHIFT_PORT."
  type        = string
  default     = "5439"
}

variable "redshift_database" {
  description = "Redshift test database name. Reused from .env.test's CENSUS_TEST_REDSHIFT_DATABASE."
  type        = string
  default     = "dev"
}

variable "redshift_username" {
  description = "Redshift test user. Reused from .env.test's CENSUS_TEST_REDSHIFT_USERNAME."
  type        = string
  sensitive   = true
}

variable "redshift_password" {
  description = "Redshift test user password. Reused from .env.test's CENSUS_TEST_REDSHIFT_PASSWORD."
  type        = string
  sensitive   = true
}

variable "redshift_table_schema" {
  description = "Schema containing the test table used by the sync resources. Not present in .env.test — set directly."
  type        = string
  default     = "public"
}

variable "redshift_table_name" {
  description = "Table name used as the sync source object. Doesn't need to actually exist — census_sync doesn't validate table existence at create time, and these syncs are all paused. Matches the literal value the existing acceptance tests (resource_sync_integration_test.go) already use, unverified, with no setup step that creates it."
  type        = string
  default     = "users"
}

# ==============================================================================
# DESTINATION CONNECTION (Salesforce, JWT OAuth) — matches the TESTING.md
# staging setup. No defaults, same reasoning as above.
# ==============================================================================

variable "salesforce_username" {
  description = "Salesforce sandbox username. Reused from .env.test's CENSUS_TEST_SALESFORCE_USERNAME."
  type        = string
  sensitive   = true
}

variable "salesforce_client_id" {
  description = "Salesforce Connected App consumer key. Reused from .env.test's CENSUS_TEST_SALESFORCE_CLIENT_ID."
  type        = string
  sensitive   = true
}

variable "salesforce_instance_url" {
  description = "Salesforce sandbox instance URL. Reused from .env.test's CENSUS_TEST_SALESFORCE_INSTANCE_URL."
  type        = string
}

variable "salesforce_jwt_signing_key" {
  description = "RSA private key used for Salesforce JWT OAuth (PEM, with literal \\n for newlines). Reused from .env.test's CENSUS_TEST_SALESFORCE_JWT_SIGNING_KEY."
  type        = string
  sensitive   = true
}

variable "salesforce_domain" {
  description = "Salesforce OAuth domain. Reused from .env.test's CENSUS_TEST_SALESFORCE_DOMAIN, passed straight through unchanged — matches how the existing acceptance tests use it (see resource_destination_test.go)."
  type        = string
}

# ==============================================================================
# OPTIONAL: dbt Cloud trigger — only used if both values are provided
# ==============================================================================

variable "dbt_cloud_project_id" {
  description = "dbt Cloud project ID for the dbt_cloud trigger sync. Leave null to skip that resource entirely (it won't be created)."
  type        = string
  default     = null
}

variable "dbt_cloud_job_id" {
  description = "dbt Cloud job ID for the dbt_cloud trigger sync. Leave null to skip that resource entirely."
  type        = string
  default     = null
}

# ==============================================================================
# OPTIONAL: Fivetran trigger — only used if both values are provided
# ==============================================================================

variable "fivetran_job_id" {
  description = "Fivetran job ID for the fivetran trigger sync. Leave null to skip that resource entirely."
  type        = string
  default     = null
}

variable "fivetran_job_name" {
  description = "Fivetran job name for the fivetran trigger sync. Leave null to skip that resource entirely."
  type        = string
  default     = null
}
