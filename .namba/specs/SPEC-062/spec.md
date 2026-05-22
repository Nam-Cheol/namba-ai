# SPEC-062

## Problem

NambaAI already has meaningful harness-quality pieces: `namba eval`, `namba report`, execution evidence manifests, queue runner evidence, Codex diagnostics evidence, `scripts/quality.sh`, and GitHub Actions CI. They are useful, but they are not yet one stable 1.0 engineering-quality contract.

The current risk is not a missing end-user feature. The risk is long-term harness drift:

- Evidence JSON shapes exist in Go structs, but not all of them have versioned JSON Schema contracts or backward-compatibility fixture tests.
- `namba eval` quantifies several routing and guardrail dimensions, but the 1.0 scorecard needs fixed metric names, explicit failure conditions, JSON and Markdown summaries, schema validation, and offline workflow health coverage.
- `namba report` emits `namba-report/v1`, but CI does not yet prove report JSON validity as a first-class quality artifact.
- Local quality and CI overlap, but the parity contract between `scripts/quality.sh` and GitHub Actions needs to be explicit and artifact-backed.
- Most Namba core behavior is concentrated in `internal/namba`, which makes evidence contracts, eval engine, reporting, hook runtime, queue state, GitHub handoff, release flow, and CLI rendering harder to evolve independently.
- Generated docs already say to update source renderers and run `namba sync`, but harness-quality changes need an explicit docs sync validation step so README and workflow guides do not drift from the actual gate.

## Goal

Create one integrated, reviewable, implementation-ready SPEC that raises NambaAI harness engineering quality to a 1.0 gate across schema stability, quantified eval/report output, local and CI enforcement, behavior-preserving module boundaries, and regenerated docs.

## Context

- Project: namba-ai
- Project type: existing Go CLI
- Mode: tdd
- Work type: core harness change
- Branch: `spec/SPEC-062-harness-quality-1-0`
- Existing evidence:
  - `scripts/quality.sh` already runs Python unittest discovery, gofmt, `go vet`, `go test`, harness eval regression, race tests, staticcheck, govulncheck, coverage, and a coverage threshold.
  - `.github/workflows/ci.yml` already has separate jobs for baseline tests, analysis, govulncheck, race, and coverage artifacts.
  - `go run ./cmd/namba eval --suite harness --format json --baseline internal/namba/testdata/evals/harness/baseline.json --fail-on-regression` currently passes 26 of 26 scenarios with 0 regressions.
  - `go run ./cmd/namba report --format json` currently emits `namba-report/v1`, but reports missing run evidence, absent queue state, and missing SPEC-062 harness sidecars until this SPEC package is completed.

## Scope

This SPEC covers the following surfaces as one integrated effort:

- Run evidence: execution evidence manifest and related validation attempt evidence.
- Queue evidence: queue runner evidence, queue state, queue reports, and queue fallback evidence.
- Diagnostics evidence: Codex diagnostics evidence under `.namba/logs/project`.
- Eval contracts: eval scenario corpus, eval result JSON, eval baseline JSON, scorecard JSON, and Markdown summary.
- Report contracts: `namba report --format json`, default Markdown output, report JSON validation, and report artifact generation.
- Local quality gate: `scripts/quality.sh` and any helper validator it calls.
- CI quality gate: `.github/workflows/ci.yml` jobs, required steps, and uploaded artifacts.
- Module boundaries: internal packages carved from `internal/namba` only when behavior-preserving and backed by tests.
- Docs: README bundles, workflow guides, `.namba/codex/README.md`, generated project docs, and the README renderer or docs source used by `namba sync`.

Docs scope rule:

- README bundles and workflow guides are in scope whenever the local or CI quality gate text changes.
- `.namba/codex/README.md` is in scope only when Codex workflow, instruction, hook, or evidence behavior changes.
- Generated project docs are refreshed by `namba sync`; implementation should not hand-edit them as the durable source of truth.

