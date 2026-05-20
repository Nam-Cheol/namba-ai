---
name: namba-plan
description: Command-style entry point for creating the next feature SPEC package.
---

Use this skill when the user explicitly says `$namba-plan`, `namba plan`, or asks to create a new feature SPEC package.

Behavior:
- Before reading project docs, checking Git state, or running the CLI, run the Namba clarification gate on the user's raw request.
- If the request is short, broad, or missing target surface, user flow, scope boundaries, constraints, acceptance criteria, or validation, do not run `namba plan` yet. Ask 1-3 concise questions first; `게시판 만들어줘` is the canonical example that must ask questions before planning.
- Treat Codex Plan mode as the preferred clarification surface and `namba plan` as the executor: when native Plan mode choice UI is available, use it to ask the clarification questions before any CLI command.
- The Plan mode output must be converted into a refined description shaped as Goal, Scope, Constraints, and Acceptance, then passed to `namba plan "<refined description>"`.
- Repo hooks and skills cannot force Codex to switch modes by themselves; if native Plan mode UI is unavailable in the current session, fall back to concise text questions in chat instead of inventing a CLI wizard.
- Continue the clarification loop across turns until you can restate the request as Goal, Scope, Constraints, and Acceptance. Only then invoke the CLI with the refined description.
- Prefer the installed `namba plan` CLI when available.
- Keep `namba plan` for feature-oriented SPEC work; use `namba harness` when the request is about reusable agent, skill, workflow, or orchestration scaffolding; use `$namba-create` when the user wants the repo-local skill or custom-agent artifact itself instead of another SPEC.
- When repo-managed MCP presets are configured, prefer them for planning context before broader web search; for example, use `context7` for library and framework docs, `sequential-thinking` for deeper decomposition, and `playwright` for browser-verified flows.
- Read `.namba/project/product.md`, `.namba/project/tech.md`, `.namba/project/mismatch-report.md`, `.namba/project/quality-report.md`, and any relevant `.namba/project/systems/*.md` artifacts before drafting the SPEC.
- Treat executable code and authoritative config as stronger planning evidence than docs, and preserve code-vs-doc conflicts instead of smoothing them out.
- Keep planning in the current workspace. When branch-per-work is enabled, create or switch to the dedicated `spec/...` branch before writing `.namba/specs/<SPEC>/`.
- Use `--current-workspace` only when the user intentionally wants to scaffold on the current branch without creating a dedicated SPEC branch.
- Do not create a planning worktree. Reserve temporary worktrees for overlapping `namba run SPEC-XXX --parallel` execution, and expect them to disappear after the run finishes cleanly.
- Create the next sequential `SPEC-XXX` package under `.namba/specs/` after that branch decision is explicit.
- Seed `.namba/specs/<SPEC>/reviews/` with product, engineering, design, and aggregate readiness artifacts.
- Unless the invocation includes `--no-review`, treat the created SPEC as an automatic handoff and immediately continue with `$namba-plan-review SPEC-XXX` after the CLI prints the new SPEC ID.
- After the automatic handoff, read `.namba/specs/<SPEC>/reviews/readiness.md`; if it does not say `Cleared reviews: 3/3`, do not treat planning as complete. Continue with `$namba-plan-review SPEC-XXX` again, or run the missing `$namba-plan-pm-review`, `$namba-plan-eng-review`, and `$namba-plan-design-review` tracks before implementation.
- When `--no-review` is present, stop after scaffold creation and point any later review work to `$namba-plan-pm-review`, `$namba-plan-eng-review`, and `$namba-plan-design-review`, or `$namba-plan-review SPEC-XXX`.
- Keep the scope concrete and implementation-ready.
