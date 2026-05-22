# Baseline Capture Plan

Phase 0 captures baseline behavior before production refactor movement. Phase 7 repeats the same checks after refactor.

## Required Pre-Refactor Commands

```sh
scripts/quality.sh
go test ./...
go test -race ./...
go run ./cmd/namba eval --suite harness --format json --baseline internal/namba/testdata/evals/harness/baseline.json --fail-on-regression
go run ./cmd/namba report --format json
go test ./internal/namba -run TestE2EWorkflowFixture -count=1
```

If `TestE2EWorkflowFixture` is not the exact test name, inspect `internal/namba/e2e_workflow_fixture_test.go` and run the fixture-backed E2E test directly with `-run`.

## Capture Artifacts

- `wc -l internal/namba/namba.go` before refactor.
- `rg -n "^func |^type |^const |^var " internal/namba/namba.go` before refactor.
- Public help output key phrases for every command.
- Representative stdout/stderr fragments for success, usage error, clarification required, invalid project root, missing file/config, and blocked/unsafe cases.
- JSON outputs from `status --json`, `report --format json`, and `eval --format json`.
- Generated file lists for `plan`, `harness`, `fix --command plan`, `project`, `sync`, `report`, `eval`, `run`, `queue`, `release`, and `init`.
- Schema validation results for report, eval, scorecard, queue evidence, diagnostics evidence, and execution evidence.

## Golden And Fixture Strategy

- Use temp repos for commands with side effects.
- Use fake process, Git, GitHub, release, and Codex adapters.
- Assert key fields and contract phrases rather than full strings.
- Use full golden snapshots only for small stable outputs such as help fragments or JSON fixtures with explicit schema ownership.
- Store fixtures under `internal/namba/testdata/fixtures` and golden fragments under `internal/namba/testdata/golden`.
- Normalize paths before assertions so Windows and POSIX separators do not create false failures.

## Required Post-Refactor Commands

Repeat all pre-refactor commands and additionally run:

```sh
go list ./...
wc -l internal/namba/namba.go
```

Post-refactor `wc -l` must show a material decrease from the Phase 0 baseline, and `go list ./...` must succeed without import cycles.
