package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/sutrolabs/terraform-provider-census/census/client"
)

// newNoRequestTestClient returns a client whose server fails the test on any
// request at all: a frozen resource must never touch the API.
func newNoRequestTestClient(t *testing.T) *client.Client {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("unexpected API request: %s %s", r.Method, r.URL.Path)
		http.NotFound(w, r)
	}))
	t.Cleanup(server.Close)

	apiClient, err := client.NewClient(&client.Config{
		BaseURL:             server.URL,
		PersonalAccessToken: "personal-token",
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	return apiClient
}

func TestFrozenDestination_CreateErrorsWithoutRequests(t *testing.T) {
	t.Parallel()

	diags := resourceDestinationCreate(context.Background(), testDestinationResourceData(t), newNoRequestTestClient(t))
	if !diags.HasError() {
		t.Fatal("expected Create to return an error")
	}
}

func TestFrozenDestination_ReadUpdateDeleteAreNoOps(t *testing.T) {
	t.Parallel()

	apiClient := newNoRequestTestClient(t)

	for name, fn := range map[string]func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics{
		"read":   resourceDestinationRead,
		"update": resourceDestinationUpdate,
		"delete": resourceDestinationDelete,
	} {
		fn := fn
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			d := testDestinationResourceData(t)
			if diags := fn(context.Background(), d, apiClient); len(diags) != 0 {
				t.Fatalf("expected no diagnostics, got %#v", diags)
			}
			if d.Id() != "2311185" {
				t.Fatalf("expected ID to be preserved, got %q", d.Id())
			}
		})
	}
}

func TestFrozenDestination_DataSourceReturnsWarningWithoutRequests(t *testing.T) {
	t.Parallel()

	d := schema.TestResourceDataRaw(t, dataSourceDestination().Schema, map[string]interface{}{
		"id":           "2311185",
		"workspace_id": "69962",
	})

	diags := dataSourceDestinationRead(context.Background(), d, newNoRequestTestClient(t))
	if diags.HasError() {
		t.Fatalf("expected no errors, got %#v", diags)
	}
	if len(diags) != 1 || diags[0].Severity != diag.Warning {
		t.Fatalf("expected exactly one warning, got %#v", diags)
	}
	if d.Id() != "2311185" {
		t.Fatalf("expected ID to be set from config, got %q", d.Id())
	}
	if got := d.Get("name").(string); got != "" {
		t.Fatalf("expected empty name, got %q", got)
	}
}

func TestFrozenDestination_SchemaInternalValidate(t *testing.T) {
	t.Parallel()

	if err := resourceDestination().InternalValidate(nil, true); err != nil {
		t.Fatalf("resource schema invalid: %v", err)
	}
	if err := dataSourceDestination().InternalValidate(nil, false); err != nil {
		t.Fatalf("data source schema invalid: %v", err)
	}
}
