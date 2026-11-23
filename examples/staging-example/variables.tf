variable "census_personal_token" {
  description = "Census Personal Access Token for authentication"
  type        = string
  sensitive   = true
}

variable "census_base_url" {
  description = "Activations API base URL (use staging URL for testing)"
  type        = string
  default     = "https://app.staging.getcensus.com"
}

variable "notification_emails" {
  description = "Email addresses that receive sync alerts for the source"
  type        = list(string)
  default     = []
}

# Source Configuration
variable "source_label" {
  description = "Label for the data source"
  type        = string
  default     = "Staging Data Warehouse"
}

variable "source_type" {
  description = "Type of data source (e.g., postgres, snowflake, bigquery)"
  type        = string
}

variable "source_credentials" {
  description = "Credentials for the data source (supports booleans, numbers, and strings)"
  type        = map(any)
  sensitive   = true
}

variable "source_table" {
  description = "Source table name for the sync"
  type        = string
}

# Destination Configuration
variable "destination_label" {
  description = "Name for the activation"
  type        = string
  default     = "Staging CRM"
}

variable "destination_type" {
  description = "Destination type for the activation (e.g., salesforce, hubspot)"
  type        = string
}

variable "destination_credentials" {
  description = "Credentials for the activation's destination connection"
  type        = map(any)
  sensitive   = true
}

variable "destination_object" {
  description = "Destination object (e.g., Salesforce object name)"
  type        = string
}

# Sync Configuration
variable "sync_label" {
  description = "Label for the sync"
  type        = string
  default     = "Staging Test Sync"
}

variable "field_mapping" {
  description = "Field mappings for the sync"
  type = list(object({
    from                  = string
    to                    = string
    is_primary_identifier = optional(bool, false)
  }))
}

variable "sync_frequency" {
  description = "Sync frequency: never, quarter_hourly, hourly, daily or weekly"
  type        = string
  default     = "daily"
}

variable "sync_hour" {
  description = "Hour of day to run sync (0-23)"
  type        = number
  default     = 9
}

variable "sync_minute" {
  description = "Minute of hour to run sync (0-59)"
  type        = number
  default     = 0
}

variable "sync_paused" {
  description = "Whether the sync is paused"
  type        = bool
  default     = true
}