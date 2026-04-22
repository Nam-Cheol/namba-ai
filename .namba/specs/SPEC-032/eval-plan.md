# SPEC-032 Eval Plan

## Goal

Define the minimum validator and regression-fixture shape that would prove the typed harness contract is working.

## Classification Fixture Matrix

| Fixture | Expected class | Expected route | Required evidence |
| --- | --- | --- | --- |
| Tighten Namba's core harness classifier and readiness validator | `core_harness_change` | `namba plan` | `contract`, `baseline`, `eval-plan` |
| Redefine built-in `namba-harness` routing semantics | `core_harness_change` | `namba plan` | `contract`, `baseline`, `eval-plan` |
| Add a brand-new reusable finance-analysis domain harness without touching built-in Namba routing | `domain_harness_change` | `namba harness` | `contract`, `baseline`, `eval-plan` |
| Adapt an existing domain harness to compose two reusable domain workflows | `domain_harness_change` | `namba harness` | `contract`, `baseline`, `eval-plan`, `harness-map` |
| Create a repo-local release-triage skill and paired agent directly | `direct_artifact_creation` | `$namba-create` | preview/apply evidence using the transient JSON transport |
| Create a repo-local artifact that also changes Namba-managed harness semantics | reject direct route, escalate to `core_harness_change` | `namba plan` | harness evidence pack |
| Refresh readiness for an older SPEC that has `contract.md` and `extraction-map.md` but no typed harness sidecar | legacy SPEC | keep current readiness behavior | legacy evidence remains accepted |

## Validator Expectations

- Reject harness-classified work when required metadata fields are missing.
- Reject harness-classified work when required evidence artifacts are missing.
- Reject `$namba-create` as the primary route when `touches_namba_core=true`.
- Reject `domain_harness_change` when the request edits Namba-owned core surfaces without escalation.
- Preserve current readiness behavior for SPECs that do not opt into the typed harness contract.

## Regression Coverage

- Table-driven classification tests for the three canonical classes plus at least one escalation case.
- Evidence-completeness tests for `contract.md`, `baseline.md`, `eval-plan.md`, and conditional `harness-map.md`.
- Transport-split tests proving `.namba/specs/<SPEC>/harness-request.json` is used only for SPEC routes while direct create continues through JSON preview/apply.
- Route-boundary tests proving `namba plan`, `namba harness`, and `$namba-create` remain non-overlapping in the covered fixtures.
- Compatibility tests proving the current review-readiness flow still works with the new evidence pack.
- Guard tests proving `namba run` mode semantics are unchanged by this slice.

## Exit Signal

This SPEC is ready to execute when the planned implementation can be tested with classification fixtures instead of relying on reviewer interpretation of prose alone.
