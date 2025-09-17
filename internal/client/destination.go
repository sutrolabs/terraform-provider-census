package client

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Destination represents a Census data destination
type Destination struct {
	ID          int                    `json:"id"`
	WorkspaceID string                 `json:"workspace_id,omitempty"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Connection  map[string]interface{} `json:"connection"`
	Status      string                 `json:"status,omitempty"`
	TestStatus  string                 `json:"test_status,omitempty"`
	LastTested  *time.Time             `json:"last_tested,omitempty"`
}

// CreateDestinationRequest represents the request to create a destination
type CreateDestinationRequest struct {
	Type              string                `json:"type"`
	ServiceConnection DestinationConnection `json:"service_connection"`
}

// DestinationConnection represents the connection configuration for a destination
type DestinationConnection struct {
	Label       string                 `json:"label"`
	Type        string                 `json:"type"`
	SyncEngine  string                 `json:"sync_engine,omitempty"`
	Credentials map[string]interface{} `json:"credentials"`
}

// UpdateDestinationRequest represents the request to update a destination
type UpdateDestinationRequest struct {
	Name       string                 `json:"name,omitempty"`
	Connection map[string]interface{} `json:"connection,omitempty"`
}

// DestinationResponse represents a single destination response
type DestinationResponse struct {
	Status string       `json:"status"`
	Data   *Destination `json:"data"`
}

// DestinationListResponse represents a paginated destination list response
type DestinationListResponse struct {
	Status     string         `json:"status"`
	Pagination PaginationInfo `json:"pagination"`
	Data       []Destination  `json:"data"`
}

// DestinationObject represents an object within a destination (table, object, etc.)
type DestinationObject struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	FullName    string                 `json:"full_name,omitempty"`
	Fields      []DestinationField     `json:"fields,omitempty"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
}

