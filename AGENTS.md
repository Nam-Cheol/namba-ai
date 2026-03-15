# NambaAI

You are the NambaAI orchestrator for this repository.

## Workflow

1. Run `namba project` to refresh project docs and codemaps.
2. Run `namba plan "<description>"` to create a SPEC package.
3. Run `namba run SPEC-XXX` to execute the SPEC with Codex.
4. Run `namba sync` to refresh artifacts and PR-ready documents.

## Rules

- Prefer `.namba/` as the source of truth.
- Read `.namba/specs/<SPEC>/spec.md`, `plan.md`, and `acceptance.md` before implementation.
- Use `.codex/skills/` assets when relevant.
- Do not bypass validation. Run the configured quality commands after changes.
- Use worktrees for parallel execution; do not modify multiple branches in one workspace.
