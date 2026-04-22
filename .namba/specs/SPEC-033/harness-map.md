# SPEC-033 Harness Map

## Purpose

Map the relationship between the existing planning/readiness contract and the new execution-evidence layer.

## Planning Layer

- `spec.md`
- `plan.md`
- `acceptance.md`
- `reviews/*.md`
- `reviews/readiness.md`
- `harness-request.json`
- `contract.md`
- `baseline.md`
- `eval-plan.md`
- `harness-map.md`

This layer answers: "Is the change well-classified, reviewable, and ready enough to implement?"

## Runtime Layer

- `.namba/logs/runs/<log-id>-request.json`
- `.namba/logs/runs/<log-id>-preflight.json`
- `.namba/logs/runs/<log-id>-execution.json`
- `.namba/logs/runs/<log-id>-validation.json`
- `.namba/logs/runs/<spec-id>-parallel.events.jsonl` when emitted
- `.namba/logs/runs/<log-id>-evidence.json`

This layer answers: "What runtime proof exists for the run that actually happened?"

## Bridge Rule

- Planning/readiness remains advisory and pre-implementation.
- Execution evidence remains advisory and post-run in v1.
- `reviews/readiness.md` stays plan-only.
- `namba sync` and sync-generated summaries are the primary v1 place to surface execution proof.
- Consumers may surface both layers together, but they must not pretend they are the same phase or the same status.

## Browser And Observability Position

- Browser artifacts are optional runtime evidence producers.
- Runtime observability artifacts already exist in structured form and should be referenced, not rediscovered by free-form scraping.
- The core contract is the typed bridge, not a blanket browser requirement.
