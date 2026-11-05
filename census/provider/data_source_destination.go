package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDestination() *schema.Resource {
	return &schema.Resource{
		DeprecationMessage: "This data source is deprecated and no longer returns data from Fivetran Activations. Use the census_activation data source instead.",

		Description: "Use this data source to get information about a Fivetran Activations (Census) destination connection.",

		ReadContext: dataSourceDestinationRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the destination to retrieve.",
			},
			"workspace_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of the workspace this destination belongs to. Ignored.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the destination connection.",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The type of destination (e.g., salesforce, hubspot, postgres).",
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
		},
	}
}

func dataSourceDestinationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.SetId(d.Get("id").(string))

	return diag.Diagnostics{{
		Severity: diag.Warning,
		Summary:  "census_destination data source is deprecated",
		Detail:   "This data source no longer queries Fivetran Activations and returns only its id. Use the census_activation data source instead.",
	}}
}
