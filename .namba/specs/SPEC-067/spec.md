# SPEC-067

## Problem Definition

`internal/namba/namba.go` is the operational center of the NambaAI CLI and has grown into a 5,977 line file. It still owns command routing, public help text, init/profile handling, project/status logic, SPEC creation, planning clarification, direct fix execution, run/sync/worktree orchestration, config loading, manifest writing, project codemap generation, README support document generation, utility helpers, and init wizard rendering.

That concentration makes long-term maintenance difficult: a small behavior change can cross unrelated command contracts, test ownership is hard to see, and impact analysis requires reading a single file that mixes pure formatting, file side effects, command dispatch, runtime execution, and generated artifact logic.

This SPEC is not a safety-checklist-only SPEC. It defines a behavior-preserving refactor that must actually decompose `internal/namba/namba.go` in phases after Phase 0 captures the current public behavior baseline.

## Refactor Goal

Refactor the `internal/namba/namba.go` monolith into maintainable, testable modules while preserving all externally observable CLI behavior.

The final implementation must reduce both the responsibility count and line count of `internal/namba/namba.go`, keep public behavior unchanged, avoid Go import cycles, and prove preservation with the Phase 0 checklist plus the configured quality gates.

## Non-Goals

- No feature additions.
- No CLI UX redesign.
- No command name changes.
- No flag or argument meaning changes.
- No help output meaning changes.
- No JSON schema changes.
- No evidence schema changes.
- No SPEC scaffold or `.namba` layout changes.
- No queue semantic changes.
- No `pr`, `land`, or `release` semantic changes.
- No live GitHub, live Codex, external API, or network-dependent tests.
- No test weakening, failure hiding, coverage threshold relaxation, or `scripts/quality.sh` bypass.

## Current Structure Investigation

Authoritative sources reviewed for this planning pass:

- `.namba/project/product.md`
- `.namba/project/tech.md`
- `.namba/project/mismatch-report.md`
- `.namba/project/quality-report.md`
- `.namba/project/systems/workspace.md`
- `README.md`
- `docs/workflow-guide.ko.md`
- `scripts/quality.sh`
- `.github/workflows/ci.yml`
- `cmd/namba/main.go`
- `internal/namba/namba.go`
- `internal/namba/report.go`
- `internal/namba/queue_command.go`
- `internal/namba/pr_land_command.go`
- `internal/namba/release_command.go`
- `internal/namba/eval_command.go`
- `internal/namba/schema_contract.go`
- existing `internal/namba/*_test.go`

Observed shape:

- `cmd/namba/main.go` is the thin process entrypoint.
- `internal/namba/namba.go` defines `App`, shared config structs, command registration, help text, public command handlers for many commands, planning/SPEC creation, execution prompt generation, sync/project docs, config loading, manifest utilities, discovery helpers, init wizard UI, and assorted low-level utilities.
- Some command groups are already split out (`report.go`, `queue_command.go`, `pr_land_command.go`, `release_command.go`, `eval_command.go`, `self_update_command.go`, `update_command.go`, `codex_access.go`), but `namba.go` remains the central dependency knot.
- Existing tests already cover help contracts, SPEC commands, planning start behavior, queue behavior, PR/land behavior, release behavior, eval/report schemas, execution evidence, hooks, runtime contracts, and E2E workflow fixtures. The refactor must strengthen characterization around public contracts before moving logic.

## `namba.go` Responsibility Map

- App construction and injected process adapters: `App`, `NewApp`.
- Top-level command routing: `Run`, usage rendering, help topic parsing, command definition tables, unknown command errors.
- Public help text for core commands: init, doctor, status, project, regen, update, plan, harness, fix, run, sync, pr, land, release, worktree.
- Command handlers still in `namba.go`: init, doctor, status, project, plan, harness, fix, run, sync, worktree, direct-fix dispatch.
- SPEC creation: planning start resolution integration, scaffold context loading, scaffold output materialization, auto review messaging.
- SPEC creation report formatting: Korean, English, Japanese, Chinese variants, generated path listing, open point/security note text.
- Clarification gate: description parsing, evidence groups, language detection, multilingual question formatting.
- SPEC scaffold builders: spec, plan, acceptance documents for feature, harness, and fix paths.
- Execution and sync setup: run option parsing, run context loading, prompt writing, direct fix prompt construction, sync context/materialization.
- Worktree and parallel run entrypoints: worktree subcommands, parallel dispatch bridge.
- Validation and shell execution: quality command runner, binary runner, project root requirement.
- Config and manifest loading: project, quality, docs, init profile, manifest read/write.
- Project analysis docs/codemaps: structure, tech, entry points, dependencies, language-specific discovery helpers.
- Generated README/support docs: change summary, PR checklist, release notes, release checklist.
- Init wizard: language/profile options, interactive rendering, prompt selection, profile validation.
- Generic utilities: normalization, key-value parsing, path existence, checksums, counters, first-line helpers.

## Public Behavior Freeze

The refactor may change internal files, helper names, package boundaries, and implementation details. It must not change:

