// Shared plumbing for the regression tests in this package: locating the
// fixture state, locating a pre-built provider binary, and wiring Terraform's
// dev_overrides so it uses that binary instead of fetching from the registry.
package regression_test

import (
	"os"
	"path/filepath"
	"testing"
)

const fixtureStatePath = "./fixtures/pre_migration.tfstate"

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
