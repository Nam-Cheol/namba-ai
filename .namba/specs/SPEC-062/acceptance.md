# Acceptance

## Integrated Scope

- [ ] SPEC-062 remains one integrated SPEC and is not split into multiple SPEC packages.
- [ ] Phase order is preserved: evidence schema stabilization, eval and report quantification, CI and local gate enforcement, internal package boundary refactor, docs sync validation.
- [ ] Each phase has implementation scope, exclusion scope, risks, validation method, deliverables, and phase-specific acceptance.

## Phase 1: Evidence Schema Stabilization

- [ ] Versioned JSON Schema contracts exist for run evidence, queue evidence, diagnostics evidence, eval result, eval baseline, and report JSON.
- [ ] Schema contracts are append-only compatible by default: unknown optional fields are allowed.
- [ ] Required field removal, type changes, schema version incompatibility, and semantic breaking changes fail tests.
- [ ] Historic fixture evidence continues to parse and validate.
- [ ] Compatibility tests include positive fixtures and negative breakage fixtures.

## Phase 2: Quantified Eval And Report Scorecard

- [ ] Machine-readable JSON scorecard and human-readable Markdown summary are produced from the same eval result data.
- [ ] Existing `namba-eval-results/v1` JSON remains backward compatible; fixed 1.0 metric names are emitted through `namba-eval-scorecard/v1` or an equivalent append-only scorecard section.
- [ ] The SPEC implementation documents the mapping from legacy eval metrics to 1.0 scorecard metrics.
- [ ] Scorecard metrics include exactly these 1.0 names: `route_selection`, `clarification_quality`, `spec_completeness`, `execution_readiness`, `review_readiness`, `dangerous_command_blocking`, `evidence_completeness`, `schema_validity`, and `offline_e2e_workflow_health`.
- [ ] Baseline comparison fails on metric regression, missing required bucket, scenario disappearance, and schema incompatibility.
- [ ] `namba report --format json` output validates against the report JSON schema.
- [ ] Existing `namba eval --format json` and `namba eval --format markdown` behavior remains backward compatible.

## Phase 3: CI And Local Gate Enforcement

- [ ] `scripts/quality.sh` is the canonical local gate, or CI has a tested one-to-one equivalence map to it.
- [ ] CI enforces gofmt, `go vet`, `go test ./...`, race tests, coverage threshold, staticcheck, govulncheck, harness eval regression, evidence schema validation, and report JSON validation.
- [ ] Phase 3 does not require docs content updates to be complete; docs sync validation is owned by Phase 5.
- [ ] CI publishes coverage report, eval scorecard JSON, eval summary Markdown, report JSON, and schema validation result artifacts.
- [ ] Any required gate failure blocks CI.
- [ ] Release workflow, security scan, and installer checksum semantics are unchanged.

## Phase 4: Module Boundary Refactor

- [ ] Refactoring is limited to behavior-preserving internal package boundary changes.
- [ ] Candidate boundaries are evaluated for evidence contract, eval engine, report or observability, hook runtime, queue or runtime state, GitHub handoff, release workflow, and CLI rendering or shared utilities.
- [ ] Public CLI behavior, command output contract, JSON contract, generated docs, and existing tests are preserved.
- [ ] An import boundary test or architecture check prevents high-level packages from depending on CLI command glue or unrelated command implementations.
- [ ] Before-and-after equivalence is proven with existing tests, fixture compatibility, report and eval JSON validation, and command output contract tests.

## Phase 5: Docs Sync Validation

- [ ] README bundles and workflow guides directly related to this gate are updated through source-of-truth changes where possible.
- [ ] `.namba/codex/README.md` is updated only if Codex workflow, instruction, hook, or evidence behavior changes.
- [ ] Project docs are refreshed by `namba sync` when the generator updates them.
- [ ] Generated Markdown is refreshed with `namba sync`.
- [ ] Docs describe the same quality gate and CI artifacts that the implementation enforces.
- [ ] Docs sync idempotence is proven by rerunning `namba sync` locally after expected generated outputs are included, and by a CI drift check using `git diff --exit-code` on committed generated output.
- [ ] Generated docs are not manually edited as the only durable source.

## Non-Regression

- [ ] Public CLI command names and usage are unchanged.
- [ ] Existing user workflow is not broken.
- [ ] No new external service dependency is introduced.
- [ ] Offline and local-first eval, report, schema validation, and quality-gate behavior is preserved.

## Required Validation

- [ ] `go test ./...` passes.
- [ ] `go vet ./...` passes.
- [ ] gofmt check over Go sources passes.
- [ ] Harness eval regression passes against the checked-in baseline.
- [ ] Report JSON validation passes.
- [ ] Evidence schema validation passes.
- [ ] Architecture or import boundary check passes.
- [ ] `scripts/quality.sh` passes or missing local tools are explicitly recorded while equivalent CI gate coverage is present.
- [ ] `namba sync` refreshes directly related generated artifacts.
