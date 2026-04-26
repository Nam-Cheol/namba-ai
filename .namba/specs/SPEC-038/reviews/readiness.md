# Review Readiness

SPEC: SPEC-038

Advisory only: use this summary to decide whether the SPEC has enough pre-implementation review depth before `namba run` or `namba pr`. Missing reviews should be visible, not silently blocking.

## Review Tracks

- Product Review
  Status: clear
  Last Reviewed: 2026-04-26
  Reviewer: Codex (`namba-product-manager`)
  Skill: `$namba-plan-pm-review`
  Artifact: `.namba/specs/SPEC-038/reviews/product.md`
- Engineering Review
  Status: clear
  Last Reviewed: 2026-04-26
  Reviewer: namba-planner
  Skill: `$namba-plan-eng-review`
  Artifact: `.namba/specs/SPEC-038/reviews/engineering.md`
- Design Review
  Status: clear
  Last Reviewed: 2026-04-26
  Reviewer: Codex
  Skill: `$namba-plan-design-review`
  Artifact: `.namba/specs/SPEC-038/reviews/design.md`

## Summary

- Cleared reviews: 3/3
- Advisory status: all current review tracks are marked clear.

## Harness Advisory

- Route: `namba plan`
- Request kind: `core_harness_change`
- Delivery mode: `spec`
- Adaptation mode: `extend_domain`
- Base contract ref: `$namba-help, $namba-create, $namba-plan, $namba-harness, $namba-run`
- Touches Namba core: `true`
- Required evidence: `contract, baseline, eval-plan`
- Evidence status: complete
- Required reviews: `product, engineering, design`
- Review artifact status: complete

## Phase-1 Evidence

- Runtime contract anchor: `.namba/specs/SPEC-038/contract.md`
- Baseline evidence: `.namba/specs/SPEC-038/baseline.md`
- Harness request: `.namba/specs/SPEC-038/harness-request.json`
- Eval plan: `.namba/specs/SPEC-038/eval-plan.md`

## Suggested Order

1. Run product review when the user/problem framing or scope is still moving.
2. Run engineering review before implementation starts on anything with architecture or failure-mode risk.
3. Run design review when UX, interaction quality, or visual direction matters to acceptance.
