package provider

import (
	"encoding/json"

	"github.com/sutrolabs/terraform-provider-census/internal/client"
)

// IsNotFoundError checks if an error is a 404 Not Found error
func IsNotFoundError(err error) bool {
	if apiErr, ok := err.(*client.APIError); ok {
		return apiErr.StatusCode == 404
	}
	return false
}

// expandConnectionConfig converts Terraform map to the format expected by the API
func expandConnectionConfig(config map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range config {
		switch v := value.(type) {
		case string:
			// Try to parse as JSON first (for complex configuration)
			var jsonValue interface{}
			if json.Unmarshal([]byte(v), &jsonValue) == nil {
				result[key] = jsonValue
			} else {
				// Not valid JSON, use as string
				result[key] = v
			}
		default:
			result[key] = v
		}
	}

	return result
}
