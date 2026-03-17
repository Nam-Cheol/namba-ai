# Release Notes Draft

Project: namba-ai
Project type: existing
Reference SPEC: SPEC-008
Generated: 2026-03-17T17:30:02+09:00

## Workflow Changes

- `namba update` regenerates `AGENTS.md`, repo-local skills, compatibility mirror skills, role cards, and repo-local Codex config from `.namba/config/sections/*.yaml`.
- `namba sync` refreshes product docs, codemaps, change summary, PR checklist, and release docs.
- `namba run SPEC-XXX --parallel` fans out into up to three git worktrees, merges only after every worker passes execution and validation, and preserves failing worktrees and branches for inspection.

## Release Guardrails

- `namba release` requires a git repository, the `main` branch, and a clean working tree.
- Validators from `.namba/config/sections/quality.yaml` run before the release tag is created.
- With no explicit version, `namba release` defaults to the next `patch` tag. Use `--bump minor|major` or `--version vX.Y.Z` when needed.
- `namba release --push` pushes both `main` and the new tag to the selected remote.

## Release Commands

```text
namba sync
namba release --bump patch
# or
namba release --version vX.Y.Z --push
```

## Expected Assets

- `namba_Windows_x86_64.zip`
- `namba_Windows_arm64.zip`
- `namba_Linux_x86_64.tar.gz`
- `namba_Linux_arm64.tar.gz`
- `namba_macOS_x86_64.tar.gz`
- `namba_macOS_arm64.tar.gz`
- `checksums.txt`
