// This test is the complement to TestPreMigrationFixture_PlansCleanlyWithCurrentSchema:
// instead of planning with no config (expect a clean delete for everything),
// it plans WITH a matching config present (expect no changes at all). See
// expected_config/README.md for why that config is a separate, independently
// maintained artifact rather than a copy of census/tests/fixtures/pre_migration/*.tf.
package regression_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
)

func TestPreMigrationFixture_PlansWithNoChangesAgainstExpectedConfig(t *testing.T) {
	tfPath, err := exec.LookPath("terraform")
	if err != nil {
		t.Skip("terraform CLI not found on PATH, skipping regression test")
	}

	fixtureState, err := os.ReadFile(fixtureStatePath)
	if err != nil {
		t.Fatalf("failed to read fixture state %s: %v", fixtureStatePath, err)
	}

	workdir := t.TempDir()
	providerBinary := providerBinaryPath(t)
	writeDevOverrides(t, workdir, providerBinary)

	copyConfigFiles(t, workdir, allExpectedConfigFiles(t)...)
	if err := os.WriteFile(filepath.Join(workdir, "terraform.tfstate"), fixtureState, 0o644); err != nil {
		t.Fatalf("failed to copy fixture state into workdir: %v", err)
	}

	t.Setenv("TF_CLI_CONFIG_FILE", filepath.Join(workdir, "dev.tfrc"))

	ctx := context.Background()
	tf, err := tfexec.NewTerraform(workdir, tfPath)
	if err != nil {
		t.Fatalf("failed to construct terraform-exec client: %v", err)
	}

	if err := tf.Init(ctx); err != nil {
		t.Fatalf("terraform init failed: %v", err)
	}

	expectedAddresses := fixtureResourceAddresses(t, fixtureState)
	if len(expectedAddresses) == 0 {
		t.Fatal("fixture state contains no resources — nothing to validate")
	}

	planPath := filepath.Join(workdir, "plan.out")
	_, err = tf.Plan(ctx, tfexec.Refresh(false), tfexec.Out(planPath))
	if err != nil {
		t.Fatalf("terraform plan failed: %v\n\nIf this is a validation error, expected_config.go is out of sync with "+
			"the current provider schema and needs to be updated to match — see the comment at the top of that file.", err)
	}

	plan, err := tf.ShowPlanFile(ctx, planPath)
	if err != nil {
		t.Fatalf("failed to read plan file: %v", err)
	}

	seen := map[string]bool{}
	for _, rc := range plan.ResourceChanges {
		seen[rc.Address] = true
		actions := rc.Change.Actions
		if isSensitivityOnlyUpdate(rc.Change) {
			t.Logf("%s: update changes only which attributes are marked sensitive, not any value — treated as no change", rc.Address)
			continue
		}
		if len(actions) != 1 || actions[0] != tfjson.ActionNoop {
			t.Errorf("%s: expected no changes, got %v — either a real regression, or expected_config.go "+
				"needs to be updated to match an intentional schema change", rc.Address, actions)
		}
	}

	for addr := range expectedAddresses {
		if !seen[addr] {
			t.Errorf("%s is in the fixture state but missing from expected_config.go (no matching resource block, or it wasn't planned at all)", addr)
		}
	}
}

// isSensitivityOnlyUpdate reports whether change is an in-place update that
// changes no values at all — only which attributes Terraform marks as
// sensitive. That happens when a new provider version adds a Sensitive
// attribute: the fixture state's recorded sensitive_attributes predate it,
// and only Terraform core (not a provider state upgrader) can update them.
// This harness plans with Refresh(false), so it sees the stale marks; a real
// `terraform plan` refreshes first, which re-marks the state from the current
// schema, and shows no changes.
func isSensitivityOnlyUpdate(change *tfjson.Change) bool {
	if len(change.Actions) != 1 || change.Actions[0] != tfjson.ActionUpdate {
		return false
	}
	if unknown, ok := change.AfterUnknown.(map[string]interface{}); ok && len(unknown) > 0 {
		return false
	}
	return reflect.DeepEqual(change.Before, change.After)
}
