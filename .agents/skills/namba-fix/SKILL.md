---
name: namba-fix
description: Command-style entry point for direct bug repair or bugfix SPEC planning.
---

State effect: mixed. Help/probe paths are read-only; direct repair mutates the current workspace; `--command plan` creates a bugfix SPEC.

Use this skill when the user explicitly says `$namba-fix`, `namba fix`, or asks to repair a bug through Namba.

Behavior:
- Prefer the installed `namba fix` CLI when available.
- Treat `namba fix "<issue description>"` as the default direct-repair path in the current workspace.
- Use `namba fix --command run "<issue description>"` when the user wants the explicit direct-repair form.
- Use `namba fix --command plan "<issue description>"` when the user wants a reviewable bugfix SPEC package under `.namba/specs/` via the same dedicated-branch planning contract.
- For `namba fix --command plan`, run the same clarification gate as `namba plan` and `namba harness` before reading project docs, checking Git state, or running the CLI. Ask 1-3 concise questions first when the issue, target surface, scope, constraints, acceptance criteria, or validation are unclear.
- When clarification is needed for the planning path, prefer Codex Plan mode and pass only the refined Goal/Scope/Constraints/Acceptance description to `namba fix --command plan "<refined issue description>"`.
- Use `--current-workspace` only with `namba fix --command plan` when the user intentionally wants to scaffold on the current branch without creating a dedicated SPEC branch.
- Do not create planning worktrees for this path; worktrees are reserved for temporary overlapping `namba run SPEC-XXX --parallel` execution.
- Keep CLI help and flag probing read-only; `namba <command> --help`, `namba <command> -h`, and `namba help <command>` must not mutate repository state.
- Keep direct repairs small, add targeted regression coverage, run validation, and finish with `namba sync`.
