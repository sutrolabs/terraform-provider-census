package provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/sutrolabs/terraform-provider-census/census/client"
	"github.com/sutrolabs/terraform-provider-census/census/client/v2"
)

const (
	// DefaultRegion is the default Census region
	DefaultRegion = "us"
	// DefaultBaseURL is the default base URL for US region
	DefaultBaseURL = "https://app.getcensus.com/api/v1"
)

type clients struct {
	V1 *client.Client
	V2 *v2.Client
}

// Version is the provider release version, sent to Census in the User-Agent.
// main sets it from its build-time version.
var Version = "dev"

// Provider returns the Census Terraform provider
func Provider() *schema.Provider {
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"personal_access_token": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("CENSUS_PERSONAL_ACCESS_TOKEN", nil),
				Description: "Personal access token for Census APIs. Used for all operations including dynamic workspace token retrieval. Can also be set via CENSUS_PERSONAL_ACCESS_TOKEN environment variable.",
			},
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  DefaultRegion,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile(`^(us|eu)$`),
					"region must be either 'us' or 'eu'",
				),
				Description: "Fivetran Activations region. Defaults to 'us', which is correct for all standard configurations; there is no need to set this as of August 2026. Only affects resources on the v1 API, and only when `base_url` is unset.",
			},
			"base_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CENSUS_BASE_URL", ""),
				Description: "Base URL of the Fivetran Activations API, e.g. `https://app.getcensus.com`. A trailing `/api/v1` or `/api/v2` is accepted and ignored. Defaults to `https://app.getcensus.com`; when unset, `region` still picks the host for resources on the v1 API. Can also be set via the CENSUS_BASE_URL environment variable.",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"census_workspace":   resourceWorkspace(),
			"census_source":      resourceSource(),
			"census_destination": resourceDestination(),
			"census_sync":        ResourceSync(),
			"census_dataset":     resourceDataset(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"census_workspace":   dataSourceWorkspace(),
			"census_source":      dataSourceSource(),
			"census_destination": dataSourceDestination(),
			"census_sync":        dataSourceSync(),
			"census_dataset":     dataSourceDataset(),
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
	region := d.Get("region").(string)
	baseURL := d.Get("base_url").(string)

	// Validate that personal access token is provided
	if personalToken == "" {
		return nil, diag.Errorf("personal_access_token is required")
	}

	v1Root, v2Root, err := apiRoots(region, baseURL)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	v1Config := &client.Config{
		PersonalAccessToken:  personalToken,
		WorkspaceAccessToken: "", // No longer used - workspace tokens are fetched dynamically
		BaseURL:              v1Root + "/api/v1",
		Region:               region,
		UserAgent:            userAgent,
	}

	v1Client, err := client.NewClient(v1Config)
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

	return &clients{V1: v1Client, V2: v2Client}, diags
}

// defaultAPIRoot is the API host used when base_url is unset. The v2 API is
// always reached here, whatever region is; region only picks the v1 host.
const defaultAPIRoot = "https://app.getcensus.com"

// apiRoots returns the host roots the v1 and v2 clients are built on
// (<root>/api/v1 and <root>/api/v2). A set base_url is used for both. Unset,
// v1 uses the host for region and v2 always uses defaultAPIRoot.
func apiRoots(region, baseURL string) (v1Root, v2Root string, err error) {
	if baseURL != "" {
		root := apiRoot(baseURL)
		return root, root, nil
	}

	switch region {
	case "us":
		return defaultAPIRoot, defaultAPIRoot, nil
	case "eu":
		return "https://app-eu.getcensus.com", defaultAPIRoot, nil
	default:
		return "", "", fmt.Errorf("unsupported region: %s", region)
	}
}

// apiRoot reduces base_url to a host root. base_url may be given with or
// without a trailing /api/v1 or /api/v2 (e.g. https://app.staging.getcensus.com
// or https://app.staging.getcensus.com/api/v1).
func apiRoot(baseURL string) string {
	root := strings.TrimSuffix(baseURL, "/")
	for _, suffix := range []string{"/api/v1", "/api/v2"} {
		if strings.HasSuffix(root, suffix) {
			return strings.TrimSuffix(root, suffix)
		}
	}
	return root
}
