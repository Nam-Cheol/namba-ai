---
name: namba-harness
description: Command-style entry point for creating the next harness-oriented SPEC package.
---

Use this skill when the user explicitly says `$namba-harness`, `namba harness`, or asks to create a harness-oriented SPEC package.

Behavior:
- Prefer the installed `namba harness` CLI when available.
- Use this path for reusable agent, skill, workflow, orchestration, or evaluation scaffolding instead of product feature delivery.
- Create the next sequential `SPEC-XXX` package under `.namba/specs/` without inventing a second artifact model.
- Seed `.namba/specs/<SPEC>/reviews/` with product, engineering, design, and aggregate readiness artifacts so the review flow stays aligned with `namba plan`.
- Keep the output Codex-native and avoid Claude-only runtime primitives in the planned contract.
