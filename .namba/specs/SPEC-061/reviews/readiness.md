# Review Readiness

SPEC: SPEC-061

Advisory only: use this summary to decide whether the SPEC has enough pre-implementation review depth before `namba run` or `namba pr`. Missing reviews should be visible, not silently blocking.

## Review Tracks

- Product Review
  Status: clear with follow-ups
  Last Reviewed: 2026-05-21
  Reviewer: Codex as `namba-product-manager`
  Skill: `$namba-plan-pm-review`
  Artifact: `.namba/specs/SPEC-061/reviews/product.md`
- Engineering Review
  Status: approved
  Last Reviewed: 2026-05-21
  Reviewer: namba-planner
  Skill: `$namba-plan-eng-review`
  Artifact: `.namba/specs/SPEC-061/reviews/engineering.md`
- Design Review
  Status: clear
  Last Reviewed: 2026-05-21
  Reviewer: namba-designer
  Skill: `$namba-plan-design-review`
  Artifact: `.namba/specs/SPEC-061/reviews/design.md`

## Summary

- Cleared reviews: 2/3
- Advisory status: follow up on product=clear with follow-ups before execution or GitHub handoff if the risk profile justifies it.

## Suggested Order

1. Run product review when the user/problem framing or scope is still moving.
2. Run engineering review before implementation starts on anything with architecture or failure-mode risk.
3. Run design review when UX, interaction quality, or visual direction matters to acceptance.
