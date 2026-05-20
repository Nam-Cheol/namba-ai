# Engineering Review

- Status: approved
- Last Reviewed: 2026-05-20
- Reviewer: Codex (`namba-planner`)
- Command Skill: `$namba-plan-eng-review`
- Recommended Role: `namba-planner`

## Focus

- Lock architecture, sequencing, failure modes, trust boundaries, and validation strategy before execution starts.

## Findings

1. High: Remote-control and remote-environment evidence ownership needed to be
   pinned to one schema boundary. Current code uses a shared
   `codex_diagnostics` payload in project, run, and queue evidence, while
   execution manifests also have runtime extension bundles. A parallel schema
   would risk consumer drift.
2. High: Run and queue evidence currently build Codex diagnostics with
   `RunCommands: false`, so this SPEC must not require flaky or blocking Codex
   probes on those paths.
3. Medium: Status vocabulary needed normalization because existing diagnostics
   already use terms such as `detected`, `not_detected`, `unavailable`,
   `timed_out`, and `advisory_mismatch`, while this SPEC introduces remote
   readiness terms such as `disabled`, `enabled`, `configured`, and
   `local_fallback`.
4. Medium: Validation needed fixture or golden coverage for evidence shape
   changes, not only the repository-wide Go validation commands.
5. Medium: `SPEC-050` and `SPEC-052` boundaries needed operational constraints
   so implementation does not reopen hook/config compatibility or duplicate
   diagnostics ownership.

## Decisions

- New platform readiness status should be additive to the shared
  `codex_diagnostics` payload unless implementation proves a separate additive
  payload is safer.
- Project diagnostics may remain the richest command-running path. Run, queue,
  hook, and parallel-run evidence should use stable local metadata,
  already-collected diagnostics, config snapshots, or neutral fallback status.
- Remote-control and remote-environment readiness stays advisory and read-only.
  This SPEC must not depend on mutating runtime APIs.

## Follow-ups

- `spec.md`, `plan.md`, and `acceptance.md` were revised to require a
  field-level schema map before evidence struct changes.
- `spec.md`, `plan.md`, and `acceptance.md` were revised to define source
  precedence for remote-control and environment statuses.
- Acceptance now requires fixture or golden coverage for changed
  `execution-evidence/v1`, `queue-runner-evidence/v1`, and
  `project-codex-diagnostics-evidence/v1` shapes.
- Acceptance now constrains hook guard changes to stdlib-only readiness
  alignment without weakening existing guard behavior.
- Aggregate review tightened the hook boundary further: guard changes are
  limited to readiness wording, stale-reference cleanup, or no-op safety
  verification, while new hook behavior remains owned by `SPEC-050`.

## Recommendation

- Proceed after preserving the schema map, status-source strategy, and
  non-blocking run/queue evidence constraints during implementation.
