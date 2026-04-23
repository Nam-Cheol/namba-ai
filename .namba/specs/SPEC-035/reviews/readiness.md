# Review Readiness

SPEC: SPEC-035

Advisory only: use this summary to decide whether the SPEC has enough pre-implementation review depth before `namba run` or `namba pr`. Missing reviews should be visible, not silently blocking.

## Review Tracks

- Product Review
  Status: approved
  Last Reviewed: 2026-04-23
  Reviewer: namba-product-manager
  Skill: `$namba-plan-pm-review`
  Artifact: `.namba/specs/SPEC-035/reviews/product.md`
- Engineering Review
  Status: clear
  Last Reviewed: 2026-04-23
  Reviewer: namba-planner
  Skill: `$namba-plan-eng-review`
  Artifact: `.namba/specs/SPEC-035/reviews/engineering.md`
- Design Review
  Status: advisory
  Last Reviewed: 2026-04-23
  Reviewer: namba-designer
  Skill: `$namba-plan-design-review`
  Artifact: `.namba/specs/SPEC-035/reviews/design.md`

## Summary

- Cleared reviews: 2/3
- Advisory status: follow up on design=advisory before execution or GitHub handoff if the risk profile justifies it.

## Suggested Order

1. Run product review when the user/problem framing or scope is still moving.
2. Run engineering review before implementation starts on anything with architecture or failure-mode risk.
3. Run design review when UX, interaction quality, or visual direction matters to acceptance.
