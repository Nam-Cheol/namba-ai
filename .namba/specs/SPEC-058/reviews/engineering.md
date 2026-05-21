# Engineering Review

- Status: clear
- Last Reviewed: 2026-05-21
- Reviewer: namba-planner
- Command Skill: `$namba-plan-eng-review`
- Recommended Role: `namba-planner`

## Focus

- Lock architecture, sequencing, failure modes, trust boundaries, and validation strategy before execution starts.

## Findings

- The architecture should use existing `App` seams instead of adding a parallel CLI abstraction.
- `internal/namba` package tests are the right home because the fixture needs package-private runner and capability seams.
- Full golden output would be brittle; JSON field assertions and focused stdout checks are safer.
- cwd changes must use the existing mutex-backed helper or equivalent isolation.

## Decisions

- Implement a reusable fake external runner that handles only expected `git`, `gh`, `codex`, and shell command shapes.
- Use deterministic fake time and fake SHAs.
- Keep each scenario in a fresh temp repo.
- Run through `App.Run` for command dispatch coverage.

## Follow-ups

- [non-blocking] If the fake command table grows too large, split it into per-phase fixtures rather than making one global fake.
- [non-blocking] Document the focused test command in the eval evidence or test comments.

## Recommendation

- Clear to implement with attention to fake-runner diagnostics and cross-platform shell handling.
