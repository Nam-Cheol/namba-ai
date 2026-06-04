# Engineering Review

- Status: cleared
- Last Reviewed: 2026-06-04
- Reviewer: Codex default reviewer
- Command Skill: `$namba-plan-eng-review`
- Recommended Role: `namba-planner`

## Focus

- Lock architecture, sequencing, failure modes, trust boundaries, and validation strategy before execution starts.

## Findings

- Current code already has a useful boundary: `resolveInitProfileWithScan` chooses interactive mode only for `!opts.Yes && a.isInteractiveTerminal()`.
- `runInit` mixes domain writes and final console output. The implementation should extract a structured init execution summary before adding TUI rendering.
- `runUpdate` has network, checksum, extraction, executable replacement, and Windows scheduling side effects; TUI must wrap state presentation without changing these trust boundaries.
- `runRegen` depends on `replaceManagedOutputs` ownership rules and session-refresh warnings; the TUI must not hide these warnings.
- Bubble Tea v2 and Lip Gloss v2 should use current v2 module paths: `charm.land/bubbletea/v2` and `charm.land/lipgloss/v2`.

## Decisions

- Add a small CLI TUI layer rather than embedding Bubble Tea directly into domain write functions.
- Model tests should exercise `Update` directly so navigation behavior is deterministic and does not require a terminal.
- Smoke tests for `namba update` must use mocked download and executable boundaries or an equivalent safe fixture; do not perform a live self-update in validation.

## Follow-ups

- If adding dependencies changes `go.mod` and `go.sum`, run `go mod tidy` and include that in validation evidence.
- If terminal detection needs refinement, keep it local to App I/O boundaries and avoid global environment-only checks.

## Recommendation

- Cleared for implementation.
