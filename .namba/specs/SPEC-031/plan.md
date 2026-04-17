# SPEC-031 Plan

1. Refresh local planning-workflow context from:
   - `AGENTS.md`
   - `internal/namba/namba.go`
   - `.agents/skills/namba-plan-review/SKILL.md`
   - any help/readme surfaces that describe planning commands
2. Lock the current-state evidence before implementation starts:
   - `namba plan`, `namba harness`, and `namba fix --command plan` scaffold into the current workspace today
   - `nextSpecID(...)` is branch-local today
   - `worktree new` exists, but it is a manual step and is not integrated with planning
3. Define the shared planning-start contract:
   - how the current workspace is classified
   - when the shared/base workspace must not be mutated in place
   - reuse the current workspace only for clean, non-root attached worktrees whose branch matches the configured `spec_branch_prefix`
   - what explicit override allows current-workspace scaffolding
4. Define the V1 SPEC id allocation rule:
   - use `git worktree list --porcelain` as the active local worktree inventory
   - determine which discovered worktree roots participate in scanning
   - ensure sequential numbering remains deterministic
   - avoid duplicate ids across concurrently active local worktrees
5. Design branch and worktree naming rules for planned SPEC creation:
   - derive names from the allocated SPEC id and description slug
   - respect configured branch-base/prefix settings where applicable
   - keep follow-up commands predictable for the operator
6. Introduce a shared implementation boundary for planning-start logic and move these commands onto it:
   - `namba plan`
   - `namba harness`
   - `namba fix --command plan`
7. Implement safe-by-default workspace handling:
   - create a dedicated worktree/branch when planning starts from the shared/base workspace
   - scaffold in place only when the current workspace is already dedicated or the operator explicitly opts in
   - refuse unsafe dirty-state cases instead of auto-stashing or moving changes
   - preserve created worktrees/branches when scaffolding fails after isolation setup, then return a resume-or-cleanup message
   - emit a clear handoff summary that reports the workspace action, branch, path, and next step
8. Align operator guidance with the new contract:
   - help/usage text
   - generated docs/readmes
   - `.agents/skills/namba-plan-review/SKILL.md` and any related planning guidance
   - consistent success, override, and refusal language across all planning entrypoints
   - document how planning auto-isolation relates to the lower-level `namba worktree new` command
9. Add regression coverage for:
   - shared/base workspace invocation
   - dedicated worktree reuse
   - explicit current-workspace override
   - dirty-workspace refusal
   - user-managed active worktrees outside `.namba/worktrees/`
   - scaffold failure after successful worktree creation
   - duplicate-id avoidance across active local worktrees
10. Run validation commands.
11. Refresh managed artifacts with a source-aligned sync step once implementation changes stabilize.