## Non-Goals

- Do not rename or remove public CLI commands or flags.
- Do not change release workflow, security scan behavior, installer checksum semantics, or release asset naming.
- Do not add any new external service dependency.
- Do not introduce live Codex, GitHub API, browser, SaaS, telemetry, or LLM-judge dependency into the eval or CI path.
- Do not split this work into multiple SPECs.
- Do not perform semantic rewrites during phase 4. Refactors are behavior-preserving only.
- Do not manually patch generated docs as the durable source of truth when a renderer or config source exists.

## Phase Plan

### Phase 1: Evidence Schema Stabilization

Implementation scope:

- Add versioned JSON Schema files for:
  - `execution-evidence/v1`
  - `queue-runner-evidence/v1`
  - `codex-diagnostics-evidence/v1`
  - `project-codex-diagnostics-evidence/v1`
  - `namba-eval-results/v1`
  - `namba-eval-baseline/v1`
  - `namba-eval-scorecard/v1`
  - `namba-report/v1`
- Add queue state compatibility fixtures without requiring a `schema_version` retrofit in this SPEC. Current `.namba/logs/queue/state.json` is runtime state read by `namba report`; `queue-runner-evidence/v1` is the versioned queue evidence contract.
- Add one local schema validation command or Go test helper that validates current generated samples and checked-in fixtures without network access.
- Add compatibility fixtures under a stable testdata path for representative historic run evidence, queue evidence, diagnostics evidence, eval result, eval baseline, and report JSON.
- Make schemas append-only compatible by default:
  - unknown optional fields are allowed
  - required field removal fails compatibility
  - type changes fail compatibility
  - schema version mismatch fails compatibility unless an explicit migration fixture is present
- Add tests that parse and validate historic fixture evidence using the same production structs or schema validator path used by CI.

Exclusions:

- No runtime behavior changes beyond validation helper plumbing.
- No schema major-version migration in this SPEC.

Risks:

- Over-strict schemas could block harmless optional fields.
- Under-strict schemas could miss breaking changes that Go structs still accept through zero values.

Validation:

- Go tests for every schema and compatibility fixture.
- Negative tests for missing required fields, required field removal, type change, incompatible schema version, and semantic breaking changes.
- `namba report --format json` output validates against `namba-report/v1`.

Artifacts:

- JSON Schema files.
- Compatibility fixtures.
- Schema validation test output.
- Optional CI artifact: schema validation JSON or Markdown summary.

Phase acceptance:

- All listed evidence surfaces have a versioned schema contract.
- Historic fixtures parse and validate.
- Unknown optional fields validate.
- Breaking changes fail tests.

### Phase 2: Eval And Report Quantification

Implementation scope:

- Extend the harness scorecard around these fixed metric names:
  - `route_selection`
  - `clarification_quality`
  - `spec_completeness`
  - `execution_readiness`
  - `review_readiness`
  - `dangerous_command_blocking`
  - `evidence_completeness`
  - `schema_validity`
  - `offline_e2e_workflow_health`
- Produce a machine-readable JSON scorecard and human-readable Markdown summary from the same eval result data.
- Preserve `namba-eval-results/v1` compatibility. Existing `namba eval --format json` metrics can remain append-only and legacy-compatible; the 1.0 metric vocabulary belongs to the derived scorecard contract `namba-eval-scorecard/v1`.
- Define an explicit metric mapping from current result metrics to scorecard metrics:
  - `clarification` maps to `clarification_quality`
  - `spec_required_fields` maps to `spec_completeness`
  - `required_reviews` and PR review scenarios map to `review_readiness`
  - guardrail deny and risk-note scenarios map to `dangerous_command_blocking`
  - evidence manifest scenarios map to `evidence_completeness`
  - schema validation scenarios map to `schema_validity`
  - deterministic local workflow scenarios map to `offline_e2e_workflow_health`
  - `route_selection` and `execution_readiness` keep their names
