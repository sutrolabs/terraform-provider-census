package v2

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Dataset is a SQL dataset as returned by the v2 /datasets endpoints.
type Dataset struct {
	ID                 int             `json:"id"`
	Name               string          `json:"name"`
	Type               string          `json:"type"`
	Description        *string         `json:"description"`
	Query              string          `json:"query"`
	GroupID            string          `json:"group_id"`
	ResourceIdentifier string          `json:"resource_identifier"`
	CachedRecordCount  *int            `json:"cached_record_count"`
	Columns            []DatasetColumn `json:"columns"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

// DatasetColumn is a column of a Dataset.
type DatasetColumn struct {
	Name     string `json:"name"`
	DataType string `json:"data_type"`
}

// CreateDatasetRequest is the body for POST /datasets. The query runs
// against the source connection identified by GroupID.
type CreateDatasetRequest struct {
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Description *string `json:"description,omitempty"`
	Query       string  `json:"query"`
	GroupID     string  `json:"group_id"`
}

// UpdateDatasetRequest is the body for PATCH /datasets/{dataset_id}. Only
// the fields that are set are sent, and the API leaves the rest unchanged.
type UpdateDatasetRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Query       *string `json:"query,omitempty"`
}

func datasetPath(datasetID int) string {
	return fmt.Sprintf("/datasets/%d", datasetID)
}

func datasetWhat(datasetID int) string {
	return fmt.Sprintf("dataset %d", datasetID)
}

// CreateDataset creates a dataset.
func CreateDataset(ctx context.Context, c *Client, req *CreateDatasetRequest) (*Dataset, error) {
	dataset, err := Create[*Dataset](ctx, c, "/datasets", req, "created dataset")
	if err != nil {
		return nil, fmt.Errorf("failed to create dataset: %w", err)
	}
	if dataset == nil || dataset.ID == 0 {
		return nil, fmt.Errorf("Activations API returned a successful create response without a dataset ID — the dataset may have been created anyway; check Fivetran before retrying to avoid a duplicate")
	}
	return dataset, nil
}

// GetDataset retrieves a dataset by ID.
func GetDataset(ctx context.Context, c *Client, datasetID int) (*Dataset, error) {
	dataset, err := Get[*Dataset](ctx, c, datasetPath(datasetID), datasetWhat(datasetID))
	if err != nil {
		return nil, fmt.Errorf("failed to get dataset: %w", err)
	}
	if dataset == nil {
		return nil, fmt.Errorf("failed to get dataset: %w", errEmptyBody(datasetWhat(datasetID)))
	}
	return dataset, nil
}

// UpdateDataset updates a dataset by ID.
func UpdateDataset(ctx context.Context, c *Client, datasetID int, req *UpdateDatasetRequest) (*Dataset, error) {
	dataset, err := Update[*Dataset](ctx, c, datasetPath(datasetID), req, datasetWhat(datasetID))
	if err != nil {
		return nil, fmt.Errorf("failed to update dataset: %w", err)
	}
	return dataset, nil
}

// DeleteDataset deletes a dataset by ID.
func DeleteDataset(ctx context.Context, c *Client, datasetID int) error {
	if err := Delete(ctx, c, datasetPath(datasetID)); err != nil {
		return fmt.Errorf("failed to delete dataset: %w", err)
	}
	return nil
}

// RefreshDatasetColumns starts a column refresh for a dataset and returns
// the key to poll GetDatasetRefreshStatus with. The response is a bare
// `{"refresh_key": int}`, with no `{status, data}` envelope.
func RefreshDatasetColumns(ctx context.Context, c *Client, datasetID int) (int, error) {
	var result struct {
		RefreshKey int `json:"refresh_key"`
	}
	if err := doBare(ctx, c, http.MethodPost, datasetPath(datasetID)+"/refresh_columns", &result); err != nil {
		return 0, fmt.Errorf("failed to refresh columns: %w", err)
	}
	return result.RefreshKey, nil
}

// GetDatasetRefreshStatus returns the state of a column refresh started by
// RefreshDatasetColumns: "processing" or "completed". The response is a bare
// `{"status": ...}`, where status is the job's state rather than the
// envelope's request outcome. A failed refresh is returned as an *APIError
// (HTTP 400) carrying the reason.
func GetDatasetRefreshStatus(ctx context.Context, c *Client, datasetID int, refreshKey int) (string, error) {
	var result struct {
		Status string `json:"status"`
	}
	path := fmt.Sprintf("%s/refresh_columns_status?refresh_key=%d", datasetPath(datasetID), refreshKey)
	if err := doBare(ctx, c, http.MethodGet, path, &result); err != nil {
		return "", fmt.Errorf("failed to get column refresh status: %w", err)
	}
	return result.Status, nil
}

// doBare sends a bodyless request and decodes a response that has no
// `{status, data}` envelope into out.
func doBare(ctx context.Context, c *Client, method, path string, out interface{}) error {
	resp, err := c.do(ctx, method, path, nil)
	if err != nil {
		return err
	}
	body, err := readBody(resp)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("failed to decode response JSON: %w\n\nRaw API response:\n%s", err, string(body))
	}
	return nil
}
