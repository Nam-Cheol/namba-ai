---
name: namba-fix
description: Command-style entry point for direct bug repair or bugfix SPEC planning.
---

State effect: mixed. Help/probe paths are read-only; direct repair mutates the current workspace; `--command plan` creates a bugfix SPEC.

Generated instruction contract for this command skill:
- Purpose: keep the role or command scope explicit, bounded, and testable.
- Boundary: honor read-only versus mutating state effects, configured sandbox mode, and assigned file or workflow ownership.
- Required output: report concrete actions, changed paths or artifacts, validation evidence, and pass/fail status or blockers.
- Pass/fail criteria: claim success only when acceptance criteria and configured validation are satisfied; otherwise name the exact blocker and impact.
- Evidence expectations: cite source artifacts such as SPEC files, `.namba/` configs, diffs, test output, PR/check links, or generated manifests instead of relying on unsupported assertions.
- Security responsibilities: never expose or commit secrets; treat auth, privacy, destructive commands, permission changes, and external network or credential use as security-sensitive.
- Destructive command and escalation policy: do not run destructive commands unless explicitly requested, and request approval for privileged, networked, or sandbox-blocked actions.
- Fallback implementer boundary: if a specialist path is unavailable and the main/default implementer takes over, stay within the assigned scope and preserve the same evidence and validation duties.
- Portability: keep durable guidance non-project-specific unless the current repository config or SPEC explicitly provides the project detail.

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
