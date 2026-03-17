# Change Summary

Project: namba-ai
Project type: existing
Latest SPEC: SPEC-009
Generated: 2026-03-17T09:59:18Z

## Workflow Docs Synced

- README and product docs describe when to use `namba update` versus `namba sync`.
- Release docs describe `namba release` guardrails on a clean `main` branch plus optional `--push` behavior.
- Parallel run docs describe the worktree fan-out and merge-blocking policy for `namba run SPEC-XXX --parallel`.

## Refresh Commands

- `namba update` regenerates `AGENTS.md`, repo-local skills, compatibility mirror skills, custom agents, and `.codex/config.toml` from `.namba/config/sections/*.yaml`.
- `namba sync` refreshes `.namba/project/*` docs, release notes/checklists, and codemaps.
