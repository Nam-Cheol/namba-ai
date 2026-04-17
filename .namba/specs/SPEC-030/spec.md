# SPEC-030

## Problem

`namba regen` can delete tracked managed outputs that are outside the regen surface.

Current evidence from the repository:

- `runRegen` calls `replaceManagedOutputs(..., true)` through the shared cleanup path.
- That owner-matching mode treats every manifest entry owned by `namba-managed` as stale-cleanup eligible, not just AGENTS, repo skills, Codex agents, repo Codex config, or `.namba/codex/*`.
- In practice that can remove non-regen managed outputs such as `README*.md`, `docs/workflow-guide*.md`, `.namba/project/*`, and historical SPEC review readiness files.
- Once those tracked files disappear, normal work on unrelated slices starts looking like broad working-tree churn and makes it harder to trust `regen` during active feature work.

## Goal

Apply the smallest safe fix so `namba regen` only cleans up stale artifacts that belong to the regen-managed namespace while preserving other managed outputs.

## Context

- Project: namba-ai
- Project type: existing
- Language: go
- Mode: tdd
- Work type: fix

## Desired Outcome

- `namba regen` continues removing stale managed repo skills, Codex agents, legacy Codex skill mirrors, repo Codex config, and `.namba/codex/*` outputs when they no longer belong in the generated scaffold.
- `namba regen` no longer deletes non-regen managed outputs such as synced README bundles, `.namba/project/*`, or SPEC review readiness artifacts just because they are manifest-owned by Namba.
- Regression coverage proves both sides of the contract:
  - stale regen-managed artifacts are still removed
  - non-regen managed artifacts survive `regen`

## Non-Goals

- Do not redesign `sync`, `project`, or manifest ownership semantics beyond what is required to scope regen cleanup safely.
- Do not fold unrelated SPEC-030 tool-surface wording work into this bugfix branch.
