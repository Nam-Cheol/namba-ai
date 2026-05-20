# SPEC-055

## Problem

The release workflow emits maintenance warnings during GitHub Actions execution:

- Node.js 20 deprecation notices from action/runtime usage in the release path.
- A `windows-latest` runner notice caused by using a moving Windows runner label.

These warnings make release runs noisy and less reproducible.

## Goal

Apply the smallest safe workflow fix so the release workflow runs without the Node.js 20 deprecation warning or `windows-latest` runner notice.

## Scope

- Target file: `.github/workflows/release.yml`.
- Pin the Windows release runner from `windows-latest` to `windows-2022`.
- Update only release workflow action/runtime usage needed to remove the Node.js 20 deprecation warning.

## Out of Scope

- Changes to other files under `.github/workflows/`.
- Changes to Namba CLI release logic.
- `windows-2025-vs2026` or future Windows image migration work.
- Tag-push validation that could publish a real GitHub Release.

## Constraints

- Preserve existing release behavior.
- Prefer explicit runner and action versions for release reproducibility.
- Verification must not publish a real GitHub Release.

## Context

- Project: namba-ai
- Project type: existing
- Language: go
- Mode: tdd
- Work type: fix
