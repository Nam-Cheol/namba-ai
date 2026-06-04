# SPEC-068

## Problem

`namba init`, `namba update`, and `namba regen` still present their interactive or long-running work through console-oriented text output. `namba init` in particular relies on prompt helpers and final `fmt.Fprintln` status logs, which makes the first-run experience less navigable than the rest of the Codex-native workflow.

The requested change is to move these surfaces to Bubble Tea v2 and Lip Gloss v2 without changing the underlying Namba domain behavior.

## Goal

Replace the interactive `namba init` console flow with a real Bubble Tea TUI wizard, and add TUI progress/result presentation for `namba update` and `namba regen`.

## Scope

- `namba init`
  - Keep the current profile choices and flow semantics.
  - Preserve the existing defaults, detection behavior, validation, generated files, and manifest behavior.
  - Replace the interactive rendering and navigation with a Bubble Tea model-driven wizard.
  - Print a plain text created/skipped summary after the TUI exits.
- `namba update`
  - Keep the current self-update behavior, checksum verification, platform targeting, and error semantics.
  - Add TUI progress/result presentation only; do not turn update into a multi-step wizard.
- `namba regen`
  - Keep the current managed-output replacement behavior and session-refresh warning semantics.
  - Add TUI progress/result presentation only; do not add new regeneration choices.

## Non-Goals

- Do not add new init choices or remove existing choices.
- Do not change generated scaffold contents except where tests require deterministic summary reporting.
- Do not change update release resolution, checksum verification, binary replacement, or Windows scheduling semantics.
- Do not change regen ownership rules for managed versus user-authored outputs.
- Do not require a TUI in CI, piped, non-TTY, or `--yes` flows.

## Implementation Constraints

- Use Bubble Tea v2 for the interactive init flow.
- Use Lip Gloss v2 for all visual styling.
- Prefer the current v2 module import paths:
  - `charm.land/bubbletea/v2`
  - `charm.land/lipgloss/v2`
- The init TUI must be model-driven with `model`, `Init`, `Update`, and `View`.
- Support keyboard navigation with up, down, left, right, Enter, Esc, and q.
- Use Bubble Tea AltScreen for the interactive wizard.
- Do not implement the TUI as plain `fmt.Println` or `fmt.Fprintln` logs.
- Keep non-interactive fallback behavior for CI and non-TTY environments.
- Separate domain init logic from TUI rendering logic so tests can cover profile resolution and rendering separately.

## Current Code Evidence

- `internal/namba/init_command.go`
  - `runInit` resolves profile, writes files, writes the manifest, and currently prints final console logs.
  - `resolveInitProfileWithScan` chooses the interactive wizard only when `!opts.Yes && a.isInteractiveTerminal()`.
  - Existing prompt functions and option lists define the current flow and must remain semantically intact.
- `internal/namba/self_update_command.go`
  - `runUpdate` downloads archive and checksums, verifies, extracts, and replaces or schedules the binary.
- `internal/namba/update_command.go`
  - `runRegen` regenerates managed outputs, preserves ownership rules, prints session-refresh guidance, and prints cached version advisory.
- `internal/namba/init_wizard_test.go`, `internal/namba/update_command_test.go`, and `internal/namba/self_update_command_test.go`
  - Existing tests characterize the init choices, regen behavior, and update semantics that must not drift.

## Acceptance

- `namba init` interactive mode runs through a Bubble Tea v2 AltScreen wizard with model-driven `Init`, `Update`, and `View`.
- `namba init --yes` and non-TTY or CI init flows remain non-interactive.
- `namba init` preserves the current choices, defaults, profile validation, generated files, and manifest behavior.
- After interactive init exits, stdout includes a plain text summary of created and skipped files.
- `namba update` shows TUI progress/result presentation without changing update semantics or adding wizard choices.
- `namba regen` shows TUI progress/result presentation without changing regen ownership rules or adding wizard choices.
- Bubble Tea model update tests cover navigation keys, confirm, back/escape, and quit behavior.
- Non-TTY fallback tests cover init, update, and regen presentation fallback.
- Smoke validation covers `namba init`, `namba update`, and `namba regen` in safe test fixtures or mocked I/O where network or binary replacement would otherwise be unsafe.

## Context

- Project: namba-ai
- Project type: existing
- Language: go
- Mode: tdd
- Work type: plan
