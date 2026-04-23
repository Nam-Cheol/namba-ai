---
name: namba-harness
description: Command-style entry point for creating the next harness-oriented SPEC package.
---

Use this skill when the user explicitly says `$namba-harness`, `namba harness`, or asks to create a harness-oriented SPEC package.

Behavior:
- Prefer the installed `namba harness` CLI when available.
- Use this path for reusable agent, skill, workflow, orchestration, or evaluation scaffolding when the user wants a reviewable SPEC first instead of generating the repo-local skill or agent artifact directly through `$namba-create`.
- Start with the same dedicated-branch planning contract as `namba plan`, and use `--current-workspace` only when the user intentionally wants to scaffold on the current branch without creating a dedicated SPEC branch.
- Do not create planning worktrees here either; temporary worktrees belong to overlapping `namba run SPEC-XXX --parallel` execution only.
- Create the next sequential `SPEC-XXX` package under `.namba/specs/` without inventing a second artifact model.
- Seed `.namba/specs/<SPEC>/reviews/` with product, engineering, design, and aggregate readiness artifacts so the review flow stays aligned with `namba plan`.
- Keep the output Codex-native and avoid Claude-only runtime primitives in the planned contract.