- Ensure the scorecard includes:
  - suite id and corpus version
  - generated timestamp
  - metric pass counts and pass rates
  - required coverage buckets
  - scenario ids and fingerprints
  - baseline comparison result
  - regression list
  - schema validation status
- Expand scenarios or adapters only where necessary to cover report JSON validity, schema validity, and offline E2E workflow health.
- Keep `namba eval --format json` and `namba eval --format markdown` behavior compatible while adding new fields append-only.
- Add report JSON validation to eval or a separate local validator so `namba report --format json` becomes a tested evidence source.

Failure conditions:

- Any scenario failure.
- Any metric regression against baseline.
- Missing required coverage bucket.
- Scenario disappearance without an intentional baseline update.
- Eval fixture, eval result, eval baseline, scorecard, or report schema incompatibility.

Exclusions:

- No probabilistic scoring.
- No live model judgement.
- No remote service call.

Risks:

- Renaming existing metrics can create noisy baseline churn.
- Offline E2E health must stay deterministic and not rely on wall-clock-sensitive state.

Validation:

- `go run ./cmd/namba eval --suite harness --format json --baseline internal/namba/testdata/evals/harness/baseline.json --fail-on-regression`
- equivalent Markdown summary generation
- scorecard JSON schema validation
- `namba report --format json` schema validation

Artifacts:

- Eval scorecard JSON.
- Eval summary Markdown.
- Updated baseline when intentionally changed.
- Report JSON sample or CI-generated artifact.

Phase acceptance:

- All fixed metric names appear in the JSON scorecard.
- JSON and Markdown summaries agree on totals, failures, and regressions.
- Baseline comparison fails on metric regression, bucket loss, scenario disappearance, or schema incompatibility.

### Phase 3: CI And Local Quality Gate Enforcement

Implementation scope:

- Make `scripts/quality.sh` the canonical local quality gate, or document and test exact CI equivalence when CI must split work into jobs.
- Ensure the CI gate enforces:
  - gofmt
  - `go vet`
  - `go test ./...`
  - race tests
  - coverage threshold
  - staticcheck
  - govulncheck
  - harness eval regression
  - evidence schema validation
  - report JSON validation
- Add CI artifacts:
  - coverage profile and coverage text report
  - eval scorecard JSON
  - eval summary Markdown
  - report JSON
  - schema validation result
- Keep the gate offline/local-first except existing Go module and tool installation behavior already present in CI.
- If CI remains split across jobs, add a parity check or documented mapping from each `scripts/quality.sh` step to the CI job and step that enforces it.
- Do not make Phase 3 depend on Phase 5 docs updates. Phase 3 may add reusable docs drift check plumbing, but Phase 5 owns deciding which docs change and proving regenerated output is current.

Exclusions:

- No release workflow semantic changes.
- No secret scan behavior change.
- No installer checksum behavior change.

Risks:

- CI time can increase if expensive checks duplicate across jobs.
- Tool installation failure can obscure actual harness failures.

Validation:

- Run `scripts/quality.sh` locally when tools are present.
- If tools are missing locally, run the subset available and document the missing tool messages.
- Review GitHub Actions YAML to ensure every local gate step has a CI equivalent.
- CI artifact paths are stable and uploaded on pull requests.

Artifacts:

- Updated `scripts/quality.sh`.
- Updated `.github/workflows/ci.yml`.
- CI artifact names and paths documented in SPEC implementation notes or docs.

Phase acceptance:

- Local and CI quality criteria are equivalent.
- CI leaves all required artifacts.
- Any schema, report, eval, coverage, staticcheck, govulncheck, race, vet, gofmt, or unit test failure blocks CI.

### Phase 4: Internal Package Boundary Refactor

Implementation scope:

- Split only behavior-preserving internal boundaries from `internal/namba` when it reduces risk or clarifies ownership.
- Candidate package boundaries:
  - evidence contract
  - eval engine
  - report and observability
  - hook runtime
  - queue and runtime state
  - GitHub handoff
  - release workflow
  - CLI rendering and shared utilities
