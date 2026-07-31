package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/sutrolabs/terraform-provider-census/census/client"
)

// writeOnlyConnectionConfigAttr is the write-only argument that carries a
// jsonencode-d object of secret connection credentials. It is read from the raw
// configuration (never from state) and merged into connection_config so secret
// payloads are not persisted in Terraform state.
const writeOnlyConnectionConfigAttr = "connection_config_wo"

// applyWriteOnlyCredentials merges the write-only connection_config_wo object
// into the credentials map. Keys set in the write-only object take precedence
// over the same key in connection_config.
func applyWriteOnlyCredentials(d *schema.ResourceData, credentials map[string]interface{}) error {
	return applyWriteOnlyCredentialsFromConfig(d.GetRawConfig(), credentials)
}

// applyWriteOnlyCredentialsFromConfig merges write-only credentials from a raw
// configuration object into the credentials map. It is separated from
// applyWriteOnlyCredentials so it can be unit tested without a live
// *schema.ResourceData.
func applyWriteOnlyCredentialsFromConfig(rawConfig cty.Value, credentials map[string]interface{}) error {
	raw, ok := writeOnlyConnectionConfigJSON(rawConfig)
	if !ok {
		return nil
	}

	if strings.TrimSpace(raw) == "" {
		return nil
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return fmt.Errorf("connection_config_wo must be a jsonencode-d object of credential key/value pairs: %w", err)
	}

	for key, value := range parsed {
		credentials[key] = coerceWriteOnlyCredentialValue(value)
	}

	return nil
}

// coerceWriteOnlyCredentialValue upgrades a string value that is itself a JSON
// object or array (e.g. a BigQuery service-account credential supplied as
// jsonencode({ private_key = file("sa.json") })) into the decoded structure, so
// Census receives it as a real object. Scalar strings are left exactly as
// written — decoding them would retype opaque secrets (e.g. a password of
// "123456" into a number, or "true"/"null" into a bool/null), which is never
// what the user intended.
func coerceWriteOnlyCredentialValue(value interface{}) interface{} {
	s, ok := value.(string)
	if !ok {
		return value
	}

	var decoded interface{}
	if err := json.Unmarshal([]byte(s), &decoded); err != nil {
		return s
	}

	switch decoded.(type) {
	case map[string]interface{}, []interface{}:
		return decoded
	default:
		return s
	}
}

// writeOnlyConnectionConfigJSON reads the raw connection_config_wo string from
// the configuration. It returns ok=false when the value is absent, null, or not
// yet known (e.g. during planning), since write-only values are never available
// from Terraform state.
func writeOnlyConnectionConfigJSON(rawConfig cty.Value) (string, bool) {
	if rawConfig.IsNull() || !rawConfig.IsKnown() {
		return "", false
	}

	rawType := rawConfig.Type()
	if !rawType.IsObjectType() || !rawType.HasAttribute(writeOnlyConnectionConfigAttr) {
		return "", false
	}

	value := rawConfig.GetAttr(writeOnlyConnectionConfigAttr)
	if value.IsNull() || !value.IsKnown() {
		return "", false
	}

	return value.AsString(), true
}

// IsNotFoundError checks if an error is a 404 Not Found error
func IsNotFoundError(err error) bool {
	var apiErr *client.APIError
	if errors.As(err, &apiErr) {
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
