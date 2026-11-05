package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/sutrolabs/terraform-provider-census/census/client/v2"
)

func dataSourceSource() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get information about a Census source connection.",

		ReadContext: dataSourceSourceRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the source to retrieve.",
			},
			"workspace_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of the workspace this source belongs to. No longer used: sources are looked up with the personal access token alone.",
				Deprecated:  "workspace_id is ignored and will be removed in a future version; remove it from your configuration.",
			},
			"group_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The group ID of the source, used to reference it from activations and datasets.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the source connection.",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The type of source (e.g., snowflake, bigquery, postgres).",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The current status of the source connection.",
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
				Description: "Timestamp when the source was last tested.",
				Deprecated:  "This field is unsupported and will be removed in a future version.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp when the source was created.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp when the source was last updated.",
				Deprecated:  "This field is unsupported and will be removed in a future version.",
			},
		},
	}
}

func dataSourceSourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*clients).V2

	id, err := strconv.Atoi(d.Get("id").(string))
	if err != nil {
		return diag.Errorf("invalid source ID: %s", d.Get("id").(string))
	}

	source, err := v2.GetSource(ctx, apiClient, id)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(source.ID))
	d.Set("name", source.Name)
	d.Set("type", source.Type)
	d.Set("group_id", source.GroupID)
	d.Set("created_at", source.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))

	// status, test_status, last_tested and updated_at are deprecated: the API does not support them.
	// They are set to the same values they have always had (empty, and the zero timestamp for updated_at).
	d.Set("status", "")
	d.Set("test_status", "")
	d.Set("updated_at", "0001-01-01T00:00:00Z")

	return nil
}
