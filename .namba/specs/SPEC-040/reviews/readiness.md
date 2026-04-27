# Review Readiness

SPEC: SPEC-040

Advisory only: use this summary to decide whether the SPEC has enough pre-implementation review depth before `namba run` or `namba pr`. Missing reviews should be visible, not silently blocking.

## Review Tracks

- Product Review
  Status: clear
  Last Reviewed: 2026-04-27
  Reviewer: Codex acting as `namba-product-manager`
  Skill: `$namba-plan-pm-review`
  Artifact: `.namba/specs/SPEC-040/reviews/product.md`
- Engineering Review
  Status: clear
  Last Reviewed: 2026-04-27
  Reviewer: Codex acting as `namba-planner`
  Skill: `$namba-plan-eng-review`
  Artifact: `.namba/specs/SPEC-040/reviews/engineering.md`
- Design Review
  Status: clear
  Last Reviewed: 2026-04-27
  Reviewer: Codex acting as `namba-designer`
  Skill: `$namba-plan-design-review`
  Artifact: `.namba/specs/SPEC-040/reviews/design.md`

## Summary

- Cleared reviews: 3/3
- Advisory status: all current review tracks are marked clear.

## Harness Advisory

- Route: `namba plan`
- Request kind: `core_harness_change`
- Delivery mode: `spec`
- Adaptation mode: `modify_core`
- Base contract ref: `namba-core-harness`
- Touches Namba core: `true`
- Required evidence: `contract, baseline, eval-plan, harness-map`
- Evidence status: complete
- Required reviews: `product, engineering, design`
- Review artifact status: complete

## Phase-1 Evidence

- Runtime contract anchor: `.namba/specs/SPEC-040/contract.md`
- Baseline evidence: `.namba/specs/SPEC-040/baseline.md`
- Harness request: `.namba/specs/SPEC-040/harness-request.json`
- Eval plan: `.namba/specs/SPEC-040/eval-plan.md`
- Harness map: `.namba/specs/SPEC-040/harness-map.md`

## Suggested Order

1. Run product review when the user/problem framing or scope is still moving.
2. Run engineering review before implementation starts on anything with architecture or failure-mode risk.
3. Run design review when UX, interaction quality, or visual direction matters to acceptance.
