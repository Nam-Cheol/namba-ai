# NambaAI

You are the NambaAI orchestrator for this repository.

## Codex-Native Mode

When the user references `namba`, `namba project`, `namba update`, `namba plan`, `namba fix`, `namba run SPEC-XXX`, or `namba sync`, treat those as Namba workflow commands inside the current Codex session.

- Prefer direct Codex-native execution for `namba run SPEC-XXX`: read the SPEC package, implement the work in-session, run validation, and sync artifacts.
- Use the installed `namba` CLI for `init`, `doctor`, `project`, `update`, `plan`, `fix`, and `sync` when it is available and the command will update repository state more reliably.
- If the `namba` CLI is unavailable, perform the equivalent workflow manually with `.namba/` as the source of truth.
- Use repo skills under `.agents/skills/` first. `.codex/skills/` exists as a compatibility mirror.
- When delegating work with Codex multi-agent features, use the custom agents under `.codex/agents/*.toml` as the agent prompt source.

## Workflow

1. Run `namba update` when template-generated Codex assets need regeneration.
2. Run `namba project` to refresh project docs and codemaps.
3. Run `namba plan "<description>"` for feature work or `namba fix "<description>"` for bug fixes.
4. Run `namba run SPEC-XXX` to execute the SPEC with Codex-native workflow.
5. Run `namba sync` to refresh artifacts and PR-ready documents.

## Rules

- Prefer `.namba/` as the source of truth.
- Read `.namba/specs/<SPEC>/spec.md`, `plan.md`, and `acceptance.md` before implementation.
- Use the `$namba` skill as the primary command surface when the user explicitly invokes Namba inside Codex.
- Do not bypass validation. Run the configured quality commands after changes.
- Use worktrees for parallel execution; do not modify multiple branches in one workspace.

Project: namba-ai
Methodology: tdd
Agent mode: multi
