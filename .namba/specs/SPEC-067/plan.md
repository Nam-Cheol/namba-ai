# SPEC-067 Plan

## Phase 0: Public Behavior Baseline And Checklist

Purpose: Freeze the current public behavior before moving logic.

Change existing files:

- `internal/namba/help_contract_test.go`
- `internal/namba/spec_command_test.go`
- `internal/namba/e2e_workflow_fixture_test.go`
- `internal/namba/eval_command_test.go`
- `internal/namba/report_test.go`
- `internal/namba/queue_command_test.go`
- `internal/namba/pr_land_command_test.go`
- `internal/namba/release_command_test.go`
- `internal/namba/schema_contract_test.go`

Create files:

- `internal/namba/command_behavior_characterization_test.go`
- `internal/namba/testdata/golden/commands/*.txt`
- `internal/namba/testdata/golden/reports/*.json`
- `internal/namba/testdata/fixtures/specs/*`

Do not change public behavior:

- All command names, help meaning, flags, arguments, output contracts, schemas, `.namba` layout, queue semantics, evidence semantics, and PR/land/release behavior.

Risks:

- Overspecifying full stdout snapshots could make harmless copy edits too expensive.
- Underspecifying output could miss automation-facing regressions.

Test method:

- Add contract assertions for key phrases, schema fields, generated paths, exit behavior, and idempotency.
- Avoid full snapshots except for small stable help fragments and JSON fixtures.
- Capture baseline with the commands listed in `baseline.md`.

Rollback criteria:

- If Phase 0 tests are brittle or fail on current `main`, revert Phase 0 test additions and narrow assertions to public contract fields only.

Acceptance:

- `behavior-preservation-checklist.md` is complete.
- Baseline command outputs and artifact expectations are recorded.
- New characterization tests pass on the pre-refactor code.

## Phase 1: Extract Pure Helpers, Usage, And SPEC Report Formatting

Purpose: Move low-risk pure logic out of `namba.go` without changing side effects.

Change existing files:

- `internal/namba/namba.go`
- `internal/namba/help_contract_test.go`
- `internal/namba/spec_command_test.go`

Create files:

- `internal/namba/usage.go`
- `internal/namba/spec_report.go`
- `internal/namba/text_helpers.go`

Do not change public behavior:

- Help output meaning.
- Unknown command errors.
- SPEC creation report content and auto review messaging.
- Multilingual report variants.

Risks:

- Help output ordering changes.
- Report text loses key automation-parsed phrases.

Test method:

- Run help contract tests.
- Run SPEC command tests that assert generated report phrases.
- Run `go test ./internal/namba -run 'TestCommandUsage|TestUsageText|TestSpec' -count=1`.

Rollback criteria:

- If help/report contract tests fail, move extracted helpers back or preserve exact call order.

Acceptance:

- `namba.go` no longer owns usage text or SPEC creation report formatting.
- All Phase 0 command routing and output checks still pass.

## Phase 2: Separate SPEC Creation Responsibilities

Purpose: Isolate plan, harness, and fix-planning scaffold creation from command dispatch.

Change existing files:

- `internal/namba/namba.go`
- `internal/namba/spec_command_test.go`
- `internal/namba/planning_start_test.go`
- `internal/namba/spec_review_test.go`
- `internal/namba/harness_contract_test.go`

Create files:

- `internal/namba/spec_creation.go`
- `internal/namba/spec_scaffold.go`
- `internal/namba/spec_clarification.go`

Do not change public behavior:

- `namba plan`, `namba harness`, and `namba fix --command plan`.
- SPEC ID allocation.
- Dedicated branch behavior and `--current-workspace` behavior.
- Ambiguous request clarification gate.
- Generated SPEC file list and review artifact layout.
- Auto review message.

Risks:

- Planning branch creation and scaffold materialization order changes.
- Clarification-required errors stop writing to stderr/stdout as before.

Test method:

- Run SPEC command tests, planning start tests, harness contract tests, and review artifact tests.
- Add temp-repo characterization for scaffold file list and idempotent failed scaffold rollback.

Rollback criteria:

- If branch/scaffold behavior changes, revert the extraction and keep only pure helper movement from Phase 1.

Acceptance:

- SPEC creation logic is outside `namba.go`.
- Existing and new scaffold contract tests pass.

## Phase 3: Separate Command Handler And Routing

Purpose: Make `namba.go` a thin app boundary instead of the command table and dispatch owner.

Change existing files:

- `internal/namba/namba.go`
- `internal/namba/help_contract_test.go`
- `internal/namba/codex_access_command_test.go`
- `cmd/namba/main_test.go`

Create files:

- `internal/namba/app.go`
- `internal/namba/commands.go`
- `internal/namba/command_errors.go`

Do not change public behavior:

- Top-level command list.
- Internal command visibility rules.
- Help topics.
- Unknown command error wording and exit behavior.
- `--help`, `-h`, and `help <command>` semantics.

Risks:

- Internal command registration leaks into public usage.
- Help routing changes for no-arg or invalid command cases.

Test method:

- Run help contract tests and command main tests.
- Add characterization for every public command help output.

Rollback criteria:

- If command table changes are noisy, keep the command table in `namba.go` and only move `App` construction first.

Acceptance:

- Command definitions and routing no longer live in `namba.go`.
- All public command help checks pass.

## Phase 4: Separate Evidence, Report, Diagnostics, Config, And Project Docs

Purpose: Move file-side-effect and generated artifact helpers into cohesive files while preserving schemas.

Change existing files:

