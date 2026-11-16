package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/sutrolabs/terraform-provider-census/census/client/v2"
)

func dataSourceActivation() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to look up a Fivetran Activations activation, either by ID or by its source's `group_id` and legacy destination ID.",

		ReadContext: dataSourceActivationRead,

		Schema: withActivationComputedSchema(map[string]*schema.Schema{
			"id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"id", "destination_id"},
				Description:  "The ID of the activation to retrieve. Conflicts with `group_id`/`destination_id`.",
			},
			"destination_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				RequiredWith: []string{"group_id"},
				Description:  "The legacy numeric ID of the activation's destination connection (e.g. a `census_destination` ID). Use with `group_id`.",
			},
			"group_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				RequiredWith: []string{"destination_id"},
				Description:  "The `group_id` of the activation's source connection. Use with `destination_id` to look up an activation by its source and destination.",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The destination type (e.g. `salesforce`).",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the activation.",
			},
		}),
	}
}

func dataSourceActivationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*clients).V2

	var (
		activation *v2.Activation
		err        error
	)

	if id, ok := d.GetOk("id"); ok {
		activation, err = v2.GetActivation(ctx, apiClient, id.(string))
	} else {
		activation, err = v2.FindActivationForGroupAndDestination(ctx, apiClient, d.Get("group_id").(string), d.Get("destination_id").(int))
	}
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(activation.ID)

	return diag.FromErr(flattenActivation(d, activation))
}
