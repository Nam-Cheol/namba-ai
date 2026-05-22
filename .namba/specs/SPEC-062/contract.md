# SPEC-062 Contract

## Contract Type

Core Namba harness contract upgrade. This SPEC stabilizes existing evidence, eval, report, CI, and package-boundary behavior instead of introducing a new user-facing workflow.

## Versioned JSON Contracts

Required schema contracts:

| Contract | Current schema version | Required validator coverage |
| --- | --- | --- |
| Execution evidence manifest | `execution-evidence/v1` | success, validation failure, missing artifact refs, hook results, runtime validation attempts |
| Queue runner evidence | `queue-runner-evidence/v1` | success, failed validation, fallback, stale runner, blocked queue state |
| Queue state runtime fixture | no schema version retrofit in SPEC-062 | parser compatibility for `.namba/logs/queue/state.json` as report input |
| Codex diagnostics evidence | `codex-diagnostics-evidence/v1` | detected, unavailable, failed doctor, redaction, root comparison |
| Project Codex diagnostics evidence | `project-codex-diagnostics-evidence/v1` | wrapper with project metadata and diagnostics payload |
| Eval result | `namba-eval-results/v1` | summary, metrics, scenarios, baseline, regressions |
| Eval baseline | `namba-eval-baseline/v1` | fingerprints, metric pass counts, required coverage buckets |
| Eval scorecard | `namba-eval-scorecard/v1` | fixed 1.0 metrics, coverage buckets, scenario fingerprints, schema and report validation |
| Report JSON | `namba-report/v1` | project, summary, runs, queue, specs, release, diagnostics, issues, warnings |

## Compatibility Rules

- Unknown optional fields are allowed.
- New optional fields are allowed without a major version bump.
- New required fields require an explicit compatibility decision and fixture update.
- Removing required fields is breaking.
- Changing the JSON type of an existing field is breaking.
- Changing semantic meaning of an existing enum or status value is breaking unless represented by a new value and covered by fixtures.
- Unsupported `schema_version` values fail validation.

## Eval Compatibility And Scorecard Contract

`namba-eval-results/v1` remains backward compatible. Existing metric names may remain in eval result JSON and baseline comparison. The 1.0 metric vocabulary is exposed by `namba-eval-scorecard/v1` or an equivalent append-only scorecard section derived from the result.

Metric mapping:

- `clarification` -> `clarification_quality`
- `spec_required_fields` -> `spec_completeness`
- `required_reviews` and PR review scenarios -> `review_readiness`
- guardrail deny and risk-note scenarios -> `dangerous_command_blocking`
- evidence manifest scenarios -> `evidence_completeness`
- schema validation scenarios -> `schema_validity`
- deterministic local workflow scenarios -> `offline_e2e_workflow_health`
- `route_selection` and `execution_readiness` keep their names

The 1.0 scorecard must contain these metric names:

- `route_selection`
- `clarification_quality`
- `spec_completeness`
- `execution_readiness`
- `review_readiness`
- `dangerous_command_blocking`
- `evidence_completeness`
- `schema_validity`
- `offline_e2e_workflow_health`

Required failure conditions:

- scenario failure
- metric regression from baseline
- required coverage bucket missing
- scenario disappearance without baseline update
- schema incompatibility

## CI Contract

The local and CI gate must enforce the same quality criteria:

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

Docs sync validation belongs to Phase 5. CI may provide the reusable drift check, but Phase 3 is not complete or incomplete based on docs content updates.

CI must publish:

- coverage report
- eval scorecard JSON
- eval summary Markdown
- report JSON
- schema validation result

## Refactor Contract

Package boundary changes are allowed only when they are behavior-preserving. Public CLI behavior, command output, JSON contracts, generated docs, and existing tests must remain compatible.
