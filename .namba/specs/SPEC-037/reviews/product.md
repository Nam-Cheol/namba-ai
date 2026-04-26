# Product Review

- Status: clear
- Last Reviewed: 2026-04-25
- Reviewer: Codex (`namba-product-manager`)
- Command Skill: `$namba-plan-pm-review`
- Recommended Role: `namba-product-manager`

## Focus

- Challenge the problem framing, scope, user value, and acceptance bar before implementation starts.

## Findings

- The core framing is sound: v1 positions hooks as an evidence-producing
  extension of `namba run`, not as a policy, approval, or rollback system. That
  boundary is consistent across `spec.md`, `plan.md`, and `acceptance.md`.
- Operator value is present and specific enough for a first slice: teams can
  attach repository-local checks, notifications, or logging scripts to run
  lifecycle events without introducing a second artifact model. This is a real
  workflow gain for operators who already depend on `.namba/logs/runs/`.
- Scope is intentionally conservative in the right places. Config-based
  registration, serial execution, advisory-by-default failures, and normalized
  event names keep the slice implementable while preserving runner
  extensibility.
- The revised package now makes missing config product-clear. `spec.md`,
  `plan.md`, and `acceptance.md` align on `.namba/hooks.toml` being optional,
  with missing config treated as a no-op that does not fail the run.
- Malformed config is also now operator-clear. The expected behavior is
  explicit: stop before preflight, record a config-error hook result with
  `status="error"`, `exit_code=-1`, `blocking=true`,
  `failure_action="stopped"`, and a concrete `error_summary`, then finalize
  evidence on that path.
- Unsupported tool observation behavior is now sufficiently clear for v1.
  `spec.md` states that Namba must not fabricate tool-boundary events, may
  expose a non-hook capability note such as `tool_observations: unsupported`,
  and must not imply `after_patch`, `after_bash`, or `after_mcp_tool` were
  evaluated when the runner lacks observation support. `acceptance.md` mirrors
  that by requiring no fake hook results.
- Non-goals are otherwise clear and useful. The SPEC explicitly excludes Codex
  compatibility mode, approvals, remote-routing, and day-one support for every
  runner, which keeps the v1 boundary disciplined.

## Decisions

- Keep the v1 product promise centered on local operator observability and
  deterministic evidence capture, not governance or enforcement workflows.
- Accept the current absent-config, malformed-config, and unsupported
  observation handling as product-sufficient for implementation.
- Do not require a cross-run summary or `namba sync` reporting enhancement in
  this slice; run evidence is a sufficient operator outcome for v1.

## Follow-ups

- Ensure implementation preserves the exact operator contract now written in
  the acceptance checklist, especially the malformed-config evidence fields and
  the "no fake hook results" rule for unsupported observations.

## Recommendation

- `clear`. The revised SPEC is product-ready for implementation; the previously
  ambiguous behaviors around missing config, malformed config, and unsupported
  tool observations are now explicit enough for operators and testable in
  acceptance.
