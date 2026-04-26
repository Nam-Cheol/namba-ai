# Engineering Review

- Status: clear
- Last Reviewed: 2026-04-25
- Reviewer: namba-planner
- Command Skill: `$namba-plan-eng-review`
- Recommended Role: `namba-planner`

## Focus

- Re-check whether the revised SPEC is implementation-ready around the runner
  observation contract, the single lifecycle/evidence owner, and parallel v1
  worker-scope semantics.

## Findings

- The revised runner observation contract is now implementation-ready against
  the current runner seam. `internal/namba/execution.go` still exposes
  `runner.Execute(...)` as a terminal call, but the SPEC now defines the
  missing typed observation boundary explicitly: `observation_type`,
  `tool_name`, `tool_use_id`, timestamps, status, bash-specific command/cwd/
  exit code, bounded summaries, and output artifacts when available. That is
  enough to add an observation sink or callback without scraping free-form
  Codex output or leaking Codex hook names into orchestration.
- The single lifecycle/evidence ownership model is now explicit enough for the
  current `executeRun` shape. `internal/namba/execution.go` still has many
  early returns for preflight, execution, validation, repair, and progress-log
  failure, and `internal/namba/execution_evidence.go` still writes the manifest
  only at end states. The revised SPEC closes that gap by assigning one
  per-run lifecycle owner responsibility for buffering hook results, writing
  hook stdout/stderr artifacts as they finish, preserving the primary run
  failure, triggering `on_failure` once, and finalizing evidence only after
  relevant hook results are recorded.
- Parallel v1 scope is now concrete and aligned with existing ownership
  surfaces. `internal/namba/parallel_lifecycle.go` already keeps aggregate
  preflight/progress/merge/cleanup at the aggregate layer while worker
  execution reuses `executeRun` with worker log ids. Restricting v1 hooks to
  worker scope, recording hook results only in worker manifests, and keeping
  aggregate parallel evidence free of duplicated hook records matches the code
  that exists today.
- The remaining engineering risk is implementation discipline rather than
  missing design decisions. This slice should still reuse existing injected
  seams for clock, command execution, binary execution, and progress
  publishing, instead of bypassing them with direct process or filesystem
  control.

## Decisions

- Engineering review is clear for implementation on the three previously
  blocked areas: runner observations, lifecycle/evidence ownership, and
  parallel v1 scope.
- Keep runner-specific observation translation inside runner adapters and keep
  hook dispatch plus evidence finalization in Namba orchestration.
- Treat worker evidence as the source of truth for v1 hook results in parallel
  mode. Aggregate artifacts may summarize worker outcomes but should not
  duplicate hook records.
- Reuse existing injected seams for hook execution and tests so timeout,
  failure-path, and artifact assertions remain deterministic.

## Follow-ups

- Decide the manifest versioning path before code lands: keep additive `hooks`
  on `execution-evidence/v1` or bump the schema version. The revised SPEC now
  makes this an explicit compatibility choice rather than a blocker.
- Introduce the lifecycle owner early in implementation instead of bolting hook
  calls onto existing early-return branches incrementally.
- Keep the regression matrix strong on worker-only parallel behavior,
  unsupported observations producing no fake hook results, and failure-path
  finalization after blocking hooks or config/spawn/timeout errors.

## Recommendation

- Clear. The revised SPEC now makes the runner observation contract, the
  per-run lifecycle/evidence owner, and the parallel v1 worker-scope semantics
  explicit enough to implement against the current codebase without inventing
  behavior mid-flight.
