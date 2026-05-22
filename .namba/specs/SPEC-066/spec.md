# SPEC-066

## Problem

NambaAI currently owns generated repo guidance surfaces under `.agents/skills/*`, `.codex/agents/*.md`, and `.codex/agents/*.toml`, but future repositories need those generated contracts to be clearer and more testable. The durable fix is core scaffold generator contract hardening: generator source should produce skill and custom-agent outputs that explain role purpose, read-only versus mutating boundaries, required outputs, pass/fail criteria, evidence expectations, security responsibilities, fallback implementer boundaries, and non-project-specific guidance.

This must be generator-first work. Manual edits to the currently checked-in generated `.agents/skills/*` or `.codex/agents/*` files are not the source of truth; generated output diffs are allowed only when they come from source changes and `namba regen`.

## Goal

Improve NambaAI core managed generators so `namba init` and `namba regen` automatically produce a stronger init/regen generated instruction surface quality contract for future repositories.

## Scope

- Update Namba-owned generator, template, render, and output contract paths, especially source-of-truth code in `internal/namba/codex.go`, `internal/namba/templates.go`, `internal/namba/create_engine.go`, `internal/namba/agent_runtime.go`, and related `internal/namba/*_test.go` files.
- Improve all Namba-managed generated instruction surfaces:
  - `.agents/skills/*`
  - `.codex/agents/*.md`
  - `.codex/agents/*.toml`
- Preserve `.namba/manifest.json` ownership and managed boundary semantics for those generated surfaces.
- Prioritize command-entry skills and project-scoped custom agents during implementation because those surfaces show the highest quality variance.
- Preserve a common quality contract across the full generated surface so lower-priority generated files inherit the same expectations.
- Include current-repository managed generated file updates only by running `namba regen` after generator/source changes.
- Treat current-repo generated diffs as secondary evidence. Primary evidence is fresh temporary repo `namba init` output and existing repo `namba regen` idempotence.

## Constraints

- Preserve `namba init` and `namba regen` parity.
- Keep `.namba/manifest.json` managed ownership correct.
- Do not perform downstream repo backfills.
- Do not manually patch currently checked-in `.agents/skills/*` or `.codex/agents/*` files as the source of truth.
- Do not introduce network-dependent checks or LLM-judge evaluations.
- Do not broaden this beyond Namba core managed generator/source-of-truth improvement.
- Keep generated guidance non-project-specific and suitable for future repositories.
- Generated files may change only as the result of generator/source changes and regeneration.

## Context

- Project: namba-ai
- Project type: existing
- Language: go
- Mode: tdd
- Work type: plan
