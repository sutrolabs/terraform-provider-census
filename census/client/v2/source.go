package v2

import (
	"context"
	"fmt"
)

type sourceGroupID struct {
	GroupID string `json:"group_id"`
}

// GetSourceGroupID resolves a source's group_id — the string identifier v2
// uses in place of v1's numeric workspace/destination coupling (see
// Activation.group_id in apidocs/v2/components/activations.yaml). Needed by
// Stage 4 (activation import) and Stage 7 (dataset group_id resolution)
// independently of this source's own full CRUD cutover (Stage 6) — it only
// needs GET /sources/{source_id}, not the rest of the source resource.
func GetSourceGroupID(ctx context.Context, c *Client, sourceID int) (string, error) {
	source, err := Get[sourceGroupID](ctx, c, fmt.Sprintf("/sources/%d", sourceID), fmt.Sprintf("source %d", sourceID))
	if err != nil {
		return "", err
	}
	if source.GroupID == "" {
		return "", fmt.Errorf("source %d has no group_id in its API response", sourceID)
	}
	return source.GroupID, nil
}
