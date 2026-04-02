# Release Notes Draft

Project: namba-ai
Project type: existing
Reference SPEC: SPEC-020

## Workflow Changes

- `namba update` self-updates the installed `namba` binary from GitHub Release assets.
- `namba regen` regenerates `AGENTS.md`, repo-local skills and command-entry skills under `.agents/skills`, `.codex/agents/*.toml` custom agents, readable `.md` role-card mirrors, and repo-local Codex config from `.namba/config/sections/*.yaml`.
- `namba sync` refreshes README bundles, product docs, codemaps, change summary, PR checklist, and release docs.
- `namba pr` prepares the current branch for GitHub review by syncing, validating, committing, pushing, opening or reusing the PR, and ensuring the Codex review marker exists.
- `namba land` optionally waits for checks, merges only when the PR is clean, and updates local `main` safely.
- `namba run SPEC-XXX` keeps the standard standalone Codex flow; `--solo` and `--team` request single-subagent or multi-subagent workflows inside one workspace; `--parallel` still fans out into up to three git worktrees and merges only after every worker passes execution and validation.
- Active collaboration defaults: one branch per SPEC/task from `main`, PRs into `main`, korean PR content, and Codex review requests via `@codex review`.

## Release Guardrails

- `namba release` requires a git repository, the `main` branch, and a clean working tree.
- Validators from `.namba/config/sections/quality.yaml` run before the release tag is created.
- With no explicit version, `namba release` defaults to the next `patch` tag. Use `--bump minor|major` or `--version vX.Y.Z` when needed.
- `namba release --push` pushes both `main` and the new tag to the selected remote.

## Release Commands

```text
namba sync
namba pr "release review"
namba land
namba release --bump patch
# or
namba release --version vX.Y.Z --push
```

## Expected Assets

- `namba_Windows_x86.zip`
- `namba_Windows_x86_64.zip`
- `namba_Windows_arm64.zip`
- `namba_Linux_x86_64.tar.gz`
- `namba_Linux_arm64.tar.gz`
- `namba_macOS_x86_64.tar.gz`
- `namba_macOS_arm64.tar.gz`
- `checksums.txt`
