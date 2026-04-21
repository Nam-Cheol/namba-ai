---
name: namba-plan
description: Command-style entry point for creating the next feature SPEC package.
---

Use this skill when the user explicitly says `$namba-plan`, `namba plan`, or asks to create a new feature SPEC package.

Behavior:
- Prefer the installed `namba plan` CLI when available.
- Keep `namba plan` for feature-oriented SPEC work; use `namba harness` when the request is about reusable agent, skill, workflow, or orchestration scaffolding; use `$namba-create` when the user wants the repo-local skill or custom-agent artifact itself instead of another SPEC.
- When repo-managed MCP presets are configured, prefer them for planning context before broader web search; for example, use `context7` for library and framework docs, `sequential-thinking` for deeper decomposition, and `playwright` for browser-verified flows.
- Read `.namba/project/product.md`, `.namba/project/tech.md`, `.namba/project/mismatch-report.md`, `.namba/project/quality-report.md`, and any relevant `.namba/project/systems/*.md` artifacts before drafting the SPEC.
- Treat executable code and authoritative config as stronger planning evidence than docs, and preserve code-vs-doc conflicts instead of smoothing them out.
- Start with the safe-by-default planning worktree preflight: reuse the current workspace only when it is already a clean attached SPEC worktree whose branch matches the configured SPEC prefix.
- From the shared/base workspace or another ambiguous clean branch, create a fresh isolated worktree/branch before writing `.namba/specs/<SPEC>/`.
- Use `--current-workspace` only when the user intentionally wants in-place scaffolding instead of the default isolated-worktree path.
- Create the next sequential `SPEC-XXX` package under `.namba/specs/` after that workspace decision is explicit.
- Seed `.namba/specs/<SPEC>/reviews/` with product, engineering, design, and aggregate readiness artifacts.
- Point follow-up review work to `$namba-plan-pm-review`, `$namba-plan-eng-review`, and `$namba-plan-design-review`, or use `$namba-plan-review` when the user wants the create-plus-review loop bundled into one skill.
- Keep the scope concrete and implementation-ready.
