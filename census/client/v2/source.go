package v2

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Source is a Census source as returned by the v2 /sources endpoints.
type Source struct {
	ID                                int                    `json:"id"`
	Name                              string                 `json:"name"`
	Type                              string                 `json:"type"`
	GroupID                           string                 `json:"group_id"`
	SyncEngine                        string                 `json:"sync_engine,omitempty"`
	WarehouseWritebackRetentionInDays *int                   `json:"warehouse_writeback_retention_in_days,omitempty"`
	Region                            string                 `json:"region,omitempty"`
	NotificationEmails                []string               `json:"notification_emails,omitempty"`
	StorageBackend                    map[string]interface{} `json:"storage_backend,omitempty"`
	CreatedAt                         time.Time              `json:"created_at"`
}

// CreateSourceRequest is the body for POST /sources.
type CreateSourceRequest struct {
	Connection SourceConnection `json:"connection"`
}

// UpdateSourceRequest is the body for PATCH /sources/{source_id}.
type UpdateSourceRequest struct {
	Connection SourceConnection `json:"connection"`
}

// SourceConnection carries the source attributes for create and update.
// Optional fields are pointers or nil-able so that leaving them unset omits
// them from the request entirely, rather than sending a zero value the API
// would apply.
type SourceConnection struct {
	Name                              string                 `json:"name"`
	Type                              string                 `json:"type,omitempty"`
	SyncEngine                        string                 `json:"sync_engine,omitempty"`
	WarehouseWritebackRetentionInDays *int                   `json:"warehouse_writeback_retention_in_days,omitempty"`
	Region                            *string                `json:"region,omitempty"`
	NotificationEmails                []string               `json:"notification_emails,omitempty"`
	StorageBackend                    map[string]interface{} `json:"storage_backend,omitempty"`
	Credentials                       map[string]interface{} `json:"credentials"`
}

type sourceGroupID struct {
	GroupID string `json:"group_id"`
}

func sourcePath(sourceID int) string {
	return fmt.Sprintf("/sources/%d", sourceID)
}

func sourceWhat(sourceID int) string {
	return fmt.Sprintf("source %d", sourceID)
}

// CreateSource creates a source.
func CreateSource(ctx context.Context, c *Client, req *CreateSourceRequest) (*Source, error) {
	source, err := Create[*Source](ctx, c, "/sources", req, "created source")
	if err != nil {
		return nil, fmt.Errorf("failed to create source: %w", err)
	}
	if source == nil || source.ID == 0 {
		return nil, fmt.Errorf("Census API returned a successful create response without a source ID — the source may have been created anyway; check Census before retrying to avoid a duplicate")
	}
	return source, nil
}

// GetSource retrieves a source by ID.
func GetSource(ctx context.Context, c *Client, sourceID int) (*Source, error) {
	source, err := Get[*Source](ctx, c, sourcePath(sourceID), sourceWhat(sourceID))
	if err != nil {
		return nil, fmt.Errorf("failed to get source: %w", err)
	}
	if source == nil {
		return nil, fmt.Errorf("failed to get source: %w", errEmptyBody(sourceWhat(sourceID)))
	}
	return source, nil
}

// UpdateSource updates a source by ID.
func UpdateSource(ctx context.Context, c *Client, sourceID int, req *UpdateSourceRequest) (*Source, error) {
	source, err := Update[*Source](ctx, c, sourcePath(sourceID), req, sourceWhat(sourceID))
	if err != nil {
		return nil, fmt.Errorf("failed to update source: %w", err)
	}
	return source, nil
}

// DeleteSource deletes a source by ID.
func DeleteSource(ctx context.Context, c *Client, sourceID int) error {
	if err := Delete(ctx, c, sourcePath(sourceID)); err != nil {
		return fmt.Errorf("failed to delete source: %w", err)
	}
	return nil
}

// RefreshSourceTables starts a table refresh for a source. Unlike most v2
// endpoints this one returns a bare `{"refresh_key": int}` with no
// `{status, data}` envelope, and nothing in it is needed here, so the body
// is drained rather than decoded.
func RefreshSourceTables(ctx context.Context, c *Client, sourceID int) error {
	resp, err := c.do(ctx, http.MethodPost, sourcePath(sourceID)+"/refresh_tables", nil)
	if err != nil {
		return fmt.Errorf("failed to refresh tables: %w", err)
	}
	if err := drain(resp); err != nil {
		return fmt.Errorf("failed to refresh tables: %w", err)
	}
	return nil
}

// GetSourceGroupID resolves a source's group_id — the string identifier v2
// uses in place of v1's numeric workspace/destination coupling (see
// Activation.group_id in apidocs/v2/components/activations.yaml). Needed by
// Stage 4 (activation import) and Stage 7 (dataset group_id resolution).
func GetSourceGroupID(ctx context.Context, c *Client, sourceID int) (string, error) {
	source, err := Get[sourceGroupID](ctx, c, sourcePath(sourceID), sourceWhat(sourceID))
	if err != nil {
		return "", err
	}
	if source.GroupID == "" {
		return "", fmt.Errorf("source %d has no group_id in its API response", sourceID)
	}
	return source.GroupID, nil
}
