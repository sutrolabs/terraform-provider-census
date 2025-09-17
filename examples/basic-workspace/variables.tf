variable "census_personal_token" {
  description = "Census personal access token for all operations"
  type        = string
  sensitive   = true
}

variable "census_region" {
  description = "Census region (us or eu)"
  type        = string
  default     = "us"
  
  validation {
    condition     = contains(["us", "eu"], var.census_region)
    error_message = "Region must be either 'us' or 'eu'."
  }
}