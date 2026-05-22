# Review Readiness

SPEC: SPEC-066

Advisory only: use this summary to decide whether the SPEC has enough pre-implementation review depth before `namba run` or `namba pr`. Missing reviews should be visible, not silently blocking.

## Review Tracks

- Product Review
  Status: cleared
  Last Reviewed: 2026-05-22
  Reviewer: Codex product review
  Skill: `$namba-plan-pm-review`
  Artifact: `.namba/specs/SPEC-066/reviews/product.md`
- Engineering Review
  Status: cleared
  Last Reviewed: 2026-05-22
  Reviewer: Codex engineering review
  Skill: `$namba-plan-eng-review`
  Artifact: `.namba/specs/SPEC-066/reviews/engineering.md`
- Design Review
  Status: cleared
  Last Reviewed: 2026-05-22
  Reviewer: Codex design review
  Skill: `$namba-plan-design-review`
  Artifact: `.namba/specs/SPEC-066/reviews/design.md`

## Generated Instruction Contract Gate

- Classification source: `.namba/specs/SPEC-066/frontend-brief.md`
- Task Classification: `core-scaffold-generator-contract`
- Classification Rationale: Namba core managed generator/source-of-truth improvement for init/regen generated instruction surfaces.
- Gate Status: `not-applicable-to-visual-product-surface`
- Evidence Status: `sufficient-for-pre-implementation`
- Product interpretation: future `namba init` users receive clearer generated skill/agent guidance; non-goals and downstream backfill exclusions are explicit.
- Engineering interpretation: generator source changes, init/regen parity, manifest ownership, golden/snapshot/unit/integration tests, and validation commands are required.
- Design interpretation: This SPEC does not implement a frontend surface; design review is instruction/contract design review.
- Security advisory: read-only versus mutating boundary, secrets prohibition, destructive command/escalation policy, and non-project-specific guidance are required generated-contract concerns.
- Frontend-major correction: Frontend-major gates are not applicable unless actual UI/frontend files are introduced.

## Summary

- Cleared reviews: 3/3
- Advisory status: cleared for `namba run SPEC-066` when the user is ready to implement.

## Suggested Order

1. Implement generator/source-of-truth changes first.
2. Prove fresh temporary repo `namba init` and existing repo `namba regen` parity.
3. Run `gofmt`, `go test ./...`, and `go vet ./...`, reporting exact cause and impact for any failure.
