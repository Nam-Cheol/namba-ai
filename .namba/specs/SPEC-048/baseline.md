# SPEC-048 Baseline

## Existing Behavior Sources

- Harness routing and evidence rules:
  - `internal/namba/harness_contract.go`
  - `internal/namba/harness_contract_test.go`
- Execution evidence manifest behavior:
  - `internal/namba/execution_evidence.go`
  - `internal/namba/execution_evidence_test.go`
- PR review opt-in behavior:
  - `internal/namba/pr_land_command.go`
  - `internal/namba/pr_land_command_test.go`
- Codex lifecycle hook behavior:
  - `.codex/hooks/namba_codex_guard.py`
  - `tests/test_namba_codex_guard.py`
- CI validation:
  - `.github/workflows/ci.yml`

## Current State

The repository already has direct unit tests for the relevant behavior. SPEC-048
adds a fixture-driven eval layer that measures those boundaries with named
golden cases and clearer diagnostics. It should reuse existing helpers where
possible instead of creating a parallel policy implementation.

## Drift Risks

- Fixture evals can duplicate unit tests without adding value.
- Hook evals can drift if they copy Python hook behavior into Go.
- Broad or generated fixture corpora can make the eval pack noisy.
- Tests that touch network, auth, wall clock, local config, or persistent repo
  state can become flaky.
