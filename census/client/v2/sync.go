package v2

import (
	"context"
	"fmt"
	"time"
)

// Sync is a Census sync as returned by the v2 /syncs endpoints
// (apidocs/v2/paths/syncs.yaml, apidocs/v2/components/syncs.yaml). Unlike
// v1, a sync is identified by its activation_id rather than a
// workspace/source/destination connection triple.
type Sync struct {
	ID                      int                        `json:"id"`
	ActivationID            string                     `json:"activation_id"`
	Status                  string                     `json:"status,omitempty"`
	CreatedAt               time.Time                  `json:"created_at"`
	UpdatedAt               time.Time                  `json:"updated_at"`
	Label                   string                     `json:"label,omitempty"`
	Operation               string                     `json:"operation,omitempty"`
	SourceAttributes        map[string]interface{}     `json:"source_attributes,omitempty"`
	DestinationAttributes   *SyncDestinationAttributes `json:"destination_attributes,omitempty"`
	Mappings                []MappingAttributes        `json:"mappings,omitempty"`
	AlertAttributes         []AlertAttribute           `json:"alert_attributes,omitempty"`
	AdvancedConfiguration   map[string]interface{}     `json:"advanced_configuration,omitempty"`
	Mode                    *SyncMode                  `json:"mode,omitempty"`
	Paused                  bool                       `json:"paused,omitempty"`
	FieldBehavior           string                     `json:"field_behavior,omitempty"`
	FieldNormalization      string                     `json:"field_normalization,omitempty"`
	FieldOrder              string                     `json:"field_order,omitempty"`
	SyncBehaviorFamily      string                     `json:"sync_behavior_family,omitempty"`
	HighWaterMarkAttribute  string                     `json:"high_water_mark_attribute,omitempty"`
	HistoricalSyncOperation string                     `json:"historical_sync_operation,omitempty"`
	MirrorStrategy          string                     `json:"mirror_strategy,omitempty"`
}

// SyncDestinationAttributes configures how records are written to the
// destination object. Unlike v1, it carries no connection_id — the
// destination connection is now determined by the sync's activation_id, per
// apidocs/v2/components/syncs.yaml's DestinationAttributes schema.
type SyncDestinationAttributes struct {
	Object            string `json:"object"`
	LeadUnionInsertTo string `json:"lead_union_insert_to,omitempty"`
}

// MappingAttributes is one field_mapping entry, shared by sync create,
// update and read requests/responses (apidocs/v2/components/syncs.yaml's
// MappingAttributes).
//
// Position isn't documented in that schema (unlike v1, which does return
// it) — kept here for parity with v1 and because the sort-by-position
// handling that field supports there is expected to carry over; confirm
// against a real v2 response before relying on it.
type MappingAttributes struct {
	From                MappingFrom `json:"from"`
	To                  string      `json:"to"`
	Position            int         `json:"position,omitempty"`
	IsPrimaryIdentifier bool        `json:"is_primary_identifier"`
	Operation           string      `json:"operation,omitempty"` // for array fields: e.g. "append", "set" — new in v2
	LookupObject        string      `json:"lookup_object,omitempty"`
	LookupField         string      `json:"lookup_field,omitempty"`
	PreserveValues      bool        `json:"preserve_values,omitempty"`
	GenerateField       bool        `json:"generate_field,omitempty"`
	SyncNullValues      *bool       `json:"sync_null_values,omitempty"`
	ArrayField          bool        `json:"array_field,omitempty"`
	FieldType           string      `json:"field_type,omitempty"`
	FollowSourceType    bool        `json:"follow_source_type,omitempty"`
}

