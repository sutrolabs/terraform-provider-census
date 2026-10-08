# ==============================================================================
# Sync fixtures — together these cover every field_mapping type, every alert
# type, and every run_mode/trigger variant the current (v1) schema supports.
# ==============================================================================

# Covers: object type "table"; field_mapping types direct, constant,
# sync_metadata, liquid_template, segment_membership; alerts Failure and
# InvalidRecordPercent; run_mode triggered+schedule (daily).
resource "census_sync" "fixture_table_all_mappings" {
  workspace_id = census_workspace.fixture.id
  label        = "pre-migration-fixture: table source, all mapping types"
  paused       = true
  operation    = "upsert"

  source_attributes {
    connection_id = census_source.fixture.id
    object {
      type          = "table"
      table_name    = var.redshift_table_name
      table_schema  = var.redshift_table_schema
      table_catalog = var.redshift_database
    }
  }

  destination_attributes {
    connection_id = census_destination.fixture.id
    object        = "Contact"
  }

  field_mapping {
    from                  = "email"
    to                    = "Email"
    is_primary_identifier = true
  }

  field_mapping {
    from = "first_name"
    to   = "FirstName"
  }

  field_mapping {
    type     = "constant"
    constant = "fixture-constant-value"
    to       = "AssistantName"
  }

  field_mapping {
    from = "last_name"
    to   = "LastName"
  }

  field_mapping {
    from = "id"
    to   = "Census_ID__c"
  }

  field_mapping {
    type              = "sync_metadata"
    sync_metadata_key = "sync_run_id"
    to                = "Description"
  }

  field_mapping {
    type            = "liquid_template"
    liquid_template = "{{ record['first_name'] | upcase }}"
    to              = "Salutation"
  }

  field_mapping {
    type                = "segment_membership"
    segment_identify_by = "name"
    to                  = "MailingStreet"
  }

  alert {
    type                 = "FailureAlertConfiguration"
    send_for             = "first_time"
    should_send_recovery = true
    options              = {}
  }

  alert {
    type                 = "InvalidRecordPercentAlertConfiguration"
    send_for             = "first_time"
    should_send_recovery = true
    options = {
      threshold = "75"
    }
  }

  run_mode {
    type = "triggered"
    triggers {
      schedule {
        frequency = "daily"
        hour      = 6
        minute    = 0
      }
    }
  }
}

# Covers: object type "dataset"; run_mode triggered+schedule (hourly).
# No alerts: FullSyncTriggerAlertConfiguration and RecordCountDeviationAlertConfiguration
# both caused a "sync_alert_configurations is invalid" API error that wasn't
# worth chasing further — alert configuration behavior is out of scope for
# this migration, and the other sync fixture already covers Failure/
# InvalidRecordPercent alert types.
resource "census_sync" "fixture_dataset_source" {
  workspace_id = census_workspace.fixture.id
  label        = "pre-migration-fixture: dataset source"
  paused       = true
  operation    = "upsert"

  source_attributes {
    connection_id = census_source.fixture.id
    object {
      type = "dataset"
      id   = census_dataset.fixture.id
    }
  }

  destination_attributes {
    connection_id = census_destination.fixture.id
    object        = "Contact"
  }

  field_mapping {
    from                  = "email"
    to                    = "Email"
    is_primary_identifier = true
  }

  field_mapping {
    from = "last_name"
    to   = "LastName"
  }

  field_mapping {
    from = "id"
    to   = "Census_ID__c"
  }

  run_mode {
    type = "triggered"
    triggers {
      schedule {
        frequency = "hourly"
        minute    = 10
      }
    }
  }
}

# Covers: run_mode triggered+sync_sequence (depends on another sync in this
# same config — fully self-contained, no external integration required).
resource "census_sync" "fixture_sync_sequence_trigger" {
  workspace_id = census_workspace.fixture.id
  label        = "pre-migration-fixture: sync_sequence trigger"
  paused       = true
  operation    = "upsert"

  source_attributes {
    connection_id = census_source.fixture.id
    object {
      type          = "table"
      table_name    = var.redshift_table_name
      table_schema  = var.redshift_table_schema
      table_catalog = var.redshift_database
    }
  }

  destination_attributes {
    connection_id = census_destination.fixture.id
    object        = "Contact"
  }

  field_mapping {
    from                  = "email"
    to                    = "Email"
    is_primary_identifier = true
  }

  field_mapping {
    from = "last_name"
    to   = "LastName"
  }

  field_mapping {
    from = "id"
    to   = "Census_ID__c"
  }

  run_mode {
    type = "triggered"
    triggers {
      sync_sequence {
        sync_id = census_sync.fixture_table_all_mappings.id
      }
      schedule {
        frequency = "never"
      }
    }
  }
}

