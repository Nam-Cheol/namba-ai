---
name: namba-plan-pm-review
description: Command-style entry point for product review of a SPEC before implementation starts.
---

Use this skill when the user explicitly says `$namba-plan-pm-review` or asks for a product review on a SPEC package.

Behavior:
- Resolve the target SPEC from an explicit `SPEC-XXX`; otherwise use the latest SPEC under `.namba/specs/`.
- Read `.namba/specs/<SPEC>/spec.md`, `plan.md`, and `acceptance.md` before writing review notes.
- Update `.namba/specs/<SPEC>/reviews/product.md` with status, reviewer, findings, decisions, follow-ups, and recommendation.
- Prefer `namba-product-manager` as the review role when subagent routing is appropriate.
- Refresh `.namba/specs/<SPEC>/reviews/readiness.md` so the advisory summary reflects the latest review state.
- Keep the review advisory by default; surface missing depth or blockers clearly without silently turning the workflow into a hard gate.
