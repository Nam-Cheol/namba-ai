# Product Review

- Status: clear
- Last Reviewed: 2026-05-05
- Reviewer: namba-product-manager
- Command Skill: `$namba-plan-pm-review`
- Recommended Role: `namba-product-manager`

## Focus

- Challenge the problem framing, scope, user value, and acceptance bar before implementation starts.

## Findings

- The problem framing is strong: operators can already create many SPECs, but the current workflow makes them manually repeat review, implementation, PR, checks, land, and main refresh for every package.
- The one-active-SPEC constraint is the right product safety line. It keeps the queue from becoming cross-SPEC batch automation that hides failures.
- The revised `waiting_for_land` behavior prevents a surprising no-auto-land flow: without `--auto-land`, the queue waits on the current SPEC instead of starting the next one before the first SPEC is landed.
- The queue-scoped `--skip-codex-review` language is now narrow enough: it means skipping creation of a new `@codex review` marker comment, not skipping plan review, GitHub review, or evidence collection.
- Review pass criteria are now explicit enough for operators to understand why a queue continues or blocks.

## Decisions

- V1 terminal completion is `landed`; implemented-only and PR-ready states skip completed phases but do not skip the whole SPEC.
- `clear-with-followups` can continue only when follow-ups are tagged as `[non-blocking]` or `[post-implementation]`.
- Ambiguous state remains a product-visible `blocked` condition, not an internal retry or silent skip.

## Follow-ups

- [non-blocking] During implementation, keep the help text short enough that `namba queue --help` feels like an operator command, not a policy document.
- [post-implementation] After the first queue implementation lands, consider whether users need `--until implemented|pr|landed` as a separate follow-up SPEC.

## Recommendation

- Clear to proceed. Product risk is now bounded by explicit waiting, review, and skip semantics.
