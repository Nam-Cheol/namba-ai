# Product Review

- Status: clear
- Last Reviewed: 2026-04-08
- Reviewer: codex
- Command Skill: `$namba-plan-pm-review`
- Recommended Role: `namba-product-manager`

## Focus

- Challenge the problem framing, scope, user value, and acceptance bar before implementation starts.

## Findings

- The problem framing is strong: users do not merely need "another template." They need a builder that can keep clarifying intent until the target artifact is unambiguous and then generate the correct surface.
- Treating `namba-create` as a skill-first entrypoint matches how Namba already exposes user-facing workflows today. It avoids forcing a larger CLI redesign into the first delivery.
- The explicit precedence rule for `skill`, `agent`, or `both` is important product behavior because it keeps the user in control even when the classifier has a different guess.
- The regen-ownership problem is part of user value, not only implementation detail. If generated assets disappear after `namba regen`, the feature will feel broken regardless of how good the interaction loop is.
- Raising the repo-managed same-workspace agent thread default from `3` to `5` makes the default collaboration posture match the stated five-role requirement more honestly.
- That increase should still be framed as same-workspace agent concurrency only, not as a blanket parallelism change across worktree execution.

## Decisions

- Phase 1 should ship as `$namba-create` rather than as a new `namba create` CLI command.
- The workflow should be preview-first and confirmation-gated.
- Explicit user target selection outranks classifier inference.
- Raise the repo-managed same-workspace `max_threads` default from `3` to `5` for `agent_mode: multi`.
- Keep worktree parallel limits as a separate workflow control and do not silently tie them to the same value.

## Follow-ups

- During implementation, keep the route-selection guidance crisp so users can still tell when `$namba-plan` or `$namba-harness` is the better entrypoint.
- Confirm the user-authored-versus-managed ownership contract early, because that decision changes the perceived trustworthiness of the feature.
- Keep docs explicit that the `3 -> 5` change is about same-workspace agent capacity, not `namba run --parallel` worktree fan-out.

## Recommendation

- Clear to proceed from a product perspective. The scope is specific, user-facing value is real, and the acceptance bar is concrete enough to protect against a vague "AI builder" implementation.
