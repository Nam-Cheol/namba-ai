# SPEC-056 Eval Contract

## Purpose

Define the minimum deterministic contract for `namba eval` so harness quality can be measured from fixtures without live model, network, GitHub, or Codex dependencies.

## Command

`namba eval` is read-only. It must not create SPECs, modify project files, call Codex, call GitHub, start browsers, require network access, or depend on user-local credentials.

## Supported Suite

V1 supports `harness`.

Future suites may be added later, but unsupported suite names must fail with a schema or usage error instead of silently running a partial corpus.

## Scenario Types

V1 scenario types:

- `route_selection`
- `delivery_mode`
- `artifact_targets`
- `required_evidence`
- `required_reviews`
- `clarification`
- `spec_fields`
- `execution_readiness`
- `guardrail`
- `evidence_manifest`
- `pr_review`

One scenario can score more than one metric when the expected fields require it.

## Scenario Schema

```json
{
  "schema_version": "namba-eval-scenario/v1",
  "id": "route_core_eval_command",
  "suite": "harness",
  "type": "route_selection",
  "tags": ["core_harness_change", "review-required"],
  "input": {
    "invocation": "plan",
    "text": "Add deterministic namba eval for harness quality"
  },
  "expected": {
    "category": "core_harness_change",
    "command": "namba plan",
    "delivery_mode": "spec",
    "artifact_targets": ["workflow", "validator", "eval-pack", "docs"],
    "required_evidence": ["contract", "baseline", "eval-plan", "harness-map"],
    "required_reviews": ["product", "engineering", "design"],
    "clarification_required": false,
    "spec_required_fields": {
      "spec": true,
      "plan": true,
      "acceptance": true
    },
    "execution_ready": true
  },
  "rationale": "Core Namba eval behavior must route through reviewable SPEC planning."
}
```

## Result JSON Schema

```json
{
  "schema_version": "namba-eval-result/v1",
  "suite": "harness",
  "generated_at": "2026-05-20T00:00:00Z",
  "summary": {
    "total": 24,
    "passed": 24,
    "failed": 0,
    "regressions": 0
  },
  "metrics": [
    {
      "name": "route_selection",
      "total": 6,
      "passed": 6,
      "failed": 0
    }
  ],
  "scenarios": [
    {
      "id": "route_core_eval_command",
      "type": "route_selection",
      "tags": ["core_harness_change"],
      "passed": true,
      "input": {},
      "expected": {},
      "actual": {},
      "failures": [],
      "rationale": "..."
    }
  ],
  "baseline": {
    "path": "internal/namba/testdata/evals/harness/baseline.json",
    "matched": true
  },
  "regressions": []
}
```

## Baseline Schema

The baseline is an approved-result lockfile, not a second source of expected truth. It should include:

- `schema_version`
- `suite`
- `corpus_version`
- `scenario_ids`
- `metric_scores`
- `required_coverage`
- `approved_result_fingerprint`

The fixture owns per-scenario expected values. The baseline owns regression expectations over the suite.

## Invariants

- Actual outcomes must be computed from Namba code paths, not selected from expected fixture categories.
- Fixture execution must be deterministic.
- JSON output must be machine-parseable without reading Markdown.
- Markdown output must be enough for a maintainer to diagnose why a scenario failed.
- Existing public commands must keep their behavior.
