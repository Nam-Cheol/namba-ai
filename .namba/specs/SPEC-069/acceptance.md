# Acceptance — SPEC-069

## Router and invariants

- [ ] Table-driven tests cover each phase's default, escalation, downgrade, fallback, and block decision for phase, role, risk flags, repair count, and Sol budget.
- [ ] `efficient`, `standard`, `deep` map immutably to explicit Luna, Terra, Sol IDs; no `gpt-5.6` alias is emitted.
- [ ] Luna implementation requires all six simple-work conditions. General work uses Terra. Sol implementations fail before execution.
- [ ] Sol medium/high budget, high predicate, no `xhigh/max`, and repeated-failure diagnostic condition are enforced for solo/fix/team/queue modes.
- [ ] The transition table fixes allowed phase edges and session semantics: `intake→plan`; `plan→design|architecture|implement`; `design|architecture→implement`; `implement→test`; `test→integration|repair`; `integration→review|repair`; `review→repair|complete`; `repair→implement|test|integration|review`. Same-model direct adjacent edges may UUID-resume; every branch, join, model change, missing UUID, and repair restart is a checkpointed fresh session. A Sol budget decrements exactly when a Sol decision turn is admitted, never for a blocked/fallback attempt.

## Execution, capability, and session safety

- [ ] Every exec and resume request encodes `-m <model>` and `-c model_reasoning_effort=<level>`; capability failures block before execution.
- [ ] Capability tests distinguish CLI syntax from exact model and effective-reasoning observation. The probe records its request, bounded outcome, and JSONL success signal; timeout, permission, model rejection, malformed/missing JSONL, and unobservable effective reasoning map to explicit status. Required Sol becomes `blocked_model_unavailable` only after this probe; an unobservable effective value remains `external_unobserved` rather than a guessed success.
- [ ] JSONL parser captures the first real thread UUID. Only explicit UUID + same model + adjacent phase resumes; no code path or invocation uses `--last`.
- [ ] Model boundary and UUID-missing/malformed cases open a fresh session with the checkpoint fields `goal`, `decisions`, `acceptance`, `affected_surfaces`, `constraints`, `open_risks`, `validation_state` and record `external_unobserved` where appropriate.
- [ ] Sol advisor and reviewer invocations are read-only with writer instructions absent; Terra/Luna writers receive the checkpoint. Direct fix, repair, queue, parallel, integrator, specialist, reviewer and managed-agent paths use the router.

## Config, migration, and generated artifacts

- [ ] New init and this repository expose only `model_routing_policy: gpt-5.6-cost-balanced-v1`; per-turn override is absent.
- [ ] Existing no-policy projects keep `legacy-static-v1` and doctor/status exposes migration state/action/conflicting keys. Adaptive plus global `model:` fails validation.
- [ ] Regen generates managed agents and `.namba/codex/model-routing.md`, preserves user agents/config, and passes twice without drift. Historical SPEC/release/run-log checksums do not change.
- [ ] Generated model-routing/operator documentation has snapshot or string-contract tests for policy mappings, absent per-turn override, `requested`/`effective`/`external_unobserved` meanings, planned dry-run state, and fallback/`blocked_model_unavailable` recovery command or condition.

## Evidence and operator UX

- [ ] `model-routing/v1` is additive. It records phase/tier, requested vs effective model/effort, rule/reasons, Sol state, fallback/state, UUID/session strategy, and observable usage; absent fields keep existing evidence/report consumers passing.
- [ ] dry-run outputs a stable ordered plan with phase, role, tier, selected model, effort, rule, reasons, session strategy, Sol remaining/max and recovery details for fallback/blocked states.
- [ ] report JSON aggregates per-model turns, fallback, and observable tokens; text output summarizes the same facts. Values not observed are `unavailable` or `external_unobserved`, never inferred.

## Regression, quality, and release

- [ ] Fixtures cover simple single-area Luna/no-Sol, ordinary Terra/no-Sol, cross-system Sol-medium then Terra, irreversible security Sol-high then Terra, and repeated failure Terra-high with Sol only under extra risk.
- [ ] Fixtures cover unavailable model, missing/malformed JSONL, missing usage, invalid policy combination, resume/checkpoint behavior, dry-run/report snapshots, and old evidence/report compatibility.
- [ ] `go test ./...`, gofmt check, `go vet ./...`, `namba eval`, `namba regen`, `namba sync`, and `scripts/quality.sh` pass.
- [ ] Before major release: legacy compatibility, adaptive new-init, conflict UX, migration/rollback docs, generated-artifact checks, scoped commit/push, clean-main release guard, and release notes are evidenced.
