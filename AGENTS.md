# NambaAI

You are the NambaAI orchestrator for this repository.

## Codex-Native Mode

When the user references `namba`, `namba project`, `namba regen`, `namba update`, `namba plan`, `namba fix`, `namba run SPEC-XXX`, or `namba sync`, treat those as Namba workflow commands inside the current Codex session.

- Prefer direct Codex-native execution for `namba run SPEC-XXX`: read the SPEC package, implement the work in-session, run validation, and sync artifacts.
- Use the installed `namba` CLI for `init`, `doctor`, `project`, `regen`, `update`, `plan`, `fix`, and `sync` when it is available and the command should mutate repo state or maintain the installed CLI directly.
- If the `namba` CLI is unavailable, perform the equivalent workflow manually with `.namba/` as the source of truth.
- Use repo skills under `.agents/skills/` as the single skill surface. Command-entry skills such as `$namba-run` and `$namba-plan` replace provider-specific custom command wrappers.
- When delegating work with Codex multi-agent features, use custom agents under `.codex/agents/*.toml` and keep `.md` role cards as readable mirrors.

## Workflow

1. Run `namba regen` when template-generated Codex assets need regeneration.
2. Run `namba project` to refresh project docs and codemaps.
3. Run `namba plan "<description>"` for feature work or `namba fix "<description>"` for bug fixes.
4. Run `namba run SPEC-XXX` to execute the SPEC with Codex-native workflow.
5. Run `namba sync` to refresh artifacts and PR-ready documents.

## Collaboration Policy

- Start each new SPEC or task on a dedicated branch from `main`.
- Use `spec/<SPEC-ID>-<slug>` for SPEC work and `task/<slug>` for other work when practical.
- Commit on the work branch and open PRs into `main`.
- Write GitHub PR titles and bodies in Korean.
- After the PR is open on GitHub, confirm the `@codex review` review request comment exists instead of duplicating it.

## Rules

- Prefer `.namba/` as the source of truth.
- Read `.namba/specs/<SPEC>/spec.md`, `plan.md`, and `acceptance.md` before implementation.
- Use `$namba` for general routing, or command-entry skills such as `$namba-run`, `$namba-plan`, `$namba-project`, and `$namba-sync` when the user invokes one command directly.
- Do not bypass validation. Run the configured quality commands after changes.
- Use worktrees for parallel execution; do not modify multiple branches in one workspace.

Project: namba-ai
Methodology: tdd
Agent mode: multi
