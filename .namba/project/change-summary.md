# Change Summary

Project: namba-ai
Project type: existing
Latest SPEC: SPEC-036

## Workflow Docs Synced

- README bundles and product docs describe when to use `namba update`, `namba regen`, `namba sync`, `namba pr`, and `namba land`.
- Release docs describe `namba release` guardrails on a clean `main` branch plus optional `--push` behavior.
- Run docs separate the default standalone flow, `namba run SPEC-XXX --solo`, `namba run SPEC-XXX --team`, and the worktree fan-out policy for `namba run SPEC-XXX --parallel`.
- AGENTS and Codex docs define the Namba output contract plus the fallback validator script at `.namba/codex/validate-output-contract.py`.
- SPEC packages can keep advisory plan-review artifacts under `.namba/specs/<SPEC>/reviews/` so product, engineering, and design review state stays visible before execution and PR handoff.
- Collaboration docs require one branch per SPEC/task from `main`, PRs into `main`, korean PR content, and Codex review requests via `@codex review`.

## Refresh Commands

- `namba update` self-updates the installed `namba` binary from GitHub Release assets.
- `namba regen` regenerates `AGENTS.md`, repo-local skills and command-entry skills under `.agents/skills`, `.codex/agents/*.toml` custom agents, readable `.md` role-card mirrors, `.namba/codex/*`, and `.codex/config.toml` from `.namba/config/sections/*.yaml`.
- `namba sync` refreshes `.namba/project/*` docs, release notes/checklists, codemaps, and any README bundles enabled in `.namba/config/sections/docs.yaml`.
- `namba pr` prepares the current branch for GitHub review by running sync and validation by default, then committing, pushing, opening or reusing the PR, and ensuring the Codex review marker exists.
- `namba land` optionally waits for checks, merges only when the PR is clean, and updates local `main` safely.

## Latest Review Readiness

- Latest readiness artifact: `.namba/specs/SPEC-036/reviews/readiness.md`
- Advisory summary: all review tracks clear
