# Review Readiness

SPEC: SPEC-050

Advisory only: use this summary to decide whether the SPEC has enough pre-implementation review depth before `namba run` or `namba pr`. Missing reviews should be visible, not silently blocking.

## Review Tracks

- Product Review
  Status: needs-revision
  Last Reviewed: 2026-05-20
  Reviewer: Namba Product Manager
  Skill: `$namba-plan-pm-review`
  Artifact: `.namba/specs/SPEC-050/reviews/product.md`
- Engineering Review
  Status: needs-revision
  Last Reviewed: 2026-05-20
  Reviewer: namba-planner
  Skill: `$namba-plan-eng-review`
  Artifact: `.namba/specs/SPEC-050/reviews/engineering.md`
- Design Review
  Status: completed
  Last Reviewed: 2026-05-20
  Reviewer: Codex (`namba-designer` perspective)
  Skill: `$namba-plan-design-review`
  Artifact: `.namba/specs/SPEC-050/reviews/design.md`

## Summary

- Cleared reviews: 0/3
- Advisory status: follow up on product=needs-revision, engineering=needs-revision, design=completed before execution or GitHub handoff if the risk profile justifies it.

## Suggested Order

1. Run product review when the user/problem framing or scope is still moving.
2. Run engineering review before implementation starts on anything with architecture or failure-mode risk.
3. Run design review when UX, interaction quality, or visual direction matters to acceptance.