- Public command names: `init`, `doctor`, `status`, `project`, `plan`, `harness`, `fix`, `run`, `queue`, `sync`, `pr`, `land`, `release`, `regen`, `update`, `eval`, `report`, `worktree`, and documented help surfaces.
- Alias behavior and help topic behavior where present.
- Required/missing argument behavior, invalid flag behavior, SPEC ID handling, description command handling, range/list handling, defaults, and current-workspace behavior.
- Exit code and error contract for success, usage errors, clarification-required errors, invalid project root errors, missing config/file errors, and blocked/unsafe scenarios.
- stdout/stderr contract, JSON output schemas, markdown/text report output, human summaries, and automation-parsed output.
- `.namba/specs`, `.namba/logs`, `.namba/project`, `.namba/releases`, `.codex/hooks`, evidence files, report/eval artifacts, README/docs managed output, and generated layout.
- SPEC ID allocation, scaffold file list, ambiguous request clarification, auto review copy, and review artifact locations.
- `namba run` evidence, hook evidence, validation evidence, failure evidence, dry-run/simulation behavior, and blocked command evidence.
- Queue start/status/resume/pause/stop semantics, item ordering, pending/running/blocked/done states, blocked reasons, resume conditions, and queue evidence.
- Eval/report baseline regression behavior, scorecard JSON, summary markdown, report JSON schema validation, and graceful degradation.
- Guardrail/security blocking, dangerous command handling, approval-risk messaging, secret/vulnerability gates, and fail-closed unsafe behavior.
- PR/land/release clean working tree requirements, release checklist/evidence, and fake/simulation boundaries for external dependencies.
- Idempotency for init, project, sync, report, and already-initialized repositories.
- Cross-platform path and launcher assumptions for Windows, POSIX shell, PowerShell, path normalization, and temp directories.

Detailed checklists live in:

- `.namba/specs/SPEC-067/behavior-preservation-checklist.md`
- `.namba/specs/SPEC-067/baseline.md`
- `.namba/specs/SPEC-067/contract.md`
- `.namba/specs/SPEC-067/eval-plan.md`
- `.namba/specs/SPEC-067/harness-map.md`

## Target Structure

Start with same-package extraction inside `internal/namba` to minimize circular dependency risk. Promote to subpackages only after Phase 0 tests prove contracts and dependency direction is clear.

Conservative target files:

- `internal/namba/app.go`: `App`, constructor, shared injected adapters.
- `internal/namba/commands.go`: top-level command definitions, routing, unknown-command behavior.
- `internal/namba/usage.go`: usage/help text and help contract helpers.
- `internal/namba/spec_creation.go`: `plan`, `harness`, `fix --command plan`, scaffold context and materialization.
- `internal/namba/spec_report.go`: SPEC creation reports and multilingual report text.
- `internal/namba/spec_clarification.go`: clarification gate parsing and prompt text.
- `internal/namba/spec_scaffold.go`: feature/fix/harness scaffold builders.
- `internal/namba/init_command.go`: init command, profile detection integration, init output.
- `internal/namba/init_wizard.go`: interactive wizard rendering and prompt helpers.
- `internal/namba/project_command.go`: `doctor`, `status`, `project`, project docs/codemap orchestration.
- `internal/namba/sync_command.go`: sync command and support doc materialization.
- `internal/namba/config.go`: config loading, defaults, key-value parsing.
- `internal/namba/manifest.go`: manifest read/write/upsert and managed-file helpers.
- `internal/namba/worktree_command.go`: worktree command handlers now embedded in `namba.go`.
- `internal/namba/project_docs.go`: structure, tech, codemap, README-support doc generation.
- Existing split files remain authoritative for `queue`, `report`, `eval`, `pr`, `land`, `release`, update, self-update, codex access, runtime, evidence, hooks, and schema contracts unless a phase explicitly moves shared helpers.

Potential later subpackages, only if dependency analysis proves acyclic and worthwhile:

- `internal/namba/output` for pure output formatting with no `App` dependency.
- `internal/namba/specs` for pure SPEC ID/scaffold helpers.
- `internal/namba/evidence` for schema-oriented evidence helpers.
- `internal/namba/project` for project analysis builders.

Do not create all candidate packages at once. The default path is same-package files first, then minimal subpackages only where they remove dependency pressure.

## Required Phases

Phase 0 must be completed before moving behavior. Phases 1 through 7 are implementation phases and must include purpose, changed files, new files, unchanged public behavior, risks, tests, rollback criteria, and acceptance criteria. The detailed execution plan is in `.namba/specs/SPEC-067/plan.md`.

## Verification Commands

Pre-refactor and post-refactor validation must use the same baseline command set:

```sh
scripts/quality.sh
go test ./...
go test -race ./...
go run ./cmd/namba eval --suite harness --format json --baseline internal/namba/testdata/evals/harness/baseline.json --fail-on-regression
go run ./cmd/namba report --format json
go test ./internal/namba -run TestE2EWorkflowFixture -count=1
```

If the exact E2E fixture test name differs, locate the fixture-backed E2E test in `internal/namba/e2e_workflow_fixture_test.go` and run it directly with `-run`.

## Acceptance Summary

- Phase 0 public behavior checklist is written before refactor movement.
- Characterization tests and golden/fixture assertions are planned and then added before risky movement.
- `internal/namba/namba.go` responsibility and size decrease.
- No import cycles are introduced.
- Existing tests are not deleted or weakened.
- `scripts/quality.sh` passes.
- `go test ./...`, `go test -race ./...`, harness eval regression, report JSON, and E2E fixture validation pass.
- The Phase 0 checklist is re-run after refactor and shows no public behavior regression.
