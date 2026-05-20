# SPEC-056 Harness Map

## Layer Map

| Layer | Owned Surface | SPEC-056 Impact |
| --- | --- | --- |
| Core harness | Namba CLI routing, harness request metadata, review readiness, deterministic guardrails, eval fixtures | Adds public `namba eval` and regression scoring for harness quality |
| Runtime harness | execution request, preflight, evidence manifest, readiness diagnostics | Reuses existing deterministic readiness and evidence helpers without live Codex |
| Domain harness | future user or domain-specific harness suites | Out of scope except the schema should allow future suites |
| Direct artifacts | repo-local skill or agent creation through `$namba-create` | Evaluated as fixture outcomes only; direct creation behavior is unchanged |
| CI | GitHub Actions running local commands | Adds deterministic `namba eval` regression check |

## Adaptation Rules

- This is `core_harness_change` because it changes Namba-owned CLI and eval contract surfaces.
- The implementation may adapt existing fixture families, but it must not make them depend on live services.
- Existing tests can migrate to the shared eval engine, but existing public command behavior must remain stable.
- Future domain suites must opt into their own fixtures and baselines rather than overloading harness v1 semantics.

## Boundary With Direct Generation

Direct artifact generation is an eval scenario category, not a new execution path in this SPEC. `$namba-create` remains responsible for direct skill or agent creation. `namba eval` only reports whether Namba classifies and guards such requests correctly.

## Boundary With Runtime Execution

`namba eval` may inspect deterministic readiness and evidence builders, but it must not invoke `namba run`, Codex, browser automation, or external APIs.