- Add an import boundary test or architecture check that prevents high-level packages from depending on CLI command glue or unrelated command implementations.
- Preserve public CLI behavior, command output contracts, JSON contracts, generated docs, and existing tests.
- Keep one integration layer that wires commands to the internal packages without duplicating logic.

Exclusions:

- No public package API promise outside this repository.
- No broad rewrite of command parsing.
- No behavior changes to release, security scan, installer checksum, or public workflow semantics.

Risks:

- Moving code can hide semantic drift behind package renames.
- Tests may overfit import paths instead of behavior.

Validation:

- Existing command tests remain green.
- Golden or fixture tests compare eval and report JSON before and after refactor.
- Architecture check fails on forbidden imports.
- Coverage does not regress below threshold.

Artifacts:

- New internal package directories only where justified.
- Import boundary or architecture test.
- Updated code references in docs or codemaps through `namba sync`.

Phase acceptance:

- Refactor is behavior-preserving.
- CLI output and JSON contracts remain compatible.
- Architecture check protects the intended package direction.

### Phase 5: README, Guide, And Project Docs Sync Validation

Implementation scope:

- Update docs source of truth rather than generated Markdown where possible:
  - README renderer in `internal/namba/readme.go`
  - docs config under `.namba/config/sections`
  - project docs generation inputs if needed
- Regenerate generated assets with `namba sync`.
- Run the docs drift gate after sync:
  - locally, run `namba sync`, include the expected generated outputs in the branch, then rerun `namba sync` and confirm it produces no additional changes
  - in CI, run `namba sync` from the committed PR state and fail if `git diff --exit-code` reports generated drift
- Ensure directly relevant docs describe:
  - schema contracts and compatibility policy
  - eval scorecard metrics and failure conditions
  - local and CI gate parity
  - CI artifacts
  - module boundary policy
  - docs sync expectations for generated outputs
- Verify README bundles and workflow guides stay in sync with the actual gate.

Exclusions:

- Do not rewrite unrelated marketing, release, or onboarding content.
- Do not manually edit generated files as the only durable change when renderer or config changes are required.

Risks:

- Generated docs can drift if only one language bundle is updated.
- Documentation can promise stricter gates than CI actually enforces.

Validation:

- `namba sync`
- Go tests covering README or docs rendering if the renderer changes.
- Docs drift check confirms `namba sync` is idempotent from the committed or staged generated state.

Artifacts:

- Updated README renderer or docs source.
- Regenerated README and workflow guide outputs.
- Refreshed `.namba/project` docs where `namba sync` updates them.

Phase acceptance:

- Docs name the same quality gate that CI and `scripts/quality.sh` enforce.
- Generated docs are reproducible through `namba sync`.
- No directly related generated doc remains stale.

## 1.0 Gate

SPEC-062 is complete only when all of these are true:

- JSON Schema compatibility tests pass for all listed evidence and report contracts.
- Harness eval regression and baseline comparison pass.
- Report JSON validation passes.
- Local quality gate and CI gate are equivalent or explicitly mapped.
- CI publishes coverage, eval JSON, eval Markdown, report JSON, and schema validation artifacts.
- Coverage threshold, staticcheck, govulncheck, gofmt, go vet, Go unit tests, and race tests are enforced.
- Architecture or import boundary check passes.
- README, workflow guide, and project docs are updated through source changes and sync.

## Validation Commands

Minimum implementation validation:

```sh
go test ./...
go vet ./...
gofmt -l cmd internal namba_test.go
go run ./cmd/namba eval --suite harness --format json --baseline internal/namba/testdata/evals/harness/baseline.json --fail-on-regression
go run ./cmd/namba report --format json
```

Full validation:

```sh
scripts/quality.sh
namba sync
git diff --check
```
