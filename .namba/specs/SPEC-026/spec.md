# SPEC-026

## Problem

`SPEC-025` shipped the phase-1 `$namba-create` contract and the ownership boundary that keeps user-authored outputs safe from `namba regen`, but the repository still does not have a real repo-tracked generator engine behind that surface.

The current gap is visible in authoritative repo state:

- `.agents/skills/namba-create/SKILL.md` describes a preview-first staged flow, but it is only instruction text.
- `internal/namba/codex.go` and `internal/namba/templates.go` generate the skill surface and routing docs, not an actual create engine that writes confirmed outputs.
- `internal/namba/namba.go` supports `plan`, `harness`, and `fix` SPEC creation, but there is no `create` execution path that turns a confirmed preview into repo artifacts.
- `internal/namba/create_workflow_test.go` currently proves the contract fixture and regen ownership behavior, but it does not prove real generation of `.agents/skills/<slug>/SKILL.md` or `.codex/agents/<slug>.toml` plus `.md`.

The remaining user problem is therefore not discovery or routing. It is the absence of a safe implementation that converts the confirmed `$namba-create` preview into durable repo-local skill and custom-agent files without breaking manifest ownership or regen rules.

## Goal

Implement a real Go generator engine for `$namba-create` that powers confirmed skill, agent, and both-mode creation; writes only to the existing allowlisted roots; updates manifest ownership safely; and keeps the user-facing surface skill-first instead of introducing a new `namba create` CLI in this slice.

## Context

- Project: namba-ai
- Project type: existing
- Language: go
- Mode: tdd
- Work type: plan
- Baseline already delivered by `SPEC-025`:
  - generated `$namba-create` skill surface
  - preview-first clarification contract
  - explicit `skill` / `agent` / `both` routing semantics
  - repo-managed `[agents].max_threads = 5` default for `agent_mode: multi`
  - ownership split so user-authored create outputs survive `namba regen`
- Existing artifact roots:
  - skills: `.agents/skills/<slug>/SKILL.md`
  - custom agents: `.codex/agents/<slug>.toml` and `.codex/agents/<slug>.md`
- Existing authoring and ownership surfaces:
  - `internal/namba/codex.go`
  - `internal/namba/templates.go`
  - `internal/namba/update_command.go`
  - `.namba/manifest.json`
- Existing regression anchors:
  - `internal/namba/create_workflow_test.go`
  - `internal/namba/update_command_test.go`
  - `internal/namba/spec_command_test.go`

## Desired Outcome

- `$namba-create` remains the primary user-facing entrypoint, but it now has a real repo-tracked engine behind the phase-1 contract.
- The engine exposes a normalized preview model that includes:
  - selected target: `skill`, `agent`, or `both`
  - normalized slug or name
  - exact output paths
  - overwrite impact
  - validation plan
  - whether session refresh guidance is expected
- A confirmed `skill` request writes only `.agents/skills/<slug>/SKILL.md`.
- A confirmed `agent` request writes `.codex/agents/<slug>.toml` and `.codex/agents/<slug>.md` together.
- A confirmed `both` request writes all expected files or none of them.
- Invalid slugs, path traversal attempts, silent overwrites, and partial agent mirror writes are rejected.
- Manifest ownership continues to distinguish user-authored create outputs from Namba-managed built-ins so `namba regen` preserves user-created artifacts.
- Generated docs and skill text clearly describe phase-2 behavior as a real generator entrypoint rather than a contract-only placeholder.
- Regression coverage proves confirmation gating, preview exactness, safe writes, and regen preservation.

## Target User

- Users who already understand `$namba-create` from the phase-1 docs and now expect confirmed previews to produce actual repo artifacts.
- Maintainers who need a safe, testable, repo-tracked path for generating skills and custom agents without hand-editing both Codex and manifest surfaces.
- Reviewers who need the generator behavior to stay deterministic, path-safe, and compatible with `namba regen`.

## Scope

- Implement an internal Go generator engine for `skill`, `agent`, and `both` requests.
- Define request, preview, and write models that normalize slug/name, selected target, overwrite policy, output paths, and refresh impact.
- Implement write logic that:
  - stays inside the existing skill and agent roots
  - writes agent `.toml` and `.md` mirrors atomically
  - treats `both` as an all-or-nothing write set
- Integrate manifest and ownership updates so successful create outputs are tracked as user-authored and survive `namba regen`.
- Update the generated `$namba-create` contract and related docs to reflect the real generator behavior while preserving the preview-first interaction.
- Add regression coverage for preview gating, write safety, ownership handling, and failure rollback behavior.

## Non-Goals

- Do not add a new `namba create` Go CLI command in this SPEC.
- Do not reopen or redesign the phase-1 decisions already delivered by `SPEC-025` unless they are strictly required to support the real generator engine.
- Do not introduce a second artifact model outside the existing `.agents/skills/*`, `.codex/agents/*`, and `.namba/manifest.json` surfaces.
- Do not change worktree `max_parallel_workers` or reopen same-workspace thread-limit decisions beyond using the current baseline.

## Design Constraints

- Keep `.namba/` as the workflow source of truth and `.namba/manifest.json` as the ownership record for generated surfaces.
- Keep `$namba-create` preview-first: no writes before confirmation.
- Keep explicit user intent authoritative over heuristic routing.
- Normalize slugs and names before they become durable file paths.
- Reject path traversal, invalid slugs, silent overwrite, and incomplete agent mirror states.
- Preserve the existing separation between Namba-managed built-ins and user-authored create outputs so `namba regen` stays safe and predictable.
- Keep the implementation wrapper-friendly so a future `namba create` CLI can call the same engine instead of replacing it.
