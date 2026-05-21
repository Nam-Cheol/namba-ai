# SPEC-059 Plan

## Phase 1: Command Contract

1. Add `report` as a public top-level command.
2. Add `--format markdown|text|json`, `--json`, optional `--spec`, optional `--since`, and optional `--fail-on`.
3. Extend `namba status` parsing to accept only `--json` while preserving current no-arg text output.
4. Add help contract tests for `report` and `status --json`.

## Phase 2: Schema And Parser Core

1. Create report schema types with `schema_version: namba-report/v1`.
2. Create a report collector that reads from a project root and returns partial results plus warnings.
3. Implement safe path normalization for `.namba` relative paths.
4. Implement tolerant JSON and JSONL readers that classify missing, corrupt, and unknown files without crashing.

## Phase 3: Aggregators

1. Aggregate run evidence, execution results, validation reports, validation attempts, and hook-blocked states.
2. Aggregate queue state, queue evidence, heartbeat freshness, queue blockers, and queue per-SPEC counts.
3. Aggregate SPEC completeness and review readiness from `spec.md`, `plan.md`, `acceptance.md`, `harness-request.json`, harness evidence files, and readiness summaries.
4. Aggregate diagnostics from project and run-level Codex diagnostics.
5. Aggregate release readiness from release checklist state and latest release note.

## Phase 4: Renderers

1. Add JSON renderer using deterministic field names and stable ordering where slices are produced.
2. Add markdown/text renderer that leads with health, top issues, queue state, run evidence, review readiness, stale candidates, diagnostics, and next action.
3. Ensure `namba status --json` reuses the same collector and emits a compact subset, not a separate schema path.

## Phase 5: Tests

1. Add parser tests for missing directories, missing files, corrupt JSON, unknown schemas, and partial JSONL corruption.
2. Add run aggregation tests for completed, execution failed, validation failed, hook failed, missing evidence, and ambiguous validation.
3. Add queue aggregation tests for none, pending, running, waiting, blocked, done, and stale heartbeat candidates.
4. Add review readiness tests for clear, pending, missing, malformed, and historical SPECs without review bundles.
5. Add release and diagnostics summary tests.
6. Add CLI rendering tests for markdown/text/JSON.
7. Run the repository-wide Go test suite.

## Phase 6: Sync

1. Run `namba sync` after implementation.
2. Confirm generated project artifacts mention the report surface where relevant.
3. Prepare PR handoff with the report contract and validation evidence.
