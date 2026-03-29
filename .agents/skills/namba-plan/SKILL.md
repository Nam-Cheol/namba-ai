---
name: namba-plan
description: Command-style entry point for creating the next feature SPEC package.
---

Use this skill when the user explicitly says `$namba-plan`, `namba plan`, or asks to create a new feature SPEC package.

Behavior:
- Prefer the installed `namba plan` CLI when available.
- Create the next sequential `SPEC-XXX` package under `.namba/specs/`.
- Seed `.namba/specs/<SPEC>/reviews/` with product, engineering, design, and aggregate readiness artifacts.
- Point follow-up review work to `$namba-plan-pm-review`, `$namba-plan-eng-review`, and `$namba-plan-design-review` when the SPEC needs pre-implementation critique.
- Keep the scope concrete and implementation-ready.
