# SPEC-044 Plan

1. Refresh project context with `namba project` and verify the current `SPEC-036` frontend gate behavior in `frontend_brief.go`, `spec_review.go`, execution prompt materialization, generated templates, and regression tests.
2. Add failing tests that prove `frontend-major` can no longer pass with only complete five-gate labels and approved design review when the negative-first contract is missing, pending, insufficient, or violated.
3. Define the Do-Not Design Contract schema for `frontend-brief.md`, including default anti-pattern library, context-specific banned patterns, allowed replacements, brand/category/trust reasoning, visual grammar, most-generic-section redesign proof, architecture handoff, and violation-check expectations.
4. Extend `frontend-major` scaffold generation for `namba plan`, `namba harness`, and `namba fix --command plan` while keeping `frontend-minor`, non-frontend, and legacy no-brief behavior intentionally lightweight.
5. Extend frontend-brief parsing, readiness rendering, and run-time pre-execution blocking so incomplete or contradictory negative-first contract state blocks architecture/implementation even when the original five gates are complete.
6. Extend design-review scaffolds, role guidance, command skills, README/workflow docs, and generated agent mirrors so `namba-designer`, `namba-frontend-architect`, and `namba-frontend-implementer` share the same negative-first handoff contract.
7. Add post-implementation violation-check handling to the execution prompt/result path and make failed or unresolved banned-pattern checks visible in run output, sync artifacts, and PR support.
8. Evaluate deterministic helper-script candidates only if implementation needs a standalone checker; any helper must be read-only by default, support `--help`, keep bounded output, avoid network/auth assumptions, and have fixture or local-server tests.
9. Run product, engineering, and design review passes under `.namba/specs/SPEC-044/reviews/` when the SPEC is ready for implementation and refresh `.namba/specs/SPEC-044/reviews/readiness.md`.
10. Run validation commands, regenerate managed assets with `namba regen` when templates change, and run `namba sync` before PR handoff.
