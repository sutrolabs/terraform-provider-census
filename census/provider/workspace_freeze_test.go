package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestFrozenWorkspace_CreateErrorsWithoutRequests(t *testing.T) {
	t.Parallel()

	diags := resourceWorkspaceCreate(context.Background(), testWorkspaceResourceData(t), newNoRequestTestClient(t))
	if !diags.HasError() {
		t.Fatal("expected Create to return an error")
	}
}

func TestFrozenWorkspace_ReadUpdateDeleteAreNoOps(t *testing.T) {
	t.Parallel()

	apiClient := newNoRequestTestClient(t)

	for name, fn := range map[string]func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics{
		"read":   resourceWorkspaceRead,
		"update": resourceWorkspaceUpdate,
		"delete": resourceWorkspaceDelete,
	} {
		fn := fn
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			d := testWorkspaceResourceData(t)
			if diags := fn(context.Background(), d, apiClient); len(diags) != 0 {
				t.Fatalf("expected no diagnostics, got %#v", diags)
			}
			if d.Id() != "69962" {
				t.Fatalf("expected ID to be preserved, got %q", d.Id())
			}
		})
	}
}

func TestFrozenWorkspace_DataSourceReturnsWarningWithoutRequests(t *testing.T) {
	t.Parallel()

	d := schema.TestResourceDataRaw(t, dataSourceWorkspace().Schema, map[string]interface{}{
		"id": "69962",
	})

	diags := dataSourceWorkspaceRead(context.Background(), d, newNoRequestTestClient(t))
	if diags.HasError() {
		t.Fatalf("expected no errors, got %#v", diags)
	}
	if len(diags) != 1 || diags[0].Severity != diag.Warning {
		t.Fatalf("expected exactly one warning, got %#v", diags)
	}
	if d.Id() != "69962" {
		t.Fatalf("expected ID to be set from config, got %q", d.Id())
	}
	if got := d.Get("name").(string); got != "" {
		t.Fatalf("expected empty name, got %q", got)
	}
}

func TestFrozenWorkspace_SchemaInternalValidate(t *testing.T) {
	t.Parallel()

	if err := resourceWorkspace().InternalValidate(nil, true); err != nil {
		t.Fatalf("resource schema invalid: %v", err)
	}
	if err := dataSourceWorkspace().InternalValidate(nil, false); err != nil {
		t.Fatalf("data source schema invalid: %v", err)
	}
}