// MappingFrom is the polymorphic "from" side of a field mapping. Data's
// shape depends on Type: a string for "column"/"liquid_template", or a
// nested object for "constant_value" — see
// apidocs/v2/components/syncs.yaml's MappingAttributes.from. "sync_metadata"
// and "segment_membership" aren't listed in that schema's oneOf, but the
// migration plan's resolved assumptions treat this as a documentation gap,
// not a removal — both are kept working exactly as v1 already supports them.
type MappingFrom struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// AlertAttribute configures one alert on a sync. Unlike v1, Options is
// required by v2 for every alert type — even types with no options of their
// own (FailureAlertConfiguration, FullSyncTriggerAlertConfiguration) still
// require sending `options: {}`. A nil Options map marshals to JSON `null`,
// not `{}`; callers must set it to a non-nil (possibly empty) map before
// sending a create/update request.
type AlertAttribute struct {
	ID                 int                    `json:"id,omitempty"`
	Type               string                 `json:"type"`
	SendFor            string                 `json:"send_for"`
	ShouldSendRecovery bool                   `json:"should_send_recovery"`
	Options            map[string]interface{} `json:"options"`
}

// SyncMode configures whether and how a sync runs, per
// apidocs/v2/components/syncs.yaml's Mode schema.
type SyncMode struct {
	Type     string        `json:"type"`
	Triggers *SyncTriggers `json:"triggers,omitempty"`
}

// SyncTriggers holds the trigger configurations for a "triggered" mode
// sync. More than one may be set at once.
type SyncTriggers struct {
	Schedule     *TriggerSchedule     `json:"schedule,omitempty"`
	DbtCloud     *DbtCloudTrigger     `json:"dbt_cloud,omitempty"`
	Fivetran     *FivetranTrigger     `json:"fivetran,omitempty"`
	SyncSequence *SyncSequenceTrigger `json:"sync_sequence,omitempty"`
}

// TriggerSchedule is a schedule-based trigger.
type TriggerSchedule struct {
	Frequency      string `json:"frequency"`
	Day            string `json:"day,omitempty"`
	Hour           int    `json:"hour,omitempty"`
	Minute         int    `json:"minute,omitempty"`
	CronExpression string `json:"cron_expression,omitempty"`
}

// DbtCloudTrigger triggers a sync when a dbt Cloud job succeeds.
type DbtCloudTrigger struct {
	ProjectID string `json:"project_id"`
	JobID     string `json:"job_id"`
}

// FivetranTrigger triggers a sync when a Fivetran job succeeds.
type FivetranTrigger struct {
	JobID   string `json:"job_id"`
	JobName string `json:"job_name"`
}

// SyncSequenceTrigger triggers a sync after another sync completes.
type SyncSequenceTrigger struct {
	SyncID int `json:"sync_id"`
}

// CreateSyncRequest is the body for POST /syncs. ActivationID, Operation,
// SourceAttributes, DestinationAttributes and Mappings are required by the
// v2 API.
type CreateSyncRequest struct {
	ActivationID          string                     `json:"activation_id"`
	Operation             string                     `json:"operation"`
	SourceAttributes      map[string]interface{}     `json:"source_attributes"`
	DestinationAttributes *SyncDestinationAttributes `json:"destination_attributes"`
	Mappings              []MappingAttributes        `json:"mappings"`

	Label                   string                 `json:"label,omitempty"`
	AlertAttributes         []AlertAttribute        `json:"alert_attributes,omitempty"`
	AdvancedConfiguration   map[string]interface{}  `json:"advanced_configuration,omitempty"`
	Mode                    *SyncMode               `json:"mode,omitempty"`
	Paused                  bool                    `json:"paused,omitempty"`
	FieldBehavior           string                  `json:"field_behavior,omitempty"`
	FieldNormalization      string                  `json:"field_normalization,omitempty"`
	FieldOrder              string                  `json:"field_order,omitempty"`
	SyncBehaviorFamily      string                  `json:"sync_behavior_family,omitempty"`
	HighWaterMarkAttribute  string                  `json:"high_water_mark_attribute,omitempty"`
	HistoricalSyncOperation string                  `json:"historical_sync_operation,omitempty"`
	MirrorStrategy          string                  `json:"mirror_strategy,omitempty"`
}

