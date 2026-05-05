# Engineering Review

- Status: clear
- Last Reviewed: 2026-05-05
- Reviewer: namba-planner
- Command Skill: `$namba-plan-eng-review`
- Recommended Role: `namba-planner`

## Focus

- Lock architecture, sequencing, failure modes, trust boundaries, and validation strategy before execution starts.

## Findings

- Initial blockers were branch resolution for existing SPECs, active-SPEC-aware PR and sync artifacts, cooperative pause and stop semantics, serialized readiness refresh after parallel reviews, and required-check ambiguity.
- The revised SPEC resolves branch ambiguity by preferring persisted `expected_branch`, accepting exactly one matching `spec/<SPEC-ID>-*` branch, deriving and persisting a deterministic fallback branch before creation, and blocking on multiple matches.
- The revised SPEC separates queue-level state from per-SPEC phase state and explicitly persists pause and stop request flags, active SPEC id, expected branch, run log id, observed head SHA, and last safe checkpoint.
- The revised SPEC avoids shelling together top-level commands blindly: queue implementation must use active-SPEC-aware PR and sync helpers where current code would otherwise summarize the repository's latest SPEC.
- Parallel review is now safe for artifacts because readiness refresh is required to happen once, serially, after all three review artifacts are complete.
- Required-check handling is implementation-ready: prefer explicit GitHub required-check evidence, or record and test the stricter fallback that all surfaced PR checks must be green.

## Decisions

- Queue should be implemented as an internal orchestrator over reusable helpers, not as a fragile shell-out conveyor of `namba sync`, `namba pr`, and `namba land`.
- Queue state needs queue-specific atomic writes; active queue state should not use direct overwrite helpers.
- Git/GitHub state inspection deserves separable test seams so resume and gate behavior can be tested without live GitHub.
- `clear-with-followups` is acceptable only with machine-readable `[non-blocking]` or `[post-implementation]` follow-up tags.

## Follow-ups

- [non-blocking] During implementation, choose file names that keep queue state discoverable, such as `.namba/logs/queue/state.json` plus a report markdown or text file.
- [post-implementation] Consider promoting shared PR/check helpers back into `pr_land_command.go` once queue behavior proves the active-SPEC-aware contract.

## Recommendation

- Clear to proceed. The remaining follow-ups are implementation-shaping, not planning blockers.
