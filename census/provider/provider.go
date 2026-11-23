package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/sutrolabs/terraform-provider-census/census/client/v2"
)

type clients struct {
	V2 *v2.Client
}

// Version is the provider release version, sent to Fivetran in the User-Agent.
// main sets it from its build-time version.
var Version = "dev"

// Provider returns the Activations Terraform provider
func Provider() *schema.Provider {
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"personal_access_token": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("CENSUS_PERSONAL_ACCESS_TOKEN", nil),
				Description: "Personal access token for the Fivetran Activations API. Can also be set via CENSUS_PERSONAL_ACCESS_TOKEN environment variable.",
			},
			"base_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CENSUS_BASE_URL", ""),
				Description: "Base URL of the Fivetran Activations API, e.g. `https://app.getcensus.com`. A trailing `/api/v1` or `/api/v2` is accepted and ignored. Defaults to `https://app.getcensus.com`. Can also be set via the CENSUS_BASE_URL environment variable.",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"census_workspace":   resourceWorkspace(),
			"census_source":      resourceSource(),
			"census_destination": resourceDestination(),
			"census_sync":        ResourceSync(),
			"census_dataset":     resourceDataset(),
			"census_activation":  resourceActivation(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"census_workspace":   dataSourceWorkspace(),
			"census_source":      dataSourceSource(),
			"census_destination": dataSourceDestination(),
			"census_sync":        dataSourceSync(),
			"census_dataset":     dataSourceDataset(),
			"census_activation":  dataSourceActivation(),
		},
	}

	// p.TerraformVersion is set by the SDK after construction and before
	// configure runs, so capture p rather than reading it now.
	p.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		return configure(ctx, d, p.UserAgent("terraform-provider-census", Version))
	}

	return p
}

func configure(ctx context.Context, d *schema.ResourceData, userAgent string) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	personalToken := d.Get("personal_access_token").(string)
	baseURL := d.Get("base_url").(string)

	// Validate that personal access token is provided
	if personalToken == "" {
		return nil, diag.Errorf("personal_access_token is required")
	}

	v2Root, err := apiRoots(baseURL)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	v2Config := &v2.Config{
		PersonalAccessToken: personalToken,
		BaseURL:             v2Root + "/api/v2",
		UserAgent:           userAgent,
	}

	v2Client, err := v2.NewClient(v2Config)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	return &clients{V2: v2Client}, diags
}

// defaultAPIRoot is the API host used when base_url is unset.
const defaultAPIRoot = "https://app.getcensus.com"

// apiRoots returns the host root the v2 client is built on (<root>/api/v2):
// base_url if set, otherwise defaultAPIRoot.
func apiRoots(baseURL string) (v2Root string, err error) {
	if baseURL != "" {
		root := apiRoot(baseURL)
		return root, nil
	}

	return defaultAPIRoot, nil
}

// apiRoot reduces base_url to a host root. base_url may be given with or
// without a trailing /api/v1 or /api/v2 (e.g. https://app.staging.getcensus.com
// or https://app.staging.getcensus.com/api/v1).
func apiRoot(baseURL string) string {
	root := strings.TrimSuffix(baseURL, "/")
	for _, suffix := range []string{"/api/v2", "/api/v1"} {
		if strings.HasSuffix(root, suffix) {
			return strings.TrimSuffix(root, suffix)
		}
	}
	return root
}
