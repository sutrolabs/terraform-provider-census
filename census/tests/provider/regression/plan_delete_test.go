
// Package regression_test checks that the provider's current schema can
// still decode a real, historical Terraform state file without erroring.
//
// It points a scratch Terraform working directory at an already-built
// provider binary via dev_overrides, loads the committed fixture state with
// no matching resource blocks in config, and plans. Since nothing is
// declared in config, every resource in state should
// show up as a clean, no-error "delete" — the same code path that threw
// `no schema available for X while reading state` when a resource type was
// removed from the provider's schema entirely. Anything other than a clean,
// uniform delete plan (an error, or an unexpected action) means the current
// code can no longer read that historical state correctly.
//
// Requires a `terraform` binary on PATH, and a provider binary already built
// via `make build` (or CENSUS_PROVIDER_BINARY_PATH pointing at one) — see
// helpers_test.go. No credentials needed — the plan runs with refresh
// disabled, so the provider never makes a real API call.
package regression_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
)

func TestPreMigrationFixture_PlansCleanlyWithCurrentSchema(t *testing.T) {
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

	// Only main.tf (provider block, no resources) — the point of this test
	// is that nothing is declared in config.
	copyConfigFiles(t, workdir, "main.tf")
	if err := os.WriteFile(filepath.Join(workdir, "terraform.tfstate"), fixtureState, 0o644); err != nil {
		t.Fatalf("failed to copy fixture state into workdir: %v", err)
	}

	// tfexec inherits the full process environment (os.Environ) unless
	// SetEnv is called, which replaces it entirely — so we just add the one
	// variable we need to the current process's environment instead.
	t.Setenv("TF_CLI_CONFIG_FILE", filepath.Join(workdir, "dev.tfrc"))

	ctx := context.Background()
	tf, err := tfexec.NewTerraform(workdir, tfPath)
	if err != nil {
		t.Fatalf("failed to construct terraform-exec client: %v", err)
	}

	if err := tf.Init(ctx); err != nil {
		t.Fatalf("terraform init failed: %v", err)
	}

	// Read the fixture's own resource addresses before planning, so we can
	// confirm afterward that every one of them was actually planned for
	// deletion — not just that whatever the plan did contain was all deletes.
	priorState, err := tf.Show(ctx)
	if err != nil {
		t.Fatalf("failed to read fixture state via terraform show: %v", err)
	}
	expectedAddresses := map[string]bool{}
	if priorState.Values != nil && priorState.Values.RootModule != nil {
		for _, r := range priorState.Values.RootModule.Resources {
			expectedAddresses[r.Address] = true
		}
	}
	if len(expectedAddresses) == 0 {
		t.Fatal("fixture state contains no resources — nothing to validate")
	}

	planPath := filepath.Join(workdir, "plan.out")
	_, err = tf.Plan(ctx, tfexec.Refresh(false), tfexec.Out(planPath))
	if err != nil {
		t.Fatalf("terraform plan failed — the current provider schema could not process the fixture state: %v", err)
	}

	plan, err := tf.ShowPlanFile(ctx, planPath)
	if err != nil {
		t.Fatalf("failed to read plan file: %v", err)
	}

	plannedDeletes := map[string]bool{}
	for _, rc := range plan.ResourceChanges {
		actions := rc.Change.Actions
		if len(actions) != 1 || actions[0] != tfjson.ActionDelete {
			t.Errorf("%s: expected a clean delete (nothing is declared in config), got %v", rc.Address, actions)
			continue
		}
		plannedDeletes[rc.Address] = true
	}

	for addr := range expectedAddresses {
		if !plannedDeletes[addr] {
			t.Errorf("%s is in the fixture state but was not planned for deletion", addr)
		}
	}
	for addr := range plannedDeletes {
		if !expectedAddresses[addr] {
			t.Errorf("plan includes %s, which is not in the fixture state", addr)
		}
	}
}
