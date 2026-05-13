# Product Review

- Status: advisory-approved
- Last Reviewed: 2026-05-13
- Reviewer: namba-product-manager
- Command Skill: `$namba-plan-pm-review`
- Recommended Role: `namba-product-manager`

## Focus

- Challenge the problem framing, scope, user value, and acceptance bar before implementation starts.

## Findings

- The core product problem is real and well-scoped: the current five-gate frontend evidence model can approve visually generic output, so adding a negative-first contract closes a quality gap in workflow governance rather than introducing subjective taste review.
- Scope discipline is mostly strong. The SPEC keeps the new burden on future `frontend-major` work, preserves advisory behavior for `frontend-minor` and non-frontend paths, and avoids introducing a second artifact model outside `.namba/specs/`.
- The main product risk is over-blocking from vague bans. The current contract shape correctly requires rationale, detection hints, allowed replacements, and exception paths; implementation should preserve that structure as a hard rule so teams are never blocked by "avoid generic UI" prose alone.
- Post-implementation violation checks are the highest-value addition, but they are also the place most likely to feel arbitrary. Product intent should remain: deterministic local checks when possible, explicit reviewer-visible evidence when judgment is required, and no dependence on networked design tools or image recognition for the v1 bar.
- Because `SPEC-044` is harness/workflow work rather than an end-user UI build, implementation must keep a clean boundary between this SPEC's own `frontend-minor` sidecar classification and the future `frontend-major` contracts it scaffolds and enforces.

## Decisions

- Proceed with implementation as an advisory-approved product direction.
- Treat the Do-Not Design Contract as a quality and trust mechanism, not as a novelty mandate. Valid use of cards, gradients, metrics, or other primitives must remain possible through explicit semantic justification and exception paths.
- Preserve rollout boundaries exactly as written: hard blocking applies to future `frontend-major` execution paths, while `frontend-minor`, non-frontend, and legacy no-brief work stay intentionally lightweight unless explicitly reclassified.

## Follow-ups

- Verify the generated scaffold includes at least one concrete example of a banned pattern paired with an allowed replacement and exception path so users understand the expected level of specificity.
- Verify run, sync, and PR-facing artifacts surface negative-first failures in a way reviewers can see immediately; a hidden failure in runner output alone is not a sufficient product outcome.
- Verify at least one test fixture covers a justified exception-path success case, not only failure cases, so the contract does not drift into blanket prohibition.

## Recommendation

- Recommend implementation. No product blocker remains, but the team should preserve specificity, visible evidence, and narrow rollout boundaries so the new gate improves trust without turning into subjective style policing.
