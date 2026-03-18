# Change Summary

Project: namba-ai
Project type: existing
Latest SPEC: SPEC-011
Generated: 2026-03-18T14:25:40+09:00

## Workflow Docs Synced

- README bundles and product docs describe when to use `namba update`, `namba regen`, and `namba sync`.
- Release docs describe `namba release` guardrails on a clean `main` branch plus optional `--push` behavior.
- Parallel run docs describe the worktree fan-out and merge-blocking policy for `namba run SPEC-XXX --parallel`.
- AGENTS and Codex docs define the Namba output contract plus the fallback validator script at `.namba/codex/validate-output-contract.py`.
- Collaboration docs require one branch per SPEC/task from `main`, PRs into `main`, korean PR content, and Codex review requests via `@codex review`.

## Refresh Commands

- `namba update` self-updates the installed `namba` binary from GitHub Release assets.
- `namba regen` regenerates `AGENTS.md`, repo-local skills and command-entry skills under `.agents/skills`, `.codex/agents/*.toml` custom agents, readable `.md` role-card mirrors, `.namba/codex/*`, and `.codex/config.toml` from `.namba/config/sections/*.yaml`.
- `namba sync` refreshes `.namba/project/*` docs, release notes/checklists, codemaps, and any README bundles enabled in `.namba/config/sections/docs.yaml`.
