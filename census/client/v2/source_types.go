package v2

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// SourceTypeField represents a field configuration for a source type
type SourceTypeField struct {
	ID                    string      `json:"id"`
	Rules                 FieldRules  `json:"rules"`
	Label                 string      `json:"label"`
	Type                  string      `json:"type"`
	Placeholder           interface{} `json:"placeholder"`
	IsPasswordTypeField   bool        `json:"is_password_type_field"`
	PossibleValues        []string    `json:"possible_values,omitempty"`
	Show                  interface{} `json:"show"`
	ConditionallyRequired interface{} `json:"conditionally_required"`
}

// FieldRules is a configuration field's validation rules. The v2 spec
// documents `rules` as a single string (e.g. "required") while v1 returned
// an array, so both shapes are accepted.
type FieldRules []string

func (r *FieldRules) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		if single == "" {
			*r = nil
		} else {
			*r = FieldRules{single}
		}
		return nil
	}

	var list []string
	if err := json.Unmarshal(data, &list); err != nil {
		return fmt.Errorf("rules must be a string or an array of strings: %w", err)
	}
	*r = list
	return nil
}

// SourceTypeConfiguration represents the configuration fields for a source type
type SourceTypeConfiguration struct {
	Fields []SourceTypeField `json:"fields"`
}

// SourceType represents a source type from the API
type SourceType struct {
	DocumentationSlug    string                  `json:"documentation_slug"`
	Label                string                  `json:"label"`
	ServiceName          string                  `json:"service_name"`
	SupportedSyncEngines []string                `json:"supported_sync_engines"`
	CreatableViaAPI      bool                    `json:"creatable_via_api"`
	ConfigurationFields  SourceTypeConfiguration `json:"configuration_fields"`
}

// GetSourceType retrieves a single source type and its field requirements.
// Fetching by service name rather than listing /source_types avoids having
// to page through every source type to find one.
func GetSourceType(ctx context.Context, c *Client, serviceName string) (*SourceType, error) {
	sourceType, err := Get[*SourceType](ctx, c, "/source_types/"+url.PathEscape(serviceName), "source type "+serviceName)
	if err != nil {
		if IsNotFoundError(err) {
			return nil, fmt.Errorf("unknown source type: %s", serviceName)
		}
		return nil, fmt.Errorf("failed to get source type: %w", err)
	}
	if sourceType == nil {
		return nil, errEmptyBody("source type " + serviceName)
	}
	return sourceType, nil
}

// ValidateSourceCredentials validates source credentials against the source type requirements
func ValidateSourceCredentials(ctx context.Context, c *Client, sourceType string, credentials map[string]interface{}) error {
	targetSourceType, err := GetSourceType(ctx, c, sourceType)
	if err != nil {
		return err
	}

	// Validate required fields and set defaults for optional ones
	for _, field := range targetSourceType.ConfigurationFields.Fields {
		// Check if field has required rule
		isRequired := false
		if len(field.Rules) > 0 {
			for _, rule := range field.Rules {
				if rule == "required" || rule == "required:notForEditing" {
					isRequired = true
					break
				}
			}
		}

		// Check for conditional requirements based on 'show' field
		if isRequired && field.Show != nil {
			// Handle conditional requirements like: "show": {"if": "ssh_tunnel_enabled"} or "show": {"if": {"ssh_tunnel_enabled": true}}
			if showMap, ok := field.Show.(map[string]interface{}); ok {
				if ifField, exists := showMap["if"]; exists {
					if ifFieldStr, ok := ifField.(string); ok {
						// Simple condition: "if": "ssh_tunnel_enabled"
						if conditionValue, conditionExists := credentials[ifFieldStr]; conditionExists {
							// Convert to boolean - only require if the condition is explicitly true
							if conditionStr, ok := conditionValue.(string); ok {
								if conditionStr != "true" && conditionStr != "1" {
									isRequired = false
								}
							} else if conditionBool, ok := conditionValue.(bool); ok {
								if !conditionBool {
									isRequired = false
								}
							} else {
								// If condition field is not set or falsy, don't require this field
								isRequired = false
							}
						} else {
							// If condition field doesn't exist, don't require this field
							isRequired = false
						}
					} else if ifFieldMap, ok := ifField.(map[string]interface{}); ok {
						// Complex condition: "if": {"ssh_tunnel_enabled": true}
						conditionMet := false
						for conditionField, expectedValue := range ifFieldMap {
							if conditionValue, conditionExists := credentials[conditionField]; conditionExists {
								// Check if the condition value matches the expected value
								if expectedBool, ok := expectedValue.(bool); ok {
									if conditionStr, ok := conditionValue.(string); ok {
										if expectedBool && (conditionStr == "true" || conditionStr == "1") {
											conditionMet = true
										} else if !expectedBool && (conditionStr == "false" || conditionStr == "0" || conditionStr == "") {
											conditionMet = true
										}
									} else if conditionBool, ok := conditionValue.(bool); ok {
										if conditionBool == expectedBool {
											conditionMet = true
										}
									}
								}
							}
							break // For now, just handle the first condition
						}
						if !conditionMet {
							isRequired = false
						}
					}
				} else if unlessField, exists := showMap["unless"]; exists {
					// Handle "unless" conditions: field is required UNLESS condition is true
					// Example: "show": {"unless": {"use_keypair": {"eq": true}}}
					conditionMet := false

					if unlessFieldMap, ok := unlessField.(map[string]interface{}); ok {
						for conditionField, expectedValue := range unlessFieldMap {
							if conditionValue, conditionExists := credentials[conditionField]; conditionExists {
								// Check if condition matches expected value
								if expectedMap, ok := expectedValue.(map[string]interface{}); ok {
									if eqValue, hasEq := expectedMap["eq"]; hasEq {
										// Handle {"eq": value} comparisons
										if expectedBool, ok := eqValue.(bool); ok {
											// Expected value is boolean
											if condBool, ok := conditionValue.(bool); ok {
												if condBool == expectedBool {
													conditionMet = true
												}
											} else if condStr, ok := conditionValue.(string); ok {
												// Convert string to boolean for comparison
												condBool := condStr == "true" || condStr == "1"
												if condBool == expectedBool {
													conditionMet = true
												}
											}
										} else if expectedStr, ok := eqValue.(string); ok {
											// Expected value is string
											if condStr, ok := conditionValue.(string); ok {
												if condStr == expectedStr {
													conditionMet = true
												}
											}
										}
									}
								} else if expectedBool, ok := expectedValue.(bool); ok {
									// Simple boolean comparison without "eq" wrapper
									if condBool, ok := conditionValue.(bool); ok {
										if condBool == expectedBool {
											conditionMet = true
										}
									} else if condStr, ok := conditionValue.(string); ok {
										condBool := condStr == "true" || condStr == "1"
										if condBool == expectedBool {
											conditionMet = true
										}
									}
								}
							}
							break // Handle first condition for now
						}
					}

					// If "unless" condition is met, field is NOT required
					if conditionMet {
						isRequired = false
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
