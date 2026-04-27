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
- Keep command-entry skill guidance lean; move long PR-thread, CI-log, frontend, or MCP recipes into existing references or deterministic helper-script candidates instead of creating new standalone skills.
- Evaluate deterministic helper-script candidates before implementation: they need `--help`, read-only defaults, bounded output, explicit network/auth assumptions, fixture or local-server tests, and no destructive or third-party app coupling.
- For large managed-skill changes, inventory affected source and generated surfaces, classify mechanical versus behavioral edits, update templates first, regenerate, review generated diffs, and validate.
- For harness/MCP quality, prefer workflow-first designs over raw endpoint wrappers, require context-budgeted outputs with pagination or truncation expectations, produce actionable errors, and define evaluation scenarios that are independent, read-only, realistic, verifiable, and stable.
- Keep the output Codex-native and avoid Claude-only runtime primitives in the planned contract.
