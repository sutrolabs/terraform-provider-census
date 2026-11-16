package v2

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// Activation is an activation as returned by the v2 /activations endpoints:
// a pairing of a source connection (identified by group_id) and a
// destination connection.
type Activation struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	GroupID     string                 `json:"group_id"`
	Source      *ActivationSource      `json:"source"`
	Destination *ActivationDestination `json:"destination"`
}

// ActivationSource is the source connection nested in an Activation.
type ActivationSource struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// ActivationDestination is the destination connection nested in an
// Activation.
type ActivationDestination struct {
	ID                      int                    `json:"id"`
	Type                    string                 `json:"type"`
	CreatedAt               *time.Time             `json:"created_at"`
	LastTestSucceeded       *bool                  `json:"last_test_succeeded"`
	LastTestedAt            *time.Time             `json:"last_tested_at"`
	IsFileSystemDestination bool                   `json:"is_file_system_destination"`
	ConnectionDetails       map[string]interface{} `json:"connection_details"`
	LastTestResults         *ActivationTestResults `json:"last_test_results"`
}

// ActivationTestResults is the outcome of the last connection test.
type ActivationTestResults struct {
	Steps []ActivationTestStep `json:"steps"`
}

// ActivationTestStep is one step of a connection test.
type ActivationTestStep struct {
	Step   string                   `json:"step"`
	Status string                   `json:"status"`
	Error  *ActivationTestStepError `json:"error"`
}

// ActivationTestStepError carries a failed step's error message.
type ActivationTestStepError struct {
	Message *string `json:"message"`
}

// ActivationRequest is the body for both POST /activations and PATCH
// /activations/{activation_id}.
// Create requires group_id, type and credentials. Update rejects type with a
// 400 and ignores group_id, so leave both empty there; credentials and name
// are optional on update.
type ActivationRequest struct {
	Activation ActivationAttributes `json:"activation"`
}

// ActivationAttributes carries the activation fields for create and update.
type ActivationAttributes struct {
	GroupID     string                 `json:"group_id,omitempty"`
	Type        string                 `json:"type,omitempty"`
	Name        string                 `json:"name,omitempty"`
	Credentials map[string]interface{} `json:"credentials,omitempty"`
}

// activationListPageSize is the per_page used when paging through
// /activations: the API's maximum page size.
const activationListPageSize = 100

func activationPath(activationID string) string {
	return "/activations/" + url.PathEscape(activationID)
}

func activationWhat(activationID string) string {
	return fmt.Sprintf("activation %s", activationID)
}

// CreateActivation creates an activation.
func CreateActivation(ctx context.Context, c *Client, req *ActivationRequest) (*Activation, error) {
	activation, err := Create[*Activation](ctx, c, "/activations", req, "created activation")
	if err != nil {
		return nil, fmt.Errorf("failed to create activation: %w", err)
	}
	if activation == nil || activation.ID == "" {
		return nil, fmt.Errorf("the API returned a successful create response without an activation ID — the activation may have been created anyway; check Fivetran Activations before retrying to avoid a duplicate")
	}
	return activation, nil
}

// GetActivation retrieves an activation by ID.
func GetActivation(ctx context.Context, c *Client, activationID string) (*Activation, error) {
	activation, err := Get[*Activation](ctx, c, activationPath(activationID), activationWhat(activationID))
	if err != nil {
		return nil, fmt.Errorf("failed to get activation: %w", err)
	}
	if activation == nil {
		return nil, fmt.Errorf("failed to get activation: %w", errEmptyBody(activationWhat(activationID)))
	}
	return activation, nil
}

// UpdateActivation updates an activation by ID.
func UpdateActivation(ctx context.Context, c *Client, activationID string, req *ActivationRequest) (*Activation, error) {
	activation, err := Update[*Activation](ctx, c, activationPath(activationID), req, activationWhat(activationID))
	if err != nil {
		return nil, fmt.Errorf("failed to update activation: %w", err)
	}
	return activation, nil
}

// DeleteActivation deletes an activation by ID. The delete response is a
// bare `{"message": "..."}` rather than the usual envelope, but nothing in it
// is needed, so the generic Delete helper (which drains the body) fits.
func DeleteActivation(ctx context.Context, c *Client, activationID string) error {
	if err := Delete(ctx, c, activationPath(activationID)); err != nil {
		return fmt.Errorf("failed to delete activation: %w", err)
	}
	return nil
}

// ListActivationsForLegacyDestination returns every activation whose
// destination connection has the given legacy numeric ID (e.g. a
// census_destination ID), paging through all results.
func ListActivationsForLegacyDestination(ctx context.Context, c *Client, destinationID int) ([]Activation, error) {
	var all []Activation

	opts := ListOptions{
		PerPage: activationListPageSize,
		Filters: map[string]string{"legacy_destination_id": strconv.Itoa(destinationID)},
	}
	for page := 1; ; page++ {
		opts.Page = page
		activations, pagination, err := List[Activation](ctx, c, "/activations", opts, fmt.Sprintf("activations for destination %d", destinationID))
		if err != nil {
			return nil, fmt.Errorf("failed to list activations: %w", err)
		}

		all = append(all, activations...)

		if pagination.NextPage == nil || len(activations) == 0 {
			return all, nil
		}
	}
}

// FindActivationForGroupAndDestination finds the activation pairing the
// source connection with the given group_id and the destination connection
// with the given legacy numeric ID.
func FindActivationForGroupAndDestination(ctx context.Context, c *Client, groupID string, destinationID int) (*Activation, error) {
	activations, err := ListActivationsForLegacyDestination(ctx, c, destinationID)
	if err != nil {
		return nil, err
	}

	var match *Activation
	for i := range activations {
		if activations[i].GroupID != groupID {
			continue
		}
		if match != nil {
			return nil, fmt.Errorf("found more than one activation for group %s and destination %d (%s and %s)", groupID, destinationID, match.ID, activations[i].ID)
		}
		match = &activations[i]
	}

	if match == nil {
		return nil, fmt.Errorf("no activation found for group %s and destination %d", groupID, destinationID)
	}
	return match, nil
}
