# Product Review

- Status: clear with follow-ups
- Last Reviewed: 2026-05-21
- Reviewer: Codex as `namba-product-manager`
- Command Skill: `$namba-plan-pm-review`
- Recommended Role: `namba-product-manager`

## Focus

- Review whether this bugfix stays tightly scoped to queue continuity for real operators, preserves user trust around dirty-state detection, and sets a realistic fallback bar for projects where remote PR handoff is unavailable.

## Findings

- The core user value is clear and appropriately urgent for a fix SPEC: queue should not stop midway through a requested SPEC list because of platform-specific transport limits or because queue-owned bookkeeping dirties the worktree. That is a meaningful operator reliability issue, not internal cleanup.
- Scope is broad but still defensible because every listed area maps back to the same queue continuity outcome. The spec does not read like a general queue redesign; it stays centered on conveyor progress, cleanliness checks, exec transport, and fallback completion paths.
- `Medium` The fallback acceptance needs careful interpretation during implementation. "Queue can continue through a documented local branch creation and local main merge fallback" is product-correct, but it does not yet distinguish what must be automatic versus what may require explicit operator-visible handoff. The implementation should not silently downgrade from remote PR flow to a partially manual path without clear surfaced status.
- `Medium` The dirty-check exception boundary is the main trust risk. The spec correctly says queue-owned metadata and run evidence must not count as user dirtiness, but product value depends on this exemption staying narrow and explainable. If implementation broadens the ignore surface too far, users can lose confidence that real local edits will still stop the queue.
- `Low` Acceptance proves continuity and transport correctness, but it does not explicitly call for operator-readable per-SPEC outcomes such as completed, blocked, or fallback-used. That visibility is important if the queue continues across multiple projects and one item needs human follow-up.

## Decisions

- Product review supports implementation without reopening the SPEC.
- This should remain a bugfix slice. Do not expand it into queue UX redesign, queue state schema churn, or generalized dirty-file policy changes unless testing proves a migration is strictly necessary.
- Remote-PR-unavailable fallback is in scope because it protects end-to-end continuity, but the product bar is continuity with explicit visibility, not silent best-effort behavior.

## Follow-ups

- During implementation and validation, keep the dirty-check exemption limited to queue-owned writes for the active run and verify non-queue user edits still fail clean checks predictably.
- Ensure fallback reporting is explicit in logs or summary output so an operator can tell when queue processing stayed fully automated versus when it relied on local branch/local main fallback.
- If e2e coverage can do so without bloating the fix, include evidence that multi-SPEC runs surface which SPEC completed, which used fallback, and which stopped on a genuine blocker.

## Recommendation

- Proceed with implementation as advisory-clear.
- The SPEC is product-justified and sufficiently scoped for a bugfix, with the main caution that continuity must not come at the cost of masking real user changes or obscuring when the queue dropped to a local-only fallback path.
