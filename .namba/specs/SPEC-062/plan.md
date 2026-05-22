# SPEC-062 Plan

## Execution Order

1. Establish the evidence schema contract before touching eval, CI, or package boundaries.
2. Extend eval and report quantification on top of the schema contract.
3. Enforce the same gate locally and in CI, with artifacts.
4. Refactor internal package boundaries only after contract tests protect behavior.
5. Update docs source and regenerate generated docs through `namba sync`.

## Phase 1: Evidence Schema Stabilization

1. Add a schema location such as `internal/namba/schemas/` or another repo-local path that keeps schema files close to the Go contract tests.
2. Define schemas for execution evidence, queue runner evidence, Codex diagnostics evidence, project Codex diagnostics evidence, eval result, eval baseline, eval scorecard, and report JSON.
   - Required versioned queue schema is `queue-runner-evidence/v1`.
   - Queue state remains runtime state without a forced `schema_version` retrofit; add parser compatibility fixtures for `.namba/logs/queue/state.json` instead.
   - Add a separate `namba-eval-scorecard/v1` schema for the 1.0 scorecard.
3. Add representative compatibility fixtures for historic and current evidence under testdata.
4. Add validation helpers that:
   - allow unknown optional fields
   - reject missing required fields
   - reject required field removal
   - reject type changes
   - reject unsupported schema versions
5. Add tests that validate fixtures and negative breakage cases.

Deliverables:

- Schema files.
- Fixture evidence files.
- Compatibility tests.
- Schema validation summary command or test output.

## Phase 2: Quantified Eval And Report Scorecard

1. Map the current eval metrics to the fixed 1.0 names:
   - `route_selection`
   - `clarification_quality`
   - `spec_completeness`
   - `execution_readiness`
   - `review_readiness`
   - `dangerous_command_blocking`
   - `evidence_completeness`
   - `schema_validity`
   - `offline_e2e_workflow_health`
   Keep existing `namba-eval-results/v1` metric names backward compatible; generate the fixed names in `namba-eval-scorecard/v1`.
2. Add adapters or scenario families for schema validity, report JSON validity, and offline E2E workflow health.
3. Add a scorecard JSON output path and Markdown summary path without breaking existing `namba eval --format json` and `--format markdown` behavior.
4. Update baseline comparison to fail on metric regression, missing required bucket, disappeared scenario, and schema incompatibility.
5. Validate `namba report --format json` against the report schema and include report health in the scorecard or schema validation summary.

Deliverables:

- Eval scorecard JSON contract.
- Eval Markdown summary.
- Updated baseline if required.
- Report JSON validation.

## Phase 3: CI And Local Gate Parity

1. Decide the parity mechanism:
   - preferred: CI invokes `scripts/quality.sh` or a reusable script subset directly
   - acceptable: CI split jobs are mapped one-to-one to `scripts/quality.sh` steps with a parity test or checked documentation
2. Extend `scripts/quality.sh` with schema validation, report JSON validation, and scorecard artifact generation.
3. Extend `.github/workflows/ci.yml` so the CI gate enforces:
   - gofmt
   - `go vet`
   - `go test ./...`
   - `go test -race ./...`
   - coverage threshold
   - staticcheck
   - govulncheck
   - harness eval regression
   - evidence schema validation
   - report JSON validation
4. Keep docs drift validation out of the Phase 3 exit criteria. Add reusable CI or script support if useful, but Phase 5 owns the docs update and idempotence proof.
5. Upload CI artifacts:
   - coverage profile and text report
   - eval scorecard JSON
   - eval summary Markdown
   - report JSON
   - schema validation result
6. Add tests or docs that prove local and CI gates are equivalent.

Deliverables:

- Updated `scripts/quality.sh`.
- Updated CI workflow.
- Artifact paths and names.
- Parity assertion.

## Phase 4: Internal Package Boundary Refactor

1. Inventory `internal/namba` dependencies before moving code.
2. Extract only behavior-preserving packages, prioritizing:
   - evidence contract
   - eval engine
   - report and observability
   - hook runtime
   - queue and runtime state
   - GitHub handoff
   - release workflow
   - CLI rendering and shared utilities
3. Keep command glue thin and isolated.
4. Add an import boundary or architecture check that prevents package layering regressions.
5. Run before-and-after equivalence tests for CLI output, eval JSON, report JSON, evidence fixtures, and generated docs.

Deliverables:

- Refactored internal packages where justified.
- Architecture check.
- Behavior-preserving validation evidence.

## Phase 5: Docs Sync Validation

1. Update docs source of truth:
   - README renderer or docs config for generated README and guide text
   - project docs generation inputs where needed
   - `.namba/codex/README.md` only if Codex workflow, instruction, hook, or evidence behavior changes
2. Regenerate docs with `namba sync`.
3. Validate generated README and workflow guide content describes the same gate that CI and `scripts/quality.sh` enforce.
4. Prove docs sync idempotence:
   - local path: run `namba sync`, include expected generated outputs, rerun `namba sync`, then confirm no additional generated diff appears
   - CI path: run `namba sync` on the committed PR state and fail with `git diff --exit-code` if generated docs drift
5. Confirm generated docs are not manually patched as the only source of truth.

Deliverables:

- Updated docs source.
- Regenerated README and guides.
- Refreshed project docs.
- Tests or diffs proving docs sync idempotence.

## Final Validation

Run the strongest available gate:

```sh
scripts/quality.sh
namba sync
git diff --check
```

If local tool availability prevents full `scripts/quality.sh`, record the missing tool output and run:

```sh
go test ./...
go vet ./...
gofmt -l cmd internal namba_test.go
go run ./cmd/namba eval --suite harness --format json --baseline internal/namba/testdata/evals/harness/baseline.json --fail-on-regression
go run ./cmd/namba report --format json
```
