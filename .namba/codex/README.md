# Codex Integration

`namba-ai` is configured for Codex-native Namba workflow.

## What `namba init .` Enables

- Creates `AGENTS.md` with Namba orchestration rules.
- Creates repo-local skills under `.agents/skills/`.
- Creates a compatibility mirror under `.codex/skills/`.
- Creates Codex delegation role cards under `.codex/agents/`.
- Creates repo-local Codex config under `.codex/config.toml`.
- Creates `.namba/` project state, configs, docs, and SPEC storage.

## How Codex Uses Namba After Init

1. Open Codex in the initialized project directory.
2. Codex loads `AGENTS.md` and repo skills.
3. Invoke `$namba` or ask Codex to use the Namba workflow.
4. Use built-in Codex delegation with the role cards in `.codex/agents/` when multi-agent work is appropriate.
5. Use `namba project`, `namba update`, `namba plan`, `namba fix`, `namba run SPEC-XXX`, and `namba sync` as workflow commands.

## Workflow Command Semantics

- `namba update` regenerates `AGENTS.md`, repo-local skills, compatibility mirror skills, role cards, and `.codex/config.toml` from `.namba/config/sections/*.yaml`.
- `namba sync` refreshes `.namba/project/*` docs, release notes/checklists, and codemaps.
- `namba release` requires a clean `main` branch and passing validators before it creates a tag. `--push` pushes both `main` and the new tag.
- `namba run SPEC-XXX --parallel` refers to the standalone runner path. It uses git worktrees, merges only after every worker passes execution and validation, and preserves failed worktrees and branches for inspection.

## Claude to Codex Mapping

- `CLAUDE.md` becomes `AGENTS.md`.
- Claude skills become repo-local Codex skills.
- Claude subagents become explicit role-card files used with Codex multi-agent delegation.
- Claude hooks become explicit validator and sync steps in Namba.
- Claude custom workflow commands become `$namba`, built-in Codex slash commands, and the `namba` CLI.

## Important Distinction

- In interactive Codex sessions, `namba run SPEC-XXX` means Codex should execute the SPEC directly in-session.
- The standalone `namba run` CLI remains available for non-interactive runner-based execution.
- Tokens and PATs are intentionally excluded from generated config. Use `gh auth login` or `glab auth login` instead.
