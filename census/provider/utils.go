package provider

import (
	"encoding/json"
	"errors"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/sutrolabs/terraform-provider-census/census/client"
)

// writeOnlyCredentialAttributes maps write-only schema arguments to the
// credential keys they populate in the Census connection payload. Write-only
// values are read from the raw configuration (never from state) and merged into
// connection_config so secret payloads are not persisted in Terraform state.
var writeOnlyCredentialAttributes = map[string]string{
	"password_wo":               "password",
	"private_key_pkcs8_wo":      "private_key_pkcs8",
	"private_key_passphrase_wo": "private_key_passphrase",
}

// applyWriteOnlyCredentials merges any configured write-only credential values
// into the credentials map. A write-only value that is set (even to an empty
// string) takes precedence over the same key in connection_config.
func applyWriteOnlyCredentials(d *schema.ResourceData, credentials map[string]interface{}) {
	applyWriteOnlyCredentialsFromConfig(d.GetRawConfig(), credentials)
}

// applyWriteOnlyCredentialsFromConfig merges write-only credential values from a
// raw configuration object into the credentials map. It is separated from
// applyWriteOnlyCredentials so it can be unit tested without a live
// *schema.ResourceData.
func applyWriteOnlyCredentialsFromConfig(rawConfig cty.Value, credentials map[string]interface{}) {
	for attr, credentialKey := range writeOnlyCredentialAttributes {
		if value, ok := writeOnlyCredentialValue(rawConfig, attr); ok {
			credentials[credentialKey] = value
		}
	}
}

// writeOnlyCredentialValue reads a write-only string argument from the raw
// configuration. It returns ok=false when the value is absent, null, or not yet
// known (e.g. during planning), since write-only values are never available
// from Terraform state.
func writeOnlyCredentialValue(rawConfig cty.Value, attr string) (string, bool) {
	if rawConfig.IsNull() || !rawConfig.IsKnown() {
		return "", false
	}

	rawType := rawConfig.Type()
	if !rawType.IsObjectType() || !rawType.HasAttribute(attr) {
		return "", false
	}

	value := rawConfig.GetAttr(attr)
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
