# SPEC-048 Eval Plan

## Fixture Pack

Create `internal/namba/testdata/evals/harness/` with:

- `README.md`
- `route_cases.json`
- `prompt_refinement_cases.json`
- `guardrail_cases.json`
- `evidence_manifest_cases.json`
- `pr_review_cases.json`

## Go Eval Tests

Go tests should load JSON fixtures for:

- route command/request outcomes
- execution evidence manifest boundaries
- PR review opt-in behavior
- fixture schema validation

Use existing helpers wherever possible. `route_cases.json` must not validate a
standalone synthetic classifier. It should validate runtime-observable command
and request outcomes, including:

- `expected_category`
- `expected_command`
- `expected_harness_request_kind_or_none`
- `expected_sidecar_persisted`

`ordinary_feature_or_product_plan` means the request remains on the normal
feature-planning path with no harness request sidecar. `planned_fix` means the
request selects `namba fix --command plan` with no new harness runtime enum.

`evidence_manifest_cases.json` must separate raw-schema validation and
builder-normalization validation. Raw cases catch missing or invalid persisted
fields. Builder cases verify deterministic normalization from
`buildExecutionEvidenceManifest` without treating builder defaults as evidence
that a raw persisted fixture is valid.

## Python Hook Eval Tests

Python tests should load hook-related fixtures only when prompt refinement or
guardrail cases need the real `.codex/hooks/namba_codex_guard.py` subprocess.
The hook remains the source of truth.

Hook fixture assertions should account for event-specific output shapes:
`PreToolUse` uses `permissionDecision`, while `PermissionRequest` may use
`decision.behavior` or `systemMessage`.

## CI

The evals should run through existing validation:

- `go test ./...`
- `python3 -m unittest discover -s tests` when Python hook evals exist
- `go vet ./...`
- existing formatting check

No CI step should need network access, GitHub auth, Codex auth, or local user
configuration.

Do not add a separate CI job unless necessary. Prefer extending the existing
`python3 -m unittest discover -s tests`, `go test ./...`, `go vet ./...`, and
formatting chain.

## Manual Validation

Before final handoff:

1. Temporarily flip one expected fixture value.
2. Confirm the relevant eval test fails with a useful case diagnostic.
3. Revert the fixture flip.
4. Search the repository for stale eval fixture references.
5. Confirm no runtime logs, eval reports, or temporary artifacts are committed.