// DestinationField represents a field within a destination object
type DestinationField struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Required   bool                   `json:"required,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// DestinationObjectsResponse represents the response for destination objects
type DestinationObjectsResponse struct {
	Status string              `json:"status"`
	Data   []DestinationObject `json:"data"`
}

// ObjectCreationRequest represents a request to create an object in a destination
type ObjectCreationRequest struct {
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// ObjectCreationResponse represents the response from object creation
type ObjectCreationResponse struct {
	Status string `json:"status"`
	Data   struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	} `json:"data"`
}

// RefreshObjectsRequest represents a request to refresh destination objects
type RefreshObjectsRequest struct {
	ObjectTypes []string `json:"object_types,omitempty"`
}

// CreateDestination creates a new destination
func (c *Client) CreateDestination(ctx context.Context, req *CreateDestinationRequest) (*Destination, error) {
	return c.CreateDestinationWithToken(ctx, req, "")
}

// CreateDestinationWithToken creates a new destination using a specific workspace token
func (c *Client) CreateDestinationWithToken(ctx context.Context, req *CreateDestinationRequest, workspaceToken string) (*Destination, error) {
	resp, err := c.makeRequestWithToken(ctx, http.MethodPost, "/destinations", req, TokenTypeWorkspace, workspaceToken)
	if err != nil {
		return nil, fmt.Errorf("failed to make create destination request: %w", err)
	}

	var result DestinationResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to create destination: %w", err)
	}

	return result.Data, nil
}

// GetDestination retrieves a destination by ID
func (c *Client) GetDestination(ctx context.Context, destinationID int) (*Destination, error) {
	return c.GetDestinationWithToken(ctx, destinationID, "")
}

// GetDestinationWithToken retrieves a destination by ID using a specific workspace token
func (c *Client) GetDestinationWithToken(ctx context.Context, destinationID int, workspaceToken string) (*Destination, error) {
	path := fmt.Sprintf("/destinations/%d", destinationID)
	resp, err := c.makeRequestWithToken(ctx, http.MethodGet, path, nil, TokenTypeWorkspace, workspaceToken)
	if err != nil {
		return nil, fmt.Errorf("failed to make get destination request: %w", err)
	}

	var result DestinationResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to get destination: %w", err)
	}

	return result.Data, nil
}

// UpdateDestination updates an existing destination
func (c *Client) UpdateDestination(ctx context.Context, destinationID int, req *UpdateDestinationRequest) (*Destination, error) {
	return c.UpdateDestinationWithToken(ctx, destinationID, req, "")
}

// UpdateDestinationWithToken updates an existing destination using a specific workspace token
func (c *Client) UpdateDestinationWithToken(ctx context.Context, destinationID int, req *UpdateDestinationRequest, workspaceToken string) (*Destination, error) {
	path := fmt.Sprintf("/destinations/%d", destinationID)
	resp, err := c.makeRequestWithToken(ctx, http.MethodPatch, path, req, TokenTypeWorkspace, workspaceToken)
	if err != nil {
		return nil, fmt.Errorf("failed to make update destination request: %w", err)
	}

	var result DestinationResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to update destination: %w", err)
	}

	return result.Data, nil
}

// DeleteDestination deletes a destination by ID
func (c *Client) DeleteDestination(ctx context.Context, destinationID int) error {
	return c.DeleteDestinationWithToken(ctx, destinationID, "")
}

// DeleteDestinationWithToken deletes a destination by ID using a specific workspace token
func (c *Client) DeleteDestinationWithToken(ctx context.Context, destinationID int, workspaceToken string) error {
	path := fmt.Sprintf("/destinations/%d", destinationID)
	resp, err := c.makeRequestWithToken(ctx, http.MethodDelete, path, nil, TokenTypeWorkspace, workspaceToken)
	if err != nil {
		return fmt.Errorf("failed to make delete destination request: %w", err)
	}

	if err := c.handleResponse(resp, nil); err != nil {
		return fmt.Errorf("failed to delete destination: %w", err)
	}

	return nil
}

// ListDestinations retrieves a list of destinations
func (c *Client) ListDestinations(ctx context.Context, opts *ListOptions) ([]Destination, *PaginationInfo, error) {
	return c.ListDestinationsWithToken(ctx, opts, "")
}

// ListDestinationsWithToken retrieves a list of destinations using a specific workspace token
func (c *Client) ListDestinationsWithToken(ctx context.Context, opts *ListOptions, workspaceToken string) ([]Destination, *PaginationInfo, error) {
	params := make(map[string]string)
	if opts != nil {
		params = opts.ToParams()
	}

	fullURL := c.buildURL("/destinations", params)
	resp, err := c.makeRequestWithToken(ctx, http.MethodGet, fullURL, nil, TokenTypeWorkspace, workspaceToken)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to make list destinations request: %w", err)
	}

	var result DestinationListResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, nil, fmt.Errorf("failed to list destinations: %w", err)
	}

	return result.Data, &result.Pagination, nil
}

// GetDestinationObjects retrieves objects for a destination
func (c *Client) GetDestinationObjects(ctx context.Context, destinationID int) ([]DestinationObject, error) {
	return c.GetDestinationObjectsWithToken(ctx, destinationID, "")
}

// GetDestinationObjectsWithToken retrieves objects for a destination using a specific workspace token
func (c *Client) GetDestinationObjectsWithToken(ctx context.Context, destinationID int, workspaceToken string) ([]DestinationObject, error) {
	path := fmt.Sprintf("/destinations/%d/objects", destinationID)
	resp, err := c.makeRequestWithToken(ctx, http.MethodGet, path, nil, TokenTypeWorkspace, workspaceToken)
	if err != nil {
		return nil, fmt.Errorf("failed to make get destination objects request: %w", err)
	}

	var result DestinationObjectsResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to get destination objects: %w", err)
	}

	return result.Data, nil
}

// CreateDestinationObject creates a new object in a destination
func (c *Client) CreateDestinationObject(ctx context.Context, destinationID int, req *ObjectCreationRequest) (string, error) {
	return c.CreateDestinationObjectWithToken(ctx, destinationID, req, "")
}

// CreateDestinationObjectWithToken creates a new object in a destination using a specific workspace token
func (c *Client) CreateDestinationObjectWithToken(ctx context.Context, destinationID int, req *ObjectCreationRequest, workspaceToken string) (string, error) {
	path := fmt.Sprintf("/destinations/%d/object_creation_requests", destinationID)
	resp, err := c.makeRequestWithToken(ctx, http.MethodPost, path, req, TokenTypeWorkspace, workspaceToken)
	if err != nil {
		return "", fmt.Errorf("failed to make create object request: %w", err)
	}

	var result ObjectCreationResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return "", fmt.Errorf("failed to create destination object: %w", err)
	}

	return result.Data.ID, nil
}

// RefreshDestinationObjects starts a refresh of destination objects
func (c *Client) RefreshDestinationObjects(ctx context.Context, destinationID int, req *RefreshObjectsRequest) error {
	return c.RefreshDestinationObjectsWithToken(ctx, destinationID, req, "")
}

// RefreshDestinationObjectsWithToken starts a refresh of destination objects using a specific workspace token
func (c *Client) RefreshDestinationObjectsWithToken(ctx context.Context, destinationID int, req *RefreshObjectsRequest, workspaceToken string) error {
	path := fmt.Sprintf("/destinations/%d/refresh_objects", destinationID)
	
	var body interface{}
	if req != nil {
		body = req
	}
	
	resp, err := c.makeRequestWithToken(ctx, http.MethodPost, path, body, TokenTypeWorkspace, workspaceToken)
	if err != nil {
		return fmt.Errorf("failed to make refresh objects request: %w", err)
	}

	if err := c.handleResponse(resp, nil); err != nil {
		return fmt.Errorf("failed to refresh objects: %w", err)
	}

	return nil
}

// GetDestinationRefreshStatus checks the status of an object refresh
func (c *Client) GetDestinationRefreshStatus(ctx context.Context, destinationID int) (*RefreshStatus, error) {
	return c.GetDestinationRefreshStatusWithToken(ctx, destinationID, "")
}

// GetDestinationRefreshStatusWithToken checks the status of an object refresh using a specific workspace token
func (c *Client) GetDestinationRefreshStatusWithToken(ctx context.Context, destinationID int, workspaceToken string) (*RefreshStatus, error) {
	path := fmt.Sprintf("/destinations/%d/refresh_objects_status", destinationID)
	resp, err := c.makeRequestWithToken(ctx, http.MethodGet, path, nil, TokenTypeWorkspace, workspaceToken)
	if err != nil {
		return nil, fmt.Errorf("failed to make get refresh status request: %w", err)
	}

	var result RefreshStatusResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to get refresh status: %w", err)
	}

	return &result.Data, nil
}

// CreateDestinationConnectLink creates a connect link for destination reauthorization
func (c *Client) CreateDestinationConnectLink(ctx context.Context, destinationID int) (*ConnectLink, error) {
	return c.CreateDestinationConnectLinkWithToken(ctx, destinationID, "")
}

// CreateDestinationConnectLinkWithToken creates a connect link for destination reauthorization using a specific workspace token
func (c *Client) CreateDestinationConnectLinkWithToken(ctx context.Context, destinationID int, workspaceToken string) (*ConnectLink, error) {
	path := fmt.Sprintf("/destinations/%d/connect_links", destinationID)
	resp, err := c.makeRequestWithToken(ctx, http.MethodPost, path, nil, TokenTypeWorkspace, workspaceToken)
	if err != nil {
		return nil, fmt.Errorf("failed to make create connect link request: %w", err)
	}

	var result ConnectLinkResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to create connect link: %w", err)
	}

	return &result.Data, nil
}

// ConnectorField represents a field configuration for a connector type
type ConnectorField struct {
	ID                   string      `json:"id"`
	Rules                interface{} `json:"rules"`
	Label                string      `json:"label"`
	Type                 string      `json:"type"`
	Placeholder          interface{} `json:"placeholder"`
	IsPasswordTypeField  bool        `json:"is_password_type_field"`
	PossibleValues       []string    `json:"possible_values,omitempty"`
	Show                 interface{} `json:"show"`
	ConditionallyRequired interface{} `json:"conditionally_required"`
}

// ConnectorConfiguration represents the configuration fields for a connector
type ConnectorConfiguration struct {
	Fields []ConnectorField `json:"fields"`
}

// Connector represents a connector type from the API
type Connector struct {
	DocumentationSlug   string                  `json:"documentation_slug"`
	Label              string                  `json:"label"`
	ServiceName        string                  `json:"service_name"`
	SupportsTest       bool                    `json:"supports_test"`
	CreatableViaAPI    bool                    `json:"creatable_via_api"`
	ConfigurationFields ConnectorConfiguration `json:"configuration_fields"`
}

// ConnectorsResponse represents the response from /connectors
type ConnectorsResponse struct {
	Status     string         `json:"status"`
	Pagination PaginationInfo `json:"pagination"`
	Data       []Connector    `json:"data"`
}

// GetConnectors retrieves all available connector types and their field requirements
func (c *Client) GetConnectors(ctx context.Context, workspaceToken string) ([]Connector, error) {
	resp, err := c.makeRequestWithToken(ctx, http.MethodGet, "/connectors", nil, TokenTypeWorkspace, workspaceToken)
	if err != nil {
		return nil, fmt.Errorf("failed to make get connectors request: %w", err)
	}

	var result ConnectorsResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to get connectors: %w", err)
	}

	return result.Data, nil
}

// ValidateDestinationCredentials validates destination credentials against the connector requirements
func (c *Client) ValidateDestinationCredentials(ctx context.Context, destinationType string, credentials map[string]interface{}, workspaceToken string) error {
	connectors, err := c.GetConnectors(ctx, workspaceToken)
	if err != nil {
		// Check if it's a 404 error (endpoint not found) - skip validation gracefully
		if apiErr, ok := err.(*APIError); ok && apiErr.StatusCode == 404 {
			// The /connectors endpoint might not be available, skip validation
			return nil
		}
		return fmt.Errorf("failed to get connectors for validation: %w", err)
	}


	// Find the connector type
	var targetConnector *Connector
	for _, connector := range connectors {
		if connector.ServiceName == destinationType {
			targetConnector = &connector
			break
		}
	}

	if targetConnector == nil {
		// If connector type is not found, skip validation (might not be available in this workspace)
		// Set empty defaults for all credentials and return
		return nil
	}

	// Validate required fields and set defaults for optional ones
	for _, field := range targetConnector.ConfigurationFields.Fields {
		// Check if field has required rule
		isRequired := false
		if field.Rules != nil {
			switch rules := field.Rules.(type) {
			case string:
				if rules == "required" || rules == "required:notForEditing" {
					isRequired = true
				}
			case []interface{}:
				for _, rule := range rules {
					if ruleStr, ok := rule.(string); ok {
						if ruleStr == "required" || ruleStr == "required:notForEditing" {
							isRequired = true
							break
						}
					}
				}
			}
		}
		
		if isRequired {
			if _, exists := credentials[field.ID]; !exists {
				return fmt.Errorf("required field '%s' (%s) is missing", field.ID, field.Label)
			}
		} else {
			// Set empty string default for missing optional fields
			if _, exists := credentials[field.ID]; !exists {
				credentials[field.ID] = ""
			}
		}
	}

	return nil
}