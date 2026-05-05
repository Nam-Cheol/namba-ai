# SPEC-042

## Problem

NambaAI can create many SPEC packages with `namba plan`, but the main CLI does not yet have a durable way to process an existing backlog from review through implementation, pull request, checks, merge, and local main refresh.

Today an operator must manually repeat the same orchestration for each SPEC:

- choose the next existing `.namba/specs/SPEC-XXX` package
- create or reuse the dedicated `spec/...` branch according to Namba git strategy
- run product, engineering, and design plan review
- run implementation with `namba run --team SPEC-XXX`
- validate, sync, and prepare a PR
- wait for required checks and mergeability
- land the PR and refresh local `main`
- continue with the next SPEC

This is slow and error-prone when a user intentionally generates a batch of SPEC packages first and wants NambaAI to move through them in order. The missing piece is a queue or conveyor command surface that only consumes already-existing SPEC packages, safely skips work that is already complete, and stops on ambiguous or risky states instead of accidentally skipping a real failure.

## Goal

Add a durable SPEC queue conveyor to the main NambaAI CLI.

The primary command surface is:

- `namba queue start <SPEC-RANGE|SPEC-LIST> [--auto-land] [--skip-codex-review] [--remote origin]`
- `namba queue status`
- `namba queue resume`
- `namba queue pause`
- `namba queue stop`

The conveyor processes existing SPECs in deterministic order. It must not create new SPEC packages. For each target SPEC it resumes from the safest known step, runs only one active SPEC at a time, allows parallelism only inside that SPEC, and keeps durable state so an interrupted run can continue without duplicating branches, review requests, PRs, comments, commits, or merges.

For v1, the end-to-end terminal target is `landed`. Implemented or PR-ready SPECs are not treated as fully done; instead the conveyor skips the completed implementation or PR phases and resumes at the next safe phase. A SPEC is skipped as a whole only when evidence proves it is already landed on the configured base branch.

## Non-Goals

- Do not add a second planning command or create SPECs from the queue command.
- Do not process multiple active SPECs at once. Cross-SPEC concurrency is out of scope.
- Do not skip failed validation, failed checks, non-mergeable PRs, dirty worktrees, unknown PR state, or conflicting Git history.
- Do not silently remove or overwrite preserved worker worktrees from failed `namba run --parallel` or `--team` internals.
- Do not require auto-land by default. Auto-land must be an explicit option and must pass GitHub merge gates.
- Do not mutate repository-wide `auto_codex_review` config to support a one-off review skip.
- Do not expose every internal phase as the primary operator UX. Internal state may be detailed, but CLI reports should prioritize the user-facing state and next safe action.

## Existing Evidence

- `internal/namba/namba.go` defines the public command table and already includes `run`, `pr`, `land`, `sync`, and `worktree` commands.
- `internal/namba/planning_start.go` owns branch-per-work planning behavior and derives `spec/<SPEC-ID>-<slug>` branches from the configured git strategy.
- `internal/namba/spec_review.go` seeds product, engineering, design, and readiness review artifacts for every SPEC package.
- `internal/namba/execution.go` owns Codex execution requests, validation attempts, run artifacts, and hook evidence under `.namba/logs/runs/`.
- `internal/namba/parallel_lifecycle.go`, `parallel_run.go`, and `parallel_progress.go` already model bounded parallel work inside one SPEC execution.
- `internal/namba/pr_land_command.go` already knows how to prepare PRs, avoid duplicate `@codex review` comments, inspect PR status checks, validate merge readiness, merge, and update local `main`.
- `.namba/config/sections/git-strategy.yaml` has `branch_per_work: true`, `branch_base: main`, `spec_branch_prefix: spec/`, `pr_base_branch: main`, `git_provider: github`, and `auto_codex_review: true`.
- `.namba/config/sections/workflow.yaml` keeps Namba worktree parallelism at `max_parallel_workers: 3`; this is separate from same-workspace Codex subagent capacity.

## Scope

1. Add the queue command surface.
   - Register a public top-level `queue` command in the same command table as `run`, `pr`, and `land`.
   - Implement `start`, `status`, `resume`, `pause`, and `stop` subcommands.
   - Support explicit SPEC lists such as `SPEC-010 SPEC-012 SPEC-014`.
   - Support inclusive ranges such as `SPEC-010..SPEC-014`.
   - Reject missing, malformed, duplicated, or unordered ambiguous targets with a clear error.
   - Confirm every target already exists under `.namba/specs/`.
   - Define `--skip-codex-review` as "do not create a new `@codex review` request marker comment"; it must not mean skipping product, engineering, design, or GitHub review evidence.
   - When `--auto-land` is absent, stop after PR/check evidence in a waiting state and do not start the next SPEC until the current SPEC is landed or resume can prove it has already landed.
   - Add branch resolution for existing SPECs:
     - prefer the `expected_branch` already persisted in queue state
     - otherwise use an exact local branch match for `spec/<SPEC-ID>-*` only when there is exactly one match
     - otherwise derive a stable fallback slug from the current SPEC title or first problem line and persist it before branch creation
     - block on multiple existing matching branches or any branch/base ambiguity

2. Persist durable queue state.
   - Store local runtime state under `.namba/logs/queue/`, which is already ignored with the rest of `.namba/logs/`.
   - Write state atomically by using a queue-specific temporary file plus rename; do not use non-atomic direct overwrite helpers for active queue state.
   - Include queue id, created and updated timestamps, target specs, options, active spec, active step, per-SPEC status, branch, PR number and URL, validation evidence, land evidence, last safe checkpoint, and last blocker.
   - Separate queue-level state from per-SPEC phase state.
   - Persist at minimum `pause_requested`, `stop_requested`, `active_spec_id`, `expected_branch`, `current_run_log_id`, `last_observed_head_sha`, and `last_safe_checkpoint`.
   - Prevent more than one active queue and more than one active SPEC at a time.
   - Keep a human-readable report artifact beside the machine-readable state.

3. Implement the conveyor state machine.
   - Proposed ordered steps: `pending`, `reviewing`, `reviewed`, `branch_ready`, `running`, `validating`, `pr_ready`, `checks_pending`, `ready_to_land`, `landing`, `landed`, `skipped`, `blocked`, `paused`, `stopped`.
   - Map internal steps to operator-facing states: `running`, `waiting`, `blocked`, and `done`.
   - Recompute Git and GitHub truth before every irreversible step instead of trusting stale state.
   - Skip completed phases when evidence proves they are already complete.
   - Skip an entire SPEC only when landed evidence proves the v1 terminal target is already satisfied.
   - Treat implemented-but-unlanded work as resumable progress, not as landed.
   - Treat PR-open and checks-green without land as `waiting_for_land` when `--auto-land` is absent.
   - `pause` is cooperative: it records a pause request and the conveyor stops at the next safe checkpoint.
   - `stop` marks the queue stopped, disables automatic continuation, and preserves evidence for operator reporting; it must not delete branches, PRs, or worktrees.

4. Review and implementation flow.
   - For each SPEC, run product, engineering, and design plan review in parallel through the existing review skill or equivalent configured Codex reviewer path.
   - Let review tracks write their own artifacts in parallel, but refresh `.namba/specs/<SPEC>/reviews/readiness.md` once, serially, after all three review artifacts are complete.
   - Review pass criteria:
     - `clear`, `cleared`, `approved`, `pass`, and `passed` permit progress.
     - `clear-with-followups` permits progress only when every follow-up bullet is machine-tagged with `[non-blocking]` or `[post-implementation]`.
     - `blocked`, missing review artifacts, or ambiguous review status stop the queue as `blocked`.
   - If review does not clear or cannot be proven, stop as `blocked` with the review artifact paths.
   - Invoke `namba run --team SPEC-XXX` for implementation.
   - Preserve the existing validation and repair behavior from `run`; do not bypass validation.
   - Run active-SPEC-aware sync support after implementation changes when generated docs or project artifacts need refresh. Do not let queue PR/checklist output accidentally reference the repository's latest SPEC when the active queue SPEC is older.

