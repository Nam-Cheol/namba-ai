# Engineering Review

- Status: approved
- Last Reviewed: 2026-05-21
- Reviewer: namba-planner
- Command Skill: `$namba-plan-eng-review`
- Recommended Role: `namba-planner`

## Focus

- Lock implementation boundaries, TDD order, queue-owned metadata rules, Windows prompt transport, fallback semantics, and regression coverage before execution starts.

## Findings

- The queue-owned allowlist blocker is resolved at the SPEC level. `spec.md` now pins the exception to queue scope, names the allowed active-run artifacts concretely, and explicitly forbids weakening non-queue clean checks for `plan`, `pr`, `land`, and `release`. That is narrow enough to keep the fix off the current global clean-check path.
- The local fallback blocker is resolved at the state-semantics level. The revised package now states when fallback is allowed, what it must do, what evidence/reporting it must emit, and which cases remain hard blockers instead of continuity cases. That gives implementation a clear boundary instead of hiding fallback inside generic PR failure handling.
- The TDD sequencing blocker is resolved. `plan.md` now front-loads failing tests for cleanliness classification, non-argv prompt transport, remote-handoff-unavailable fallback, and merge-conflict blocking before behavior changes, then finishes with targeted and full validation. That ordering is concrete and supports the repository’s TDD mode.
- The queue dry-run wording blocker is resolved. `spec.md` now defines the term as fixture-driven or `namba run --dry-run`-backed regression coverage and explicitly says it does not require a new public `namba queue --dry-run` flag. `acceptance.md` mirrors that intent cleanly.
- The execution-layer seam is sufficiently constrained for implementation. The package now clearly requires a non-argv prompt transport while preserving existing capability-based flag/config resolution and request artifacts, which is a precise enough target for the execution/codex runner layer.
- Remaining engineering risk is normal implementation risk rather than SPEC ambiguity. The main caution is to keep the queue-only cleanliness logic and fallback path additive, with no accidental broadening of shared Git or PR behaviors.

## Decisions

- The prior engineering blockers are resolved in the updated SPEC package.
- Implementation should proceed with a queue-scoped clean-check path, a shared execution-layer prompt transport seam, and an explicit fallback checkpoint/evidence path.
- No new public queue dry-run CLI surface is required by this package.

## Follow-ups

- During implementation, keep the queue-owned allowlist keyed to the active SPEC and active queue state only; do not treat arbitrary run-log files as safe by default.
- Keep fallback detection limited to remote-handoff-unavailable cases after sync and validation success, and preserve hard blocking for merge conflicts, failed checks, and review blockers.
- Validate with the targeted queue/execution fixture tests first, then `go test ./...`, `gofmt -l "cmd" "internal" "namba_test.go"`, and `go vet ./...`.

## Recommendation

- Proceed to implementation. The revised package resolves the prior ambiguity around queue-owned metadata scope, local fallback semantics, TDD sequencing, and queue dry-run meaning, and is now specific enough for a small safe fix.
