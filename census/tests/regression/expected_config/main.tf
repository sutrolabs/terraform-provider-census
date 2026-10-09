terraform {
  required_providers {
    census = {
      source = "sutrolabs/census"
    }
  }
}

provider "census" {
  personal_access_token = "unused-refresh-is-disabled"
}
