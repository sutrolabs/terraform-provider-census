# Expected config

Real Terraform files declaring, with literal values only (no variables, no
real credentials), what the CURRENT provider schema expects in order to
reproduce every resource recorded in `census/tests/fixtures/pre_migration.tfstate`
with zero drift. Used by `TestPreMigrationFixture_PlansWithNoChangesAgainstExpectedConfig`.

This is **not** the same thing as `census/tests/fixtures/pre_migration/*.tf` —
that config is what actually created the fixture, under whatever schema
existed at capture time, and should stay untouched as a historical record.
This directory tracks the *current* schema instead, and has to be updated by
hand whenever a provider change alters what these resources expect (a new
required attribute, a renamed field, etc.). Sometimes a change means there's
no config that reproduces the old resource with zero drift at all — only a
specific, intentional one — in which case the corresponding test needs to
assert that specific change instead of a bare no-op.

`connection_config` is hardcoded to `{}` (not a real credential) on the
source and destination because the fixture state has it redacted to `{}` —
there's no real value left to match against.