// UpdateSyncRequest is the body for PATCH /syncs/{sync_id}. No field is
// required — only send what should change. Per the spec: "though this
// endpoint is a PATCH and top level attributes respect patch semantics,
// this does not apply to nested objects. Any nested objects that are
// changed must be sent in full, including any unchanged values."
type UpdateSyncRequest struct {
	Label                   string                     `json:"label,omitempty"`
	Operation               string                     `json:"operation,omitempty"`
	SourceAttributes        map[string]interface{}     `json:"source_attributes,omitempty"`
	DestinationAttributes   *SyncDestinationAttributes `json:"destination_attributes,omitempty"`
	Mappings                []MappingAttributes        `json:"mappings,omitempty"`
	AlertAttributes         []AlertAttribute           `json:"alert_attributes,omitempty"`
	AdvancedConfiguration   map[string]interface{}     `json:"advanced_configuration,omitempty"`
	Mode                    *SyncMode                  `json:"mode,omitempty"`
	Paused                  bool                       `json:"paused"`
	FieldBehavior           string                     `json:"field_behavior,omitempty"`
	FieldNormalization      string                     `json:"field_normalization,omitempty"`
	FieldOrder              string                     `json:"field_order,omitempty"`
	SyncBehaviorFamily      string                     `json:"sync_behavior_family,omitempty"`
	HighWaterMarkAttribute  string                     `json:"high_water_mark_attribute,omitempty"`
	HistoricalSyncOperation string                     `json:"historical_sync_operation,omitempty"`
	MirrorStrategy          string                     `json:"mirror_strategy,omitempty"`
}

// syncCreateResult is the `data` shape returned by POST /syncs — unlike
// most v2 create endpoints, it does not return the full Sync object, only
// its ID and activation_id (apidocs/v2/paths/syncs.yaml).
type syncCreateResult struct {
	SyncID       int    `json:"sync_id"`
	ActivationID string `json:"activation_id"`
}

func syncPath(syncID int) string {
	return fmt.Sprintf("/syncs/%d", syncID)
}

func syncWhat(syncID int) string {
	return fmt.Sprintf("sync %d", syncID)
}

// CreateSync creates a sync. The create response only carries the new
// sync's ID and activation_id, not the rest of its state — callers that
// need the full object should follow up with GetSync.
func CreateSync(ctx context.Context, c *Client, req *CreateSyncRequest) (*Sync, error) {
	result, err := Create[syncCreateResult](ctx, c, "/syncs", req, "created sync")
	if err != nil {
		return nil, fmt.Errorf("failed to create sync: %w", err)
	}
	if result.SyncID == 0 {
		return nil, fmt.Errorf("Census API returned a successful create response without a sync ID — the sync may have been created anyway; check Census before retrying to avoid a duplicate")
	}
	return &Sync{ID: result.SyncID, ActivationID: result.ActivationID}, nil
}

// GetSync retrieves a sync by ID.
func GetSync(ctx context.Context, c *Client, syncID int) (*Sync, error) {
	sync, err := Get[*Sync](ctx, c, syncPath(syncID), syncWhat(syncID))
	if err != nil {
		return nil, fmt.Errorf("failed to get sync: %w", err)
	}
	if sync == nil {
		return nil, fmt.Errorf("failed to get sync: %w", errEmptyBody(syncWhat(syncID)))
	}
	return sync, nil
}

// UpdateSync updates a sync by ID.
func UpdateSync(ctx context.Context, c *Client, syncID int, req *UpdateSyncRequest) (*Sync, error) {
	sync, err := Update[*Sync](ctx, c, syncPath(syncID), req, syncWhat(syncID))
	if err != nil {
		return nil, fmt.Errorf("failed to update sync: %w", err)
	}
	return sync, nil
}

// DeleteSync deletes a sync by ID.
func DeleteSync(ctx context.Context, c *Client, syncID int) error {
	if err := Delete(ctx, c, syncPath(syncID)); err != nil {
		return fmt.Errorf("failed to delete sync: %w", err)
	}
	return nil
}
