package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/sutrolabs/terraform-provider-census/census/client/v2"
)

func resourceActivation() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Fivetran Activations activation, which pairs a source with a destination. Creating one also creates its destination connection. Use it in place of `census_destination`.",

		CreateContext: resourceActivationCreate,
		ReadContext:   resourceActivationRead,
		UpdateContext: resourceActivationUpdate,
		DeleteContext: resourceActivationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceActivationImport,
		},

		Schema: withActivationComputedSchema(map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the activation.",
			},
			"group_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The `group_id` of the activation's source connection (see `census_source.group_id`). Changing it replaces the activation.",
			},
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The destination type: the `service_name` of an activation type from the `GET /activation_types` API that can be created through the API. Changing it replaces the activation.",
			},
			"credentials": {
				Type:        schema.TypeMap,
				Required:    true,
				Sensitive:   true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Credentials for the destination connection. The keys depend on `type`. The API does not return credentials, so they are not read back.",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The name of the activation. Defaults to the destination connection's name.",
			},
		}),
	}
}

// withActivationComputedSchema adds the computed attributes shared with the
// data source to s.
func withActivationComputedSchema(s map[string]*schema.Schema) map[string]*schema.Schema {
	for key, value := range activationComputedSchema() {
		s[key] = value
	}
	return s
}

// activationComputedSchema is the read-only state of the activation's
// destination connection, shared by the resource and the data source.
func activationComputedSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"last_test_succeeded": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "Whether the last test of the destination connection succeeded. `false` if it has not been tested.",
		},
		"last_tested_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "When the destination connection was last tested (RFC 3339), or empty if it has not been tested.",
		},
		"last_test_results": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "The steps of the last test of the destination connection.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"step": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The name of the test step.",
					},
					"status": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The status of the step (`success` or `error`).",
					},
					"error_message": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The error message if the step failed.",
					},
				},
			},
		},
		"connection_details": {
			Type:        schema.TypeMap,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "The destination connection's credential values that the API returns. Non-string values are JSON-encoded.",
		},
		"is_file_system_activation": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "Whether the destination is a file system destination (for example S3, SFTP or Google Cloud Storage). Syncs to a file system destination use the file path as the object name.",
		},
	}
}

func resourceActivationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*clients).V2

	activation, err := v2.CreateActivation(ctx, apiClient, &v2.ActivationRequest{
		Activation: v2.ActivationAttributes{
			GroupID:     d.Get("group_id").(string),
			Type:        d.Get("type").(string),
			Name:        d.Get("name").(string),
			Credentials: expandActivationCredentials(d.Get("credentials").(map[string]interface{})),
		},
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(activation.ID)

	return resourceActivationRead(ctx, d, meta)
}

func resourceActivationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*clients).V2

	activation, err := v2.GetActivation(ctx, apiClient, d.Id())
	if err != nil {
		if v2.IsNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	// The API response doesn't include credentials, so they are left as
	// configured.
	return diag.FromErr(flattenActivation(d, activation))
}

func resourceActivationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*clients).V2

	// group_id and type are ForceNew: the API rejects type on update and
	// ignores group_id. Credentials are only sent when they changed.
	attributes := v2.ActivationAttributes{Name: d.Get("name").(string)}
	if d.HasChange("credentials") {
		attributes.Credentials = expandActivationCredentials(d.Get("credentials").(map[string]interface{}))
	}

	if _, err := v2.UpdateActivation(ctx, apiClient, d.Id(), &v2.ActivationRequest{Activation: attributes}); err != nil {
		return diag.FromErr(err)
	}

	return resourceActivationRead(ctx, d, meta)
}

func resourceActivationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*clients).V2

	if err := v2.DeleteActivation(ctx, apiClient, d.Id()); err != nil {
		if v2.IsNotFoundError(err) {
			return nil
		}
		return diag.FromErr(err)
	}

	return nil
}

