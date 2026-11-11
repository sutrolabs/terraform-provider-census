package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/sutrolabs/terraform-provider-census/census/client/v2"
)

func TestAPIRoot(t *testing.T) {
	t.Parallel()

	testCases := map[string]string{
		"https://app.staging.getcensus.com":         "https://app.staging.getcensus.com",
		"https://app.staging.getcensus.com/":        "https://app.staging.getcensus.com",
		"https://app.staging.getcensus.com/api/v1":  "https://app.staging.getcensus.com",
		"https://app.staging.getcensus.com/api/v1/": "https://app.staging.getcensus.com",
		"https://app.staging.getcensus.com/api/v2":  "https://app.staging.getcensus.com",
		"http://localhost:3000/proxy/api/v1":        "http://localhost:3000/proxy",
	}
	for in, want := range testCases {
		if got := apiRoot(in); got != want {
			t.Errorf("apiRoot(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestAPIRoots(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		region, baseURL string
		wantV1, wantV2  string
	}{
		{"us", "", "https://app.getcensus.com", "https://app.getcensus.com"},
		{"eu", "", "https://app-eu.getcensus.com", "https://app.getcensus.com"},
		{"us", "https://app.staging.getcensus.com/api/v1", "https://app.staging.getcensus.com", "https://app.staging.getcensus.com"},
		{"eu", "https://app.staging.getcensus.com", "https://app.staging.getcensus.com", "https://app.staging.getcensus.com"},
	}
	for _, tc := range testCases {
		v1Root, v2Root, err := apiRoots(tc.region, tc.baseURL)
		if err != nil || v1Root != tc.wantV1 || v2Root != tc.wantV2 {
			t.Errorf("apiRoots(%q, %q) = (%q, %q, %v), want (%q, %q)", tc.region, tc.baseURL, v1Root, v2Root, err, tc.wantV1, tc.wantV2)
		}
	}

	if _, _, err := apiRoots("mars", ""); err == nil {
		t.Error("expected an unknown region to be rejected")
	}
}

func configureForTest(t *testing.T, raw map[string]interface{}) (interface{}, diag.Diagnostics) {
	t.Helper()

	raw["personal_access_token"] = "test-token"
	return configure(context.Background(), schema.TestResourceDataRaw(t, Provider().Schema, raw), "test-agent")
}

// TestConfigure_ClientsShareBaseURLRoot checks that both clients call the
// base_url host, under /api/v1 and /api/v2, whichever form base_url takes.
func TestConfigure_ClientsShareBaseURLRoot(t *testing.T) {
	t.Parallel()

	for _, suffix := range []string{"", "/", "/api/v1", "/api/v2"} {
		suffix := suffix
		t.Run("base_url suffix "+suffix, func(t *testing.T) {
			t.Parallel()

			var paths []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				paths = append(paths, r.URL.Path)
				w.Write([]byte(`{"status": "success", "data": {"id": 1, "name": "s", "type": "redshift", "created_at": "2026-01-01T00:00:00Z"}}`))
			}))
			t.Cleanup(server.Close)

			meta, diags := configureForTest(t, map[string]interface{}{"base_url": server.URL + suffix})
			if diags.HasError() {
				t.Fatalf("configure failed: %#v", diags)
			}
			c := meta.(*clients)

			if _, err := v2.GetSource(context.Background(), c.V2, 1); err != nil {
				t.Fatalf("v2 GetSource failed: %v", err)
			}
			c.V1.GetWorkspace(context.Background(), 1)

			want := []string{"/api/v2/sources/1", "/api/v1/workspaces/1"}
			if len(paths) < 2 || paths[0] != want[0] || paths[1] != want[1] {
				t.Fatalf("expected requests to %v, got %v", want, paths)
			}
		})
	}
}
