# SPEC-024

## Status

Superseded draft. Do not execute `namba run SPEC-024` and do not use this package for a new implementation handoff.

## Original Problem

This draft originally proposed the phase-1 `$namba-create` workflow for preview-first creation of repo-local skills and project-scoped custom agents.

## Resolution

That phase-1 scope already landed through `SPEC-025` on `main`, including the generated `$namba-create` skill surface, the preview-first clarification contract, explicit `skill` / `agent` / `both` routing, the repo-managed `[agents].max_threads = 5` baseline, and the regen ownership split for user-authored outputs.

## Outcome

- Keep `SPEC-024` only as historical context for the abandoned pre-merge draft.
- Do not execute or review this SPEC as an active work item.
- Track any remaining post-`SPEC-025` generator-engine work in `SPEC-026`.

## Context

- Project: namba-ai
- Project type: existing
- Language: go
- Mode: tdd
- Work type: retired draft
- Superseded by:
  - `SPEC-025` for the phase-1 skill-first `$namba-create` contract
  - `SPEC-026` for the follow-up real generator engine
