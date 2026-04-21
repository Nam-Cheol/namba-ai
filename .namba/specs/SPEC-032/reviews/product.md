# Product Review

- Status: clear
- Last Reviewed: 2026-04-21
- Reviewer: namba-product-manager
- Command Skill: `$namba-plan-pm-review`
- Recommended Role: `namba-product-manager`

## Focus

- Challenge the problem framing, scope, user value, and acceptance bar before implementation starts.

## Findings

- Route-choice clarity is now sufficient to implement. The `Operator Cheat Sheet`, routing precedence, decision rules, and canonical examples turn the classifier into a usable command-selection contract instead of a reviewer-only internal model.
- The direct-flow product boundary is now clear enough for v1. SPEC routes persist `.namba/specs/<SPEC>/harness-request.json`, while direct creation reuses transient JSON through `$namba-create` preview/apply without fabricating a fake SPEC package.
- The conditional nature of `harness-map.md` is now aligned across the package well enough to proceed. It no longer reads as a blanket requirement for every domain harness request.
- The main UX safety rail is now explicit: when the user cannot tell whether they need a contract change or only an artifact, Namba should prefer the higher-order planning route instead of guessing direct creation.

## Decisions

- Accept the three-route product model for v1: `namba plan`, `namba harness`, and `$namba-create`.
- Accept the JSON transport split as the correct product boundary between reviewable SPEC flows and direct artifact flows.
- Accept `harness-map.md` as conditionally required rather than universally required.
- Treat route-choice clarity as a user-facing outcome, not only a classifier outcome.

## Follow-ups

- Normalize the final `harness-map.md` trigger into one exact validator sentence during implementation so operator-facing docs and validator output do not drift.
- Reuse the same canonical route examples in user-facing docs and skills rather than paraphrasing them differently.
- Make the direct-flow preview output visibly explain when a request has been escalated out of `$namba-create` into `namba plan`.

## Recommendation

- Recommend approval to proceed. The remaining work is editorial tightening around the conditional `harness-map.md` rule, not a product-framing blocker.
