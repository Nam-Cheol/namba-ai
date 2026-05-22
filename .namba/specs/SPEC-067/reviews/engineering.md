# Engineering Review

- Status: cleared
- Last Reviewed: 2026-05-22
- Reviewer: Codex fallback engineering review
- Command Skill: `$namba-plan-eng-review`
- Recommended Role: `namba-planner`

## Focus

Challenge architecture, sequencing, testability, rollback, and validation before implementation starts.

## Findings

- `internal/namba/namba.go` currently mixes command routing, usage text, SPEC creation, report formatting, clarification, run/sync/worktree glue, config/manifest helpers, project docs, README support docs, and init wizard logic.
- A same-package extraction-first approach is appropriate. It reduces file size and responsibility without immediately creating import-cycle pressure.
- The target package examples from the request are useful directionally, but creating many subpackages at once would raise cycle risk. The SPEC correctly defers subpackages until dependency analysis proves they are acyclic.
- Phase ordering is low-to-high risk: pure formatting first, SPEC creation second, routing third, artifact/evidence fourth, runtime/queue fifth, release/security sixth, final verification seventh.
- The validation command list matches repository quality policy and CI coverage, including eval/report schema checks and race tests.

## Decisions

- Add characterization tests before moving behavior.
- Use temp repos and fake adapters for side effects and external integration boundaries.
- Keep assertions contract-focused rather than brittle full-output snapshots.
- Require `go list ./...` after refactor to catch dependency cycles.
- Require `wc -l internal/namba/namba.go` before and after to prove material size reduction.

## Follow-ups

- Implementation should record the Phase 0 line count and responsibility count before Phase 1.
- Each phase should be committed or otherwise checkpointed separately when practical.
- Any schema or generated layout change should be treated as a regression unless separately approved by a new SPEC.

## Recommendation

Cleared for implementation planning. Do not start refactor before Phase 0 tests and baseline artifacts exist.
