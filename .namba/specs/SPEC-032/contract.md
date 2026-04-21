# SPEC-032 Harness Contract

## Purpose

Define one typed contract that lets NambaAI classify harness-related requests consistently across `namba plan`, `namba harness`, and `$namba-create`.

The code and tests will remain authoritative. This document is the human-readable contract anchor.

## Canonical Classes

| `request_kind` | Meaning | Default entrypoint | Escalation rule |
| --- | --- | --- | --- |
| `core_harness_change` | Changes Namba-owned planning, routing, review, validator, runtime, or built-in harness surfaces | `namba plan` | Never downgrade to `$namba-create` while core contract ambiguity remains |
| `domain_harness_change` | Adds or adapts a reusable domain harness on top of the current Namba core contract | `namba harness` | Escalate to `core_harness_change` if `touches_namba_core=true` becomes true |
| `direct_artifact_creation` | The primary output is the artifact itself, not a reviewable harness SPEC | `$namba-create` | Escalate to `namba plan` if the artifact request would redefine the core harness contract |

## Minimal Metadata

| Field | Type | Allowed values or meaning | Why it exists |
| --- | --- | --- | --- |
| `request_kind` | enum | `core_harness_change`, `domain_harness_change`, `direct_artifact_creation` | Primary classifier outcome |
| `delivery_mode` | enum | `spec`, `direct` | Prevents SPEC planning and direct generation from being conflated |
| `adaptation_mode` | enum | `modify_core`, `extend_domain`, `compose_domain`, `generate_artifact` | Differentiates core modification from domain extension |
| `base_contract_ref` | string or null | SPEC id, contract id, or domain harness id being adapted | Makes adaptation explicit instead of implied |
| `touches_namba_core` | bool | `true` when Namba-owned surfaces change | Hard guardrail for escalation |
| `artifact_targets` | list | normalized outputs such as `skill`, `agent`, `workflow`, `validator`, `eval-pack`, `docs` | Separates harness intent from generated artifact shape |
| `required_evidence` | list | `contract`, `baseline`, `eval-plan`, optional `harness-map` | Makes review and validation material discoverable |
| `required_reviews` | list | `product`, `engineering`, `design` in v1 | Makes review expectations explicit without outrunning the current review runtime |

## Transport

- For SPEC routes, persist the typed object at `.namba/specs/<SPEC>/harness-request.json`.
- For direct routes, transport the same object transiently through the existing JSON-based `$namba-create` preview/apply flow.
- V1 intentionally does not introduce nested YAML parsing for this contract.
- V1 intentionally does not fabricate a SPEC package for direct creation just to persist the model.

## Operator Cheat Sheet

- Changing Namba's own harness, routing, validator, or built-in workflow contract -> `namba plan`
- Adding or adapting a reusable domain harness on top of Namba -> `namba harness`
- Creating the skill or agent artifact directly, with no unresolved contract change -> `$namba-create`

## Routing Precedence

Use the safer higher-order route when classification is ambiguous:

1. `namba plan`
2. `namba harness`
3. `$namba-create`

That precedence means direct generation never wins over unresolved harness-contract change.

## Decision Rules

1. If the request changes Namba-managed core surfaces such as `internal/namba/*`, built-in `namba-*` skills or agents, generated workflow docs, validator contracts, or review/runtime behavior, classify as `core_harness_change`.
2. Else, if the request introduces or adapts a reusable harness for a user or domain while preserving Namba core semantics, classify as `domain_harness_change`.
3. Else, if the user's goal is direct creation of a repo-local skill, agent, or paired artifact and the target is previewable without planning a new harness contract, classify as `direct_artifact_creation`.
4. If a request mixes domain harness work with direct artifact generation, keep the harness classification and treat direct generation as a later execution step.
5. If a request starts as `domain_harness_change` but later requires edits to Namba-owned core surfaces, reclassify it as `core_harness_change`.
6. If the user cannot tell whether they need a new contract or only an artifact, prefer the higher-order planning route first instead of guessing direct creation.

## Canonical Examples

| Request | Expected class | Route |
| --- | --- | --- |
| Change built-in `namba-harness` routing semantics | `core_harness_change` | `namba plan` |
| Add a reusable domain harness for a new industry workflow | `domain_harness_change` | `namba harness` |
| Create one repo-local skill and one custom agent directly | `direct_artifact_creation` | `$namba-create` |
| Create a new artifact and also change core route semantics | `core_harness_change` | `namba plan` first |
| Adapt an existing domain harness to compose multiple domain workflows | `domain_harness_change` | `namba harness` |
| Ambiguous "I want a skill and agent, but maybe this should become reusable harness behavior" | unresolved | prefer `namba plan` until the boundary is explicit |

## Evidence Profiles

| `request_kind` | Required evidence | Optional evidence |
| --- | --- | --- |
| `core_harness_change` | `contract.md`, `baseline.md`, `eval-plan.md` | `harness-map.md` when the change adapts an existing harness contract, or when the route boundary with direct artifact creation needs explicit documentation |
| `domain_harness_change` | `contract.md`, `baseline.md`, `eval-plan.md` | `harness-map.md` only when the domain harness adapts or composes an existing harness, or when the route boundary with direct artifact creation needs explicit documentation |
| `direct_artifact_creation` | preview/apply evidence from `$namba-create` | no harness evidence pack unless planning is reintroduced |

## Compatibility And Validator Hook

- Existing SPECs without `harness-request.json` stay on the legacy readiness path.
- Earlier evidence such as `extraction-map.md` remains valid for earlier SPECs; this contract does not retroactively rewrite history.
- New evidence requirements apply only to harness-classified slices that opt into this typed contract.
- V1 enforcement lives in a narrow harness-evidence validator helper invoked by readiness refresh and `namba sync` for harness-classified SPECs.
- That helper remains advisory in readiness output in v1; it should not redefine `namba run` execution validation.

## Invariants

- Keep `.namba/specs/<SPEC>/` as the planning artifact model.
- Preserve the existing planning-start and review-readiness contracts.
- Do not redesign `namba run` mode semantics in order to land this contract.
- Do not let direct generation become a silent backdoor for core harness mutation.