5. PR, checks, and land flow.
   - Create or reuse the active SPEC PR through existing PR behavior or a queue-specific helper with the same GitHub safety gates.
   - Add an explicit queue option to skip creating a new `@codex review` request marker comment for the PR without editing global config.
   - Do not duplicate the configured review marker if it already exists.
   - Poll or inspect required checks through GitHub and require required checks to be green before auto-land. The preferred proof is explicit required-check data from GitHub branch protection or PR checks. If that proof is unavailable, v1 may choose the stricter fallback that all surfaced PR checks must be green, but the fallback must be recorded in queue evidence and tests.
   - Require GitHub mergeability, non-draft PR state, expected base branch, and no ambiguous review or check state before auto-land.
   - When `--auto-land` is absent, stop at a clear waiting state after PR and checks evidence rather than merging.
   - When `--auto-land` is present and gates pass, call land behavior and refresh local `main` before moving to the next SPEC.

6. Failure and resume behavior.
   - Failed validation, failed GitHub checks, non-mergeable PRs, dirty worktrees, missing `gh`, auth failures, diverged branches, and ambiguous PR state must become `blocked`, not skipped.
   - `resume` must continue from the next safe step after revalidating Git and GitHub state.
   - `pause` should request a pause at the next safe checkpoint; v1 does not need unsafe mid-process termination.
   - `stop` should mark the queue stopped and preserve enough state for a report, without deleting branches, PRs, or worktrees.

7. Operator UX.
   - `status` should print a compact queue report: active SPEC, current step, completed targets, skipped targets with reasons, blockers, PR links, and next command.
   - The default report should lead with one operator-facing state (`running`, `waiting`, `blocked`, or `done`) plus a concise detail such as `validating`, `waiting_for_checks`, `waiting_for_land`, or `PR not mergeable`.
   - The default report should summarize completed and skipped targets by count, with detailed skip reasons available through verbose output or the report artifact.
   - Blocked output should name the exact gate, evidence path, and suggested recovery action.
   - Waiting output should distinguish `waiting_for_checks`, `ready_to_land`, and `waiting_for_land` so the operator knows whether to wait, resume, or merge.
   - `pause_requested`, `paused`, `stopped`, and `blocked` must use distinct language and next-command guidance.
   - Reports should be useful in both Korean-speaking repository workflows and plain terminal logs; use existing project language settings where practical.

## Context

- Project: namba-ai
- Project type: existing
- Language: go
- Mode: tdd
- Work type: plan

## Risks

- Auto-landing across multiple SPECs can merge the wrong branch if GitHub state is stale or the current branch is not what the state file expects.
- Treating "implemented" as equivalent to "landed" would hide unmerged work; the state machine must distinguish phase completion from terminal SPEC completion.
- Parallel review and `namba run --team` may use Codex subagents, while Namba worktree parallelism remains bounded by workflow config. These concurrency models must not be conflated.
- Pause and stop semantics can be misleading if users expect process-level interruption. The first implementation should clearly define safe-checkpoint behavior.
- Existing `namba pr` always honors profile-level `auto_codex_review`; queue-level skip requires an explicit helper or option, not a temporary config mutation.
- Current `namba sync` and PR body support tend to summarize the repository's latest SPEC; queue work needs active-SPEC-aware helpers so older backlog SPECs do not receive misleading PR artifacts.
- Required GitHub check detection can be weaker than mergeability if the implementation only reads status rollups; ambiguous required-check evidence must block instead of auto-landing.

## Implementation Readiness

This SPEC is ready for product, engineering, and design review. The highest-risk implementation topics are the safe state machine, active-SPEC-aware PR and sync artifacts, branch resolution for already-existing SPECs, cooperative pause and stop behavior, and GitHub required-check evidence.
