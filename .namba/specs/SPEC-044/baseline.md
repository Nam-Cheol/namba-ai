# SPEC-044 Baseline

## Existing Frontend Gate

- `SPEC-036` created the current frontend synthesis gate.
- `internal/namba/frontend_brief.go` owns frontend task classification, `frontend-brief.md` generation, fixed-label parsing, evidence-state derivation, design-review comparison, readiness lines, and frontend run-block errors.
- `internal/namba/spec_review.go` owns the design-review scaffold and aggregate readiness rendering.
- `internal/namba/templates.go` owns generated Namba skills, custom agent mirrors, AGENTS guidance, README text, and workflow-guide text.
- `internal/namba/execution_test.go` already proves incomplete `frontend-major` synthesis blocks before runner dispatch.
- `internal/namba/frontend_brief_test.go` already covers multiline decision fields, pending design-review fields, status-only remediation, and classification boundaries.
- `internal/namba/spec_command_test.go` already covers frontend-brief scaffold generation and initial readiness rendering.
- `internal/namba/spec_review_test.go` already covers design-review scaffold fields and readiness output.
- `internal/namba/templates_test.go` already asserts generated frontend role guidance.

## Current Strengths

- `frontend-major` work has explicit fixed-label gate state.
- Missing, insufficient, invalid, or mismatched frontend evidence blocks execution.
- Design review must resolve approved direction, banned patterns, open questions, and unresolved questions before a major run can proceed.
- Readiness summaries expose frontend gate state alongside advisory product, engineering, and design review status.
- Role routing already distinguishes design synthesis, frontend architecture planning, and approved UI implementation.

## Current Gaps

- The gate focuses on positive evidence and can approve work without naming the generic fallback that must be avoided.
- The existing `Banned Patterns` field is too small to carry default-library reasoning, context-specific detection hints, allowed replacements, visual grammar, or exception paths.
- Design review can be marked approved without proving brand, category, and trust reasoning.
- Architecture guidance can still infer component and layout plans from generic primitives when the handoff is vague.
- The execution path blocks before implementation but does not yet fail after implementation when the output falls back to a banned generic UI pattern.
- Generated docs and role cards mention anti-generic composition, but they do not define a negative-first contract or violation-check protocol.

## Implementation-Risk Notes

- Adding fixed labels to the top of `frontend-brief.md` may make older frontend briefs invalid unless compatibility behavior is intentionally designed.
- A purely deterministic code scanner will produce false positives if it treats words like `card` or `grid` as automatic violations; the contract should require detection hints plus replacement and exception rules.
- A purely self-reported model check is too soft if it is not surfaced in run evidence and PR support artifacts.
- Template changes should be made in `internal/namba/templates.go` first, then regenerated with `namba regen`.

