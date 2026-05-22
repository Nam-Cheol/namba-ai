# SPEC-062 Eval Plan

## Eval Purpose

Quantify NambaAI harness quality as a deterministic local scorecard suitable for CI and operator review.

## Required Metrics

| Metric | What it measures |
| --- | --- |
| `route_selection` | Correct routing between direct artifact creation, ordinary plan, harness plan, core harness plan, and planned fix |
| `clarification_quality` | Vague requests require clarification while structured Goal, Scope, Constraints, Acceptance prompts proceed |
| `spec_completeness` | SPEC packages contain required files, phase details, acceptance, and required harness evidence sidecars |
| `execution_readiness` | Scenarios that should run are marked ready and blocked or ambiguous scenarios are not |
| `review_readiness` | Product, engineering, and design review readiness is represented without silently becoming a hard gate |
| `dangerous_command_blocking` | Destructive commands are denied and risky approval requests get explanatory risk notes |
| `evidence_completeness` | Run, queue, diagnostics, eval, baseline, and report evidence required by the workflow is present and valid |
| `schema_validity` | JSON evidence and report outputs validate against versioned schemas |
| `offline_e2e_workflow_health` | A local workflow path can progress through plan, eval, report, validation, and docs sync without external services |

## Required Buckets

- direct artifact generation
- domain feature change
- core runtime or harness change
- security-sensitive change
- release-related change
- docs-only change
- ambiguous clarification
- unsafe or blocked command
- missing evidence
- review-required
- schema-validity
- offline-e2e-workflow-health

## Output Contract

JSON scorecard fields:

- schema version
- suite
- corpus version
- generated timestamp
- metrics
- required coverage buckets
- scenario results
- scenario fingerprints
- baseline comparison
- regressions
- schema validation status
- report validation status

Compatibility rule:

- Keep `namba-eval-results/v1` and its existing metric names backward compatible.
- Emit the fixed 1.0 metric names through `namba-eval-scorecard/v1` or an equivalent append-only scorecard section.
- The baseline comparison may continue to use existing scenario fingerprints and pass counts, but the scorecard must also fail if any mapped 1.0 metric regresses.

Markdown summary fields:

- suite and corpus version
- total pass and fail counts
- metric table
- failed scenarios first
- baseline regression summary
- schema and report validation summary
- next action when failing

## Regression Rules

Fail the eval gate when:

- any scenario fails
- a metric pass count regresses from baseline
- a required bucket is missing
- a scenario in the baseline disappears
- schema validation fails
- report JSON validation fails

## Offline E2E Health

The offline E2E scenario should use local fixtures and command helpers only. It must not call live Codex, GitHub API, browser automation, telemetry, or LLM judging.
