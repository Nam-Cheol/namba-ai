# SPEC-058 Baseline

## Current Test Baseline

- `internal/namba` already has command-level tests for planning, execution, queue, PR, land, release, evidence, harness contracts, and sync behavior.
- `project_analysis_regression_test.go` validates `init` plus `project` flows from the exported package boundary.
- `internal/namba/execution_smoke_test.go` is live Codex coverage but is skipped unless `CODEX_SMOKE=1`.
- CI baseline runs `go test ./...`, so default E2E coverage must be offline and deterministic.

## Existing Seams

- `App.runCmd` and `App.runCmdWithInput` can fake external process execution.
- `App.lookPath` can fake dependency presence.
- `App.detectCodexCapabilities` can avoid probing a live Codex binary.
- `App.now` can make timestamps deterministic.
- Existing helpers such as `canonicalTempDir`, `chdirExecution`, `testCodexCapabilities`, `isCodexExec`, `isShellCommand`, and JSON readers provide a starting point.

## Coverage Gap

- No default E2E fixture currently verifies a complete Namba user workflow across bootstrap, project analysis, planning, harness planning, execution evidence, queue state, status output, and PR or land simulation.
- Existing tests validate many individual pieces but can miss cross-command contract drift.
