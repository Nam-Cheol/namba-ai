# Evaluation And Validation Plan

## Characterization Tests To Add

- `internal/namba/command_behavior_characterization_test.go`
  - Public command help key phrases.
  - Unknown command errors.
  - Invalid flag and missing argument behavior.
  - Clarification-required planning failure without SPEC creation.
  - Command stdout/stderr routing for representative success and failure cases.

- `internal/namba/spec_scaffold_characterization_test.go`
  - `plan`, `harness`, and `fix --command plan` file lists.
  - SPEC ID allocation.
  - Auto review copy.
  - Current workspace behavior.

- `internal/namba/artifact_layout_characterization_test.go`
  - `.namba/specs`, `.namba/logs`, `.namba/project`, `.namba/releases`, `.codex/hooks`, report/eval output, and evidence paths.

- `internal/namba/idempotency_characterization_test.go`
  - `init`, `project`, `sync`, and `report` repeated execution in temp repos.

## Existing Tests To Preserve And Use

- `internal/namba/help_contract_test.go`
- `internal/namba/spec_command_test.go`
- `internal/namba/planning_start_test.go`
- `internal/namba/e2e_workflow_fixture_test.go`
- `internal/namba/eval_command_test.go`
- `internal/namba/report_test.go`
- `internal/namba/queue_command_test.go`
- `internal/namba/pr_land_command_test.go`
- `internal/namba/release_command_test.go`
- `internal/namba/schema_contract_test.go`
- `internal/namba/hook_guard_test.go`
- `internal/namba/hook_runtime_test.go`
- `internal/namba/runtime_contract_test.go`
- `internal/namba/execution_evidence_test.go`

## Quality Gates

Required before and after refactor:

```sh
scripts/quality.sh
go test ./...
go test -race ./...
go run ./cmd/namba eval --suite harness --format json --baseline internal/namba/testdata/evals/harness/baseline.json --fail-on-regression
go run ./cmd/namba report --format json
go test ./internal/namba -run TestE2EWorkflowFixture -count=1
```

Required after refactor:

```sh
go list ./...
wc -l internal/namba/namba.go
```

## Pass Criteria

- No public behavior regression from `behavior-preservation-checklist.md`.
- New characterization tests pass.
- Existing tests pass without deletion or weakening.
- Schema validation remains green.
- Harness eval shows no regression against `internal/namba/testdata/evals/harness/baseline.json`.
- `scripts/quality.sh` passes with coverage threshold intact.
