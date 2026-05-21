# SPEC-056

## Problem

NambaAI already has deterministic harness-quality examples under `internal/namba/testdata/evals/harness/`, but they are locked inside Go test helpers and cannot be run as a public operator-facing evaluation command.

That leaves three gaps:

- Harness quality cannot be quantified from the CLI before or after changing planning, review, runtime, or guardrail behavior.
- Existing fixture cases are not represented by one stable scenario schema, result schema, or baseline comparison contract.
- CI catches many unit regressions through `go test ./...`, but it does not expose a focused "harness quality score changed" signal with scenario-level diagnostics.

The requested feature is a core Namba harness change because it adds public CLI behavior, evaluation fixtures, baseline comparison, JSON and Markdown reporting, and CI regression checks for Namba-owned planning and harness-runtime boundaries.

## Goal

Add a deterministic `namba eval` command that can quantify how well NambaAI structures ambiguous development requests into route, delivery, evidence, review, clarification, SPEC, and execution-readiness decisions.

## Context

- Project: namba-ai
- Project type: existing
- Language: go
- Mode: tdd
- Work type: plan

## Scope

- Add a public read-only CLI command: `namba eval`.
- Support at least the `harness` suite in v1.
- Define and load a deterministic eval scenario schema.
- Evaluate these dimensions:
  - route selection accuracy: core, domain, direct, ordinary, or planned fix classification
  - delivery mode
  - artifact target selection
  - required evidence selection
  - required review selection
  - clarification-required decision
  - SPEC required-field completeness
  - execution-readiness decision
- Produce human-readable Markdown and CI-parseable JSON.
- Compare current results against a checked-in baseline.
- Extend CI so regression is visible without live Codex, GitHub API, network calls, or LLM judging.

## Non-Goals

- Do not evaluate actual Codex model quality.
- Do not run live Codex, GitHub API, browser, SaaS dashboard, or telemetry flows.
- Do not change existing public behavior for `namba plan`, `namba harness`, `$namba-create`, `namba run`, `namba pr`, or `namba release`.
- Do not introduce an LLM judge or probabilistic scoring path.
- Do not require network access for fixture execution, baseline comparison, or CI.

## Proposed Design

### CLI Contract

`namba eval` is a read-only command with these v1 shapes:

```sh
namba eval
namba eval --suite harness
namba eval --fixture internal/namba/testdata/evals/harness/scenarios.json
namba eval --format markdown
namba eval --format json
namba eval --baseline internal/namba/testdata/evals/harness/baseline.json
namba eval --fail-on-regression
namba eval --update-baseline
namba eval --case route_core_eval_command
```

Defaults:

- suite: `harness`
- format: `markdown`
- fixture: embedded default harness scenario corpus
- baseline: checked-in default baseline when present

Exit behavior:

- `0`: all scenarios pass and no baseline regression is detected
- `1`: scenario failure or baseline regression
- `2`: invalid fixture, invalid baseline, unsupported suite, or unsupported format

### Fixture Location

Keep the source corpus near current tests:

- `internal/namba/testdata/evals/harness/scenarios.json`
- `internal/namba/testdata/evals/harness/baseline.json`
- `internal/namba/testdata/evals/harness/README.md`

Use `go:embed` for the default corpus and baseline so installed binaries can run `namba eval` without depending on a source checkout. Keep `--fixture` and `--baseline` for repo-local or future external corpora.

### Engine Shape

Move the reusable logic currently implied by `internal/namba/harness_eval_test.go` into production code:

- `internal/namba/eval_command.go`
- `internal/namba/eval_schema.go`
- `internal/namba/eval_engine.go`
- `internal/namba/eval_render.go`

Tests should call the same engine that the CLI calls. The engine must calculate actual outcomes from Namba code paths, not switch behavior based on expected fixture labels.

### Result Contract

The JSON output should include:

- `schema_version`
- `suite`
- `generated_at`
- `summary`
- `metrics`
- `scenarios`
- `baseline`
- `regressions`

Each scenario result should include:

- `id`
- `type`
- `tags`
- `passed`
- `input`
- `expected`
- `actual`
- `failures`
- `rationale`

Markdown output should show the same information in a concise report:

- summary counts
- metric scores
- failed scenarios first
- per-scenario expected versus actual fields
- baseline regression notes

### Baseline Contract

The baseline should be checked in as JSON and intentionally small:

- scenario ids and expected fingerprints
- aggregate metric pass counts
- required category coverage
- corpus version

Regression detection should fail when:

- a previously passing scenario fails
- a scenario disappears without updating the baseline
- a metric pass rate decreases
- a required category coverage bucket is missing
- fixture schema version or result schema version is incompatible

## Minimum Fixture Corpus

V1 should include at least 24 scenarios across these buckets:

- direct artifact generation: 3
- domain feature change: 3
- core runtime or harness change: 4
- security-sensitive change: 3
- release-related change: 2
- docs-only change: 2
- ambiguous request requiring clarification: 3
- unsafe or blocked command scenario: 2
- missing evidence scenario: 2
- review-required scenario: 2

Existing fixture families can be folded into the unified corpus:

- route cases
- mention or plugin ambiguity cases
- prompt-refinement cases
- guardrail cases
- evidence manifest cases
- PR review opt-in cases

## Implementation Notes

- Keep eval execution deterministic and local.
- Avoid wall-clock-sensitive assertions by allowing injected timestamps or stable normalization in test mode.
- Prefer existing helpers where they are already production code:
  - `evaluateSpecCreationClarification`
  - `inferredPlanningHarnessRequest`
  - `harnessRouteForRequest`
  - `validateHarnessEvidence`
  - `parsePRArgs`
  - execution evidence manifest builders
- Replace test-only wrappers that make route decisions from `expected_category`.
- Preserve current command usage and help contracts.

## Validation

Minimum validation after implementation:

```sh
go test ./...
```

Recommended validation:

```sh
go test ./...
go vet ./...
gofmt -l cmd internal namba_test.go
go run ./cmd/namba eval --suite harness --format json --baseline internal/namba/testdata/evals/harness/baseline.json --fail-on-regression
```
