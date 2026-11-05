package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDestination() *schema.Resource {
	return &schema.Resource{
		DeprecationMessage: "This resource is deprecated and no longer manages anything in Fivetran. Use the census_activation resource instead.",

		Description: "Managed a Fivetran Activations (Census) data destination connection.",

		CreateContext: resourceDestinationCreate,
		ReadContext:   resourceDestinationRead,
		UpdateContext: resourceDestinationUpdate,
		DeleteContext: resourceDestinationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		// Every attribute stays in the schema so state written by earlier provider
		// versions remains readable. None of them force replacement: Update is a
		// no-op, so a config change must never plan a destroy+create (which would
		// end in a failing Create).
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the destination.",
			},
			"workspace_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The ID of the workspace this destination belongs to. Ignored.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the destination connection.",
			},
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The type of destination (e.g., salesforce, hubspot, postgres).",
			},
			"connection_config": {
				Type:        schema.TypeMap,
				Required:    true,
				Sensitive:   true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Connection configuration for the destination. Contents vary by destination type.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The current status of the destination connection.",
				Deprecated:  "This field is unsupported and will be removed in a future version.",
			},
			"test_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the last connection test.",
				Deprecated:  "This field is unsupported and will be removed in a future version.",
			},
			"last_tested": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp when the destination was last tested.",
				Deprecated:  "This field is unsupported and will be removed in a future version.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp when the destination was created.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp when the destination was last updated.",
				Deprecated:  "This field is unsupported and will be removed in a future version.",
			},
			"auto_refresh_objects": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether to automatically refresh object metadata after creation.",
			},
		},
	}
}

func resourceDestinationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.Errorf("census_destination can no longer be created; destination management has been removed from this provider")
}

func resourceDestinationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceDestinationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceDestinationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
