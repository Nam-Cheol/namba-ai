# Design Review

- Status: clear
- Last Reviewed: 2026-05-01
- Reviewer: `namba-designer` via Codex
- Command Skill: `$namba-plan-design-review`
- Recommended Role: `namba-designer`

## Focus

Clarify information hierarchy, terminology, scanability, and operator workflow fit before implementation starts.

- Evidence Status: complete for documentation/operator-guidance scope
- Gate Decision: clear
- Approved Direction: docs, generated guidance, and CLI help wording only; no product UI surface is in scope
- Banned Patterns: release-note prose without a decision-first summary, ambiguous "update" wording, conflating Codex subagents with Namba worktrees, overexplaining non-adopted `/goal` workflows
- Open Questions: whether 5 worktree workers should be accepted as a separate implementation decision
- Unresolved Questions: none blocking first implementation slice
- Design Review Axes: hierarchy, terminology, scanability, workflow fit, non-UI applicability
- Keep / Fix / Quick Wins: keep the capability-based verdict up front; fix update/concurrency ambiguity; compress version deltas into a stable matrix or impact buckets

## Review Checklist

- Operator-facing docs should lead with the operating verdict: update is feasible and capability probing remains the contract.
- Version details should be tabled or grouped by operator impact so readers do not have to parse raw upstream chronology.
- `codex update` and `namba update` must be visually and semantically distinct anywhere commands are listed together.
- Same-workspace Codex agent threads and Namba git-worktree workers should never share one label such as "parallelism" without qualification.
- Permission-profile guidance should read as boundary-setting policy, not as a feature announcement.
- No frontend, palette, motion, or visual-composition work is required; this review is about documentation UX.

## Findings

- This SPEC is documentation UX, not application UI. Traditional visual review axes such as palette, motion, and component styling are largely non-applicable here.
- The strongest information architecture is: decision-first summary, version/impact matrix, local contract boundaries, implementation slices, acceptance checks.
- The primary design risk is operator misread, not visual inconsistency. The likely failure modes are ambiguous "update" semantics, blurred concurrency terminology, and excessive emphasis on upstream features Namba is not adopting yet.
- `/goal` belongs in a clearly marked future-facing note. If it appears beside required workflows without a label, readers may treat it as a new runtime dependency.
- The most generic section risk is any long release-summary block that mirrors upstream changelogs. That content should be redesigned into compact impact-grouped guidance rather than narrative recap.

## Decisions

- Treat this as an operator-guidance review, not a visual-design gate.
- Keep a compact matrix or grouped bullets for release deltas instead of prose-heavy release narration.
- Use explicit noun pairs in docs: `Codex subagent threads` versus `Namba worktree workers`.
- Keep guidance conservative, operational, and tied to accepted behavior only.
- Avoid productizing `/goal` until Namba deliberately adopts it in a separate workflow decision.

## Follow-ups

- During implementation, audit generated docs and help text for repeated but inconsistent wording across README, workflow guides, and `.namba/codex/README.md`.
- Where multiple commands appear together, visually separate upgrade commands from execution commands so operators can scan intent quickly.
- If a docs section still feels generic, redesign that section around "what changed for Namba operators" instead of "what changed upstream."

## Recommendation

Clear for implementation. No design-specific blocker remains, and the main requirement is disciplined documentation hierarchy rather than any UI treatment.