- `internal/namba/namba.go`
- `internal/namba/report.go`
- `internal/namba/codex_diagnostics.go`
- `internal/namba/project_analysis.go`
- `internal/namba/schema_contract.go`
- `internal/namba/report_test.go`
- `internal/namba/schema_contract_test.go`
- `internal/namba/project_analysis_inventory_test.go`

Create files:

- `internal/namba/config.go`
- `internal/namba/manifest.go`
- `internal/namba/project_command.go`
- `internal/namba/project_docs.go`
- `internal/namba/sync_command.go`

Do not change public behavior:

- `namba status`, `namba project`, `namba sync`, `namba report`.
- Report JSON schema.
- Eval/report generated artifacts.
- `.namba/project` and `.namba/logs/project` output layout.
- Missing data graceful degradation.

Risks:

- Managed file manifest classification changes.
- Project docs drift because helper order changes.
- Report JSON field order or schema fields change.

Test method:

- Run report tests, schema contract tests, sync/readme tests, project analysis tests.
- Run `go run ./cmd/namba report --format json` and validate schema tests.

Rollback criteria:

- If generated docs churn unexpectedly, revert movement and isolate only config/manifest helpers.

Acceptance:

- Config, manifest, project docs, sync, and report-facing helpers are no longer mixed into `namba.go`.
- Schema tests and generated artifact checks pass.

## Phase 5: Separate Queue, Runtime, Harness, Run, And Worktree Boundaries

Purpose: Finish moving runtime-facing command glue and worktree handlers out of `namba.go`, aligning with existing runtime and queue files.

Change existing files:

- `internal/namba/namba.go`
- `internal/namba/execution.go`
- `internal/namba/runtime_harness.go`
- `internal/namba/queue_command.go`
- `internal/namba/worktree_command_test.go`
- `internal/namba/execution_test.go`
- `internal/namba/e2e_workflow_fixture_test.go`

Create files:

- `internal/namba/run_command.go`
- `internal/namba/worktree_command.go`
- `internal/namba/direct_fix_command.go`

Do not change public behavior:

- `namba run SPEC-XXX`.
- `--solo`, `--team`, `--parallel`, dry-run or simulation behavior if present.
- Hook evidence, validation evidence, failure evidence, blocked command evidence.
- Queue start/status/resume/pause/stop and durable state.
- Worktree new/list/remove/clean behavior.

Risks:

- Runtime adapter injection changes.
- Worktree cleanup or queue resume ordering changes.
- Evidence file names or schemas drift.

Test method:

- Run execution tests, runtime contract tests, queue tests, parallel lifecycle tests, worktree tests, and E2E fixture tests.
- Verify queue evidence schema through existing schema tests.

Rollback criteria:

- If runtime evidence changes, revert run/worktree movement before touching queue or parallel helpers.

Acceptance:

- Run/worktree/direct-fix command glue is outside `namba.go`.
- Queue/runtime/harness evidence contracts remain green.

## Phase 6: Separate Release, PR, Land, Security, Init, And Update Boundaries

Purpose: Ensure remaining command groups have clear ownership and security-sensitive behavior remains explicit.

Change existing files:

- `internal/namba/namba.go`
- `internal/namba/pr_land_command.go`
- `internal/namba/release_command.go`
- `internal/namba/self_update_command.go`
- `internal/namba/update_command.go`
- `internal/namba/hook_runtime.go`
- `internal/namba/hook_guard_test.go`
- `internal/namba/init_scan.go`
- `internal/namba/init_wizard_test.go`

Create files:

- `internal/namba/init_command.go`
- `internal/namba/init_wizard.go`
- `internal/namba/security_guardrails.go`

Do not change public behavior:

- `namba init`, `namba doctor`, `namba regen`, `namba update`.
- `namba pr`, `namba land`, `namba release`.
- Clean working tree requirements.
- Release checklist and release notes behavior.
- Dangerous command blocking and approval-risk messaging.
- Generated `.codex/hooks` and launcher behavior.

Risks:

- Security-sensitive checks become less visible.
- Init idempotency or generated hook files drift.
- Release/pr/land fake adapter tests stop matching command sequence.

Test method:

- Run init scan/wizard tests, hook guard tests, PR/land tests, release tests, update tests, and readme contract tests.
- Use fake adapters for GitHub, Git, release, and process interactions.

Rollback criteria:

- If security or release behavior changes, revert the security-sensitive extraction and leave that boundary in existing files until tests are strengthened.

Acceptance:

- Init/wizard/security glue is extracted or explicitly isolated.
- PR/land/release/update behavior remains unchanged.

## Phase 7: Final Verification, Dead Code Cleanup, And Dependency Cycle Check

Purpose: Prove behavior preservation and remove leftover duplicate logic.

Change existing files:

- `internal/namba/namba.go`
- Any files touched by earlier phases that retain dead wrappers.

Create files:

- None expected unless dependency graph documentation is useful.

Do not change public behavior:

- All Phase 0 checklist items.

Risks:

- Cleanup accidentally removes compatibility wrappers still used by tests or external command paths.
- Import cycle appears if late subpackage promotion is attempted.

Test method:

- Run the full pre-refactor validation command set again.
- Run new characterization tests and golden/fixture assertions.
- Run `go list ./...` to confirm no import cycles.
- Compare `wc -l internal/namba/namba.go` before and after.
- Re-check Phase 0 checklist manually and through tests.

Rollback criteria:

- If any public behavior regression appears, revert the last phase only and preserve smaller completed phases that remain green.

Acceptance:

- `internal/namba/namba.go` has materially fewer lines and responsibilities than the Phase 0 baseline.
- No dependency cycles.
- Phase 0 checklist is complete after refactor.
- `scripts/quality.sh` passes.
