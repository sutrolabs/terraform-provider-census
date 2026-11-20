// Shared plumbing for the regression tests in this package: locating the
// fixture state, locating a pre-built provider binary, and wiring Terraform's
// dev_overrides so it uses that binary instead of fetching from the registry.
package regression_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const fixtureStatePath = "fixtures/pre_migration.tfstate"

// expectedConfigDir holds the real .tf files tracking what the CURRENT
// provider schema expects in order to reproduce the fixture's resources with
// no drift. See expected_config/README.md for why this is a separate,
// independently maintained artifact rather than a copy of
// census/tests/regression/fixtures/pre_migration/*.tf.
const expectedConfigDir = "expected_config"

// copyConfigFiles copies the named files (relative to expectedConfigDir)
// into workdir, for a test to run terraform against.
func copyConfigFiles(t *testing.T, workdir string, filenames ...string) {
	t.Helper()

	for _, name := range filenames {
		content, err := os.ReadFile(filepath.Join(expectedConfigDir, name))
		if err != nil {
			t.Fatalf("failed to read %s/%s: %v", expectedConfigDir, name, err)
		}
		if err := os.WriteFile(filepath.Join(workdir, name), content, 0o644); err != nil {
			t.Fatalf("failed to copy %s into workdir: %v", name, err)
		}
	}
}

// allExpectedConfigFiles lists every .tf file in expectedConfigDir.
func allExpectedConfigFiles(t *testing.T) []string {
	t.Helper()

	matches, err := filepath.Glob(filepath.Join(expectedConfigDir, "*.tf"))
	if err != nil {
		t.Fatalf("failed to list %s: %v", expectedConfigDir, err)
	}
	if len(matches) == 0 {
		t.Fatalf("no .tf files found in %s", expectedConfigDir)
	}

	names := make([]string, len(matches))
	for i, m := range matches {
		names[i] = filepath.Base(m)
	}
	return names
}

// defaultProviderBinaryPath matches the Makefile's BUILD_DIR/BINARY_NAME
// (`make build` produces bin/terraform-provider-census at the repo root).
const defaultProviderBinaryPath = "../../../bin/terraform-provider-census"

// providerBinaryPathEnvVar lets CI (or a local override) point at a binary
// built somewhere other than the default location.
const providerBinaryPathEnvVar = "CENSUS_PROVIDER_BINARY_PATH"

// providerBinaryPath locates an already-built provider binary rather than
// compiling one itself — these tests exercise a built artifact, they don't
// produce one. Run `make build` (or set CENSUS_PROVIDER_BINARY_PATH) first.
func providerBinaryPath(t *testing.T) string {
	t.Helper()

	path := defaultProviderBinaryPath
	if override := os.Getenv(providerBinaryPathEnvVar); override != "" {
		path = override
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("failed to resolve provider binary path %s: %v", path, err)
	}

	if _, err := os.Stat(abs); err != nil {
		t.Skipf("provider binary not found at %s — run `make build` first, or set %s to point at a built binary (%v)",
			abs, providerBinaryPathEnvVar, err)
	}

	return abs
}

// fixtureSyncActivationID is the activation_id every pre-migration sync
// fixture resolves to through fakeCensusSyncAPI below. It must match the
// literal activation_id hardcoded in expected_config/sync.tf: the point of
// mocking the lookup is to make the schema-version backfill land on exactly
// the value that config already expects, so there's no diff.
const fixtureSyncActivationID = "fake_id"

// fakeCensusSyncAPI starts a local HTTP server standing in for the Activations v2
// API's GET /syncs/{id} endpoint. It exists because
// ResourceSyncStateUpgradeV1 (the sync resource's schema-version-1-to-2 state
// upgrader) looks up each sync's activation_id over the network when
// upgrading pre-activation_id state — and that upgrade runs as part of
// decoding prior state, independent of Refresh(false): Terraform always
// calls UpgradeResourceState for any resource instance whose recorded
// schema_version is behind the provider's current one, refresh setting
// notwithstanding. Every request gets the same canned activation_id
// regardless of which sync ID was requested, since all three pre-migration
// sync fixtures are expected to land on fixtureSyncActivationID. Returns a
// base_url to set via the CENSUS_BASE_URL environment variable.
func fakeCensusSyncAPI(t *testing.T) string {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/v2/syncs/") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"success","data":{"id":1,"activation_id":%q,"created_at":"2023-01-01T00:00:00Z","updated_at":"2023-01-01T00:00:00Z"}}`, fixtureSyncActivationID)
	}))
	t.Cleanup(server.Close)

	return server.URL
}

func writeDevOverrides(t *testing.T, workdir, providerBinary string) {
	t.Helper()

	binDir := filepath.Dir(providerBinary)
	rc := `provider_installation {
  dev_overrides {
    "sutrolabs/census" = "` + binDir + `"
  }
  direct {}
}
`
	if err := os.WriteFile(filepath.Join(workdir, "dev.tfrc"), []byte(rc), 0o644); err != nil {
		t.Fatalf("failed to write dev_overrides config: %v", err)
	}
}

// fixtureResourceAddresses lists every resource instance address in a
// Terraform state file by reading its JSON directly. `terraform show` can't
// be used for this: it refuses to render state whose recorded schema_version
// is older than the provider's (it doesn't run state upgraders), which is
// exactly the situation this harness exists to test across schema bumps.
func fixtureResourceAddresses(t *testing.T, state []byte) map[string]bool {
	t.Helper()

	var parsed struct {
		Resources []struct {
			Module    string `json:"module"`
			Mode      string `json:"mode"`
			Type      string `json:"type"`
			Name      string `json:"name"`
			Instances []struct {
				IndexKey interface{} `json:"index_key"`
			} `json:"instances"`
		} `json:"resources"`
	}
	if err := json.Unmarshal(state, &parsed); err != nil {
		t.Fatalf("failed to parse fixture state JSON: %v", err)
	}

	addresses := map[string]bool{}
	for _, r := range parsed.Resources {
		base := r.Type + "." + r.Name
		if r.Mode == "data" {
			base = "data." + base
		}
		if r.Module != "" {
			base = r.Module + "." + base
		}

		for _, inst := range r.Instances {
			switch key := inst.IndexKey.(type) {
			case nil:
				addresses[base] = true
			case float64:
				addresses[fmt.Sprintf("%s[%d]", base, int(key))] = true
			case string:
				addresses[fmt.Sprintf("%s[%q]", base, key)] = true
			default:
				t.Fatalf("unexpected index_key %#v on %s", key, base)
			}
		}
	}
	return addresses
}
