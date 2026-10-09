# ==============================================================================
# PRE-MIGRATION STATE FIXTURE
# ==============================================================================
# Creates one real resource of every type this provider manages, covering
# every field_mapping type, every alert type, and every run_mode/trigger
# variant, against the CURRENT (v1-backed) provider. The resulting state file
# becomes the regression baseline every migration stage is checked against.
#
# Do not add anything here related to the v2 migration itself — this config
# must reflect pre-migration (v1) behavior only.

terraform {
  required_providers {
    census = {
      source  = "sutrolabs/census"
      version = "~> 0.2.0"
    }
    # To test against an uncommitted local build instead of the published
    # registry version, run `make install` from the repo root and add a
    # dev_overrides entry for "sutrolabs/census" in your ~/.terraformrc
    # pointing at $GOPATH/bin, then remove the `version` constraint above.
  }
}

provider "census" {
  # personal_access_token and base_url are intentionally omitted here — the
  # provider schema already falls back to the CENSUS_PERSONAL_ACCESS_TOKEN and
  # CENSUS_BASE_URL environment variables (see census/provider/provider.go)
  # when the attribute isn't set in config at all. Source .env.test before
  # running terraform and these are picked up with no mapping required.
}
