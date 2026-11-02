package provider

import (
	"testing"

	"github.com/sutrolabs/terraform-provider-census/census/client/v2"
)

func TestV2BaseURL(t *testing.T) {
	t.Parallel()

	testCases := map[string]string{
		"": v2.BaseURL,
		"https://app.staging.getcensus.com/api/v1":  "https://app.staging.getcensus.com/api/v2",
		"https://app.staging.getcensus.com/api/v1/": "https://app.staging.getcensus.com/api/v2",
		"http://localhost:3000":                     "http://localhost:3000/api/v2",
	}

	for override, want := range testCases {
		if got := v2BaseURL(override); got != want {
			t.Errorf("v2BaseURL(%q) = %q, want %q", override, got, want)
		}
	}
}
