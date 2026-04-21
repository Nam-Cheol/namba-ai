# Engineering Review

- Status: clear
- Last Reviewed: 2026-04-17
- Reviewer: Codex
- Command Skill: `$namba-plan-eng-review`
- Recommended Role: `namba-planner`

## Focus

- Lock architecture, sequencing, failure modes, trust boundaries, and validation strategy before execution starts.

## Findings

- 2026-04-17: The shared-resolver boundary is the right abstraction. `runPlan`, `runHarness`, and `runFixPlanSubcommand` are already thin wrappers over `createSpecPackage(...)` in `internal/namba/namba.go`, and the repo already has reusable git primitives such as `parseGitWorktrees(...)` and `hasWorkingTreeChanges(...)`. This should land as one planner-specific decision object ahead of `loadSpecPackageScaffoldContext(...)`, not as three separate command patches.
- 2026-04-17: The revised SPEC now fixes the largest classifier gap. Reuse is no longer implied by any clean non-base branch; it is pinned to a clean, non-root attached worktree whose branch matches the configured `spec_branch_prefix`. That is the right conservative default for avoiding accidental scaffolding into unrelated feature branches.
- 2026-04-17: The allocator source of truth is now correctly scoped to `git worktree list --porcelain`, which closes the earlier collision risk for user-managed attached worktrees outside `.namba/worktrees/`.
- 2026-04-17: The relationship to `namba worktree new` is now explicit enough for V1. Planning auto-isolation can use git-strategy-aware naming/base rules while the manual worktree command remains a lower-level escape hatch, as long as docs keep that distinction visible.
- 2026-04-17: The partial-failure policy is now explicit and implementation-friendly. If isolation is created and scaffolding then fails, the new worktree/branch is preserved and the operator gets a resume-or-cleanup message. That aligns with the repository's broader bias toward preserving inspectable state on failure.

## Decisions

- Proceed with a shared planning-start resolver; do not split workspace classification, SPEC id allocation, naming, and output reporting across individual command entry points.
- Treat `git worktree list --porcelain` plus `.namba/config/sections/git-strategy.yaml` as the authoritative engineering inputs for workspace resolution.
- Keep the first implementation slice bounded to `namba plan`, `namba harness`, and `namba fix --command plan`; do not widen scope unless sharing a lower-level helper with `namba worktree new` clearly reduces divergence.

## Follow-ups

- Add regression coverage for: an unrelated clean feature branch, a non-Namba external worktree, dirty-workspace refusal, explicit in-place override, and scaffold failure after successful worktree creation.
- Keep implementation and help output aligned with the documented distinction between planning auto-isolation and the lower-level manual `namba worktree new` command.
- Keep validation expectations explicit in the implementation slice: unit coverage for the resolver/allocator plus command-level regression tests for operator-visible behavior.

## Recommendation

- Advisory recommendation: clear to proceed. The highest-risk planning contracts are now explicit enough to start implementation, with the remaining work focused on code-level execution and regression coverage rather than another planning round.