// resourceActivationImport accepts either an activation ID or a
// `group_id:destination_id` pair, where group_id is the source connection's
// group_id and destination_id is a legacy numeric destination connection ID
// (e.g. from census_destination).
func resourceActivationImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if !strings.Contains(d.Id(), ":") {
		return []*schema.ResourceData{d}, nil
	}

	groupID, destinationID, err := parseActivationImportID(d.Id())
	if err != nil {
		return nil, err
	}

	activation, err := v2.FindActivationForGroupAndDestination(ctx, meta.(*clients).V2, groupID, destinationID)
	if err != nil {
		return nil, err
	}

	d.SetId(activation.ID)
	return []*schema.ResourceData{d}, nil
}

// parseActivationImportID parses a `group_id:destination_id` import ID.
func parseActivationImportID(id string) (groupID string, destinationID int, err error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("invalid import ID %q: expected an activation ID or group_id:destination_id", id)
	}

	groupID = parts[0]
	if groupID == "" {
		return "", 0, fmt.Errorf("invalid import ID %q: group_id must not be empty", id)
	}

	destinationID, err = strconv.Atoi(parts[1])
	if err != nil || destinationID <= 0 {
		return "", 0, fmt.Errorf("invalid import ID %q: destination_id must be a positive integer", id)
	}

	return groupID, destinationID, nil
}

// expandActivationCredentials converts the credentials map for the API.
// Values that are JSON objects or arrays (e.g. from jsonencode) are decoded;
// every other value is sent as the string it was written as, so a secret
// such as "123456" or "true" is never retyped into a number or boolean.
func expandActivationCredentials(credentials map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(credentials))
	for key, value := range credentials {
		result[key] = coerceWriteOnlyCredentialValue(value)
	}
	return result
}

// flattenActivation writes an API activation into d. Shared by the resource
// and the data source.
func flattenActivation(d *schema.ResourceData, activation *v2.Activation) error {
	values := map[string]interface{}{
		"name":     activation.Name,
		"group_id": activation.GroupID,
	}

	dest := activation.Destination
	if dest == nil {
		dest = &v2.ActivationDestination{}
	}

	if dest.Type != "" {
		values["type"] = dest.Type
	}

	lastTestedAt := ""
	if dest.LastTestedAt != nil {
		lastTestedAt = dest.LastTestedAt.Format(time.RFC3339)
	}

	values["last_test_succeeded"] = dest.LastTestSucceeded != nil && *dest.LastTestSucceeded
	values["last_tested_at"] = lastTestedAt
	values["last_test_results"] = flattenActivationTestResults(dest.LastTestResults)
	values["connection_details"] = flattenConnectionDetails(dest.ConnectionDetails)
	values["is_file_system_activation"] = dest.IsFileSystemDestination

	for key, value := range values {
		if err := d.Set(key, value); err != nil {
			return fmt.Errorf("failed to set %s: %w", key, err)
		}
	}
	return nil
}

func flattenActivationTestResults(results *v2.ActivationTestResults) []interface{} {
	if results == nil {
		return []interface{}{}
	}

	steps := make([]interface{}, 0, len(results.Steps))
	for _, step := range results.Steps {
		message := ""
		if step.Error != nil && step.Error.Message != nil {
			message = *step.Error.Message
		}
		steps = append(steps, map[string]interface{}{
			"step":          step.Step,
			"status":        step.Status,
			"error_message": message,
		})
	}
	return steps
}

// flattenConnectionDetails converts the free-form connection_details object
// into a string map. Strings are kept as-is; anything else (numbers, bools,
// nested objects) is JSON-encoded so no information is lost.
func flattenConnectionDetails(details map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(details))
	for key, value := range details {
		switch v := value.(type) {
		case nil:
			continue
		case string:
			result[key] = v
		default:
			encoded, err := json.Marshal(v)
			if err != nil {
				result[key] = fmt.Sprint(v)
				continue
			}
			result[key] = string(encoded)
		}
	}
	return result
}
