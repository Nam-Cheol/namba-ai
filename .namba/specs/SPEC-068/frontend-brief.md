# Frontend Brief

Task Classification: frontend-minor
Classification Rationale: This is a terminal UI change for Go CLI commands, not a browser frontend. The supported advisory classification is `frontend-minor`; the user-facing work is interaction, navigation, status presentation, and fallback behavior.
Frontend Gate Status: not-applicable
Problem Gate: not-applicable
Reference Gate: not-applicable
Critique Gate: not-applicable
Decision Gate: not-applicable
Prototype Gate: not-applicable
Prototype Evidence: n/a

## Current Pattern

- `namba init` uses terminal prompt helpers and plain summary logs.
- `namba update` and `namba regen` use plain progress/result logs.

## Intended Change

- `namba init` becomes a real Bubble Tea v2 AltScreen wizard while keeping the existing flow.
- `namba update` and `namba regen` get TUI progress/result presentation only.
- Non-interactive output remains available for CI and non-TTY use.

## Notes

- Use `charm.land/bubbletea/v2` and `charm.land/lipgloss/v2`.
- Keep styling restrained and legible; this is an operational CLI, not a marketing surface.
