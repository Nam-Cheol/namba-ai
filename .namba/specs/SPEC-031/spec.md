# SPEC-031

## Goal

Make SPEC planning start from an explicit, isolated branch/worktree flow so `namba plan`, `namba harness`, `namba fix --command plan`, and bundled plan-review workflows no longer silently scaffold into the wrong workspace.

## Context

- Project: namba-ai
- Project type: existing
- Language: go
- Mode: tdd
- Work type: plan
- Planning surface: `namba plan "<description>"`
- Command choice rationale: this is a workflow/orchestration slice on Namba's planning entry points, not a direct bugfix and not a reusable skill/agent artifact by itself.
- Verified local context as of 2026-04-17:
  - `AGENTS.md` says each new SPEC or task should start on a dedicated branch from `main`, and parallel execution should use worktrees instead of mixing branches in one workspace.
  - `internal/namba/namba.go` routes `runPlan`, `runHarness`, and `runFixPlanSubcommand` straight into `createSpecPackage(...)`, which writes the scaffold into the current workspace immediately.
  - `createSpecPackage(...)` allocates the next SPEC id through `nextSpecID(filepath.Join(root, specsDir))`, and `nextSpecID(...)` scans only the current workspace's `.namba/specs/` directory.
  - `runWorktreeNewSubcommand(...)` already exists, but it is a separate manual step. It creates a new worktree from `HEAD` and is not integrated with the planning commands.
  - `.agents/skills/namba-plan-review/SKILL.md` resolves whether to use `namba plan`, `namba harness`, or `namba fix --command plan`, but it does not first create or validate a dedicated worktree/branch.
  - During the 2026-04-17 `SPEC-030` bugfix split, the team had to move the fix into a separate worktree manually and hand-author `SPEC-031` because branch-local `nextSpecID(...)` would otherwise reuse the same next `SPEC-XXX`.

## Target Reader

- Namba maintainers who want the actual planning entry points to match the repository's documented branching/worktree policy.
- Operators who expect "start a SPEC" to be safe by default instead of mutating whichever workspace they happen to be standing in.
- Future implementers of planning and review automation who need one explicit workspace-isolation contract instead of scattered branch/worktree assumptions.

## Problem

Planning policy and planning behavior are currently misaligned:

1. Policy says "dedicated branch/worktree," but commands scaffold in place
   - The repo contract says new SPEC work should start isolated from `main`.
   - The current planning commands mutate the current workspace immediately.
2. Workspace intent is implicit instead of explicit
   - Starting a SPEC from the root workspace, from `main`, or from an unrelated dirty branch all look the same to the current scaffolder.
   - That makes it easy to create a planning package in the wrong place and only realize the mistake after follow-up work starts.
3. SPEC numbering is branch-local today
   - `nextSpecID(...)` looks only at the current workspace's `.namba/specs/`.
   - Parallel worktrees created from older refs can therefore choose the same next `SPEC-XXX`, which turns isolation into a numbering-collision hazard.
4. Plan-review ergonomics hide the isolation decision
   - `$namba-plan-review` bundles creation plus review flow, but it currently inherits the same "scaffold here" behavior.
   - The user expectation that planning should also prepare the worktree is therefore reasonable, even though the current implementation does not do it.

## Desired Outcome

- Planning commands share one workspace-start contract before any `.namba/specs/<SPEC>` files are written.
- Starting a SPEC from the shared root workspace or another unsafe context does not silently scaffold into place.
- Dedicated worktree/branch creation is either automatic or explicitly guided by the same planning preflight instead of being a separate tribal-knowledge step.
- Active local worktrees do not allocate duplicate sequential `SPEC-XXX` ids.
- The current-workspace path remains available only as an explicit, opt-in escape hatch for advanced use.
- Skills, docs, and command output explain which workspace action happened, where the scaffold lives, and what the next operator step is.

## V1 Success Definition

- `namba plan`, `namba harness`, and `namba fix --command plan` all pass through one shared planning-start preflight instead of writing into the current workspace directly.
- The preflight can distinguish between:
  - the shared/root workspace or base branch
  - an existing dedicated worktree/branch that is already the right place to scaffold
  - an unsafe dirty or ambiguous workspace that should not be mutated automatically
- When invoked from the shared/root workspace, the planning flow can create a dedicated worktree/branch rooted from the configured base branch and scaffold there.
- When invoked from an already dedicated worktree/branch, the planning flow can scaffold in place without nesting another worktree.
- The operator can explicitly choose "scaffold here" when that is intentional, rather than getting it by accident.
- SPEC id allocation is deterministic across the active local worktree set discovered from local git worktree state, so parallel local planning does not reuse the same next `SPEC-XXX`.
- The operator-facing output reports the chosen SPEC id, workspace action, branch, and worktree path clearly enough that the next step is unambiguous.
- The same isolation rules are reflected in plan-review guidance so bundled review flows do not quietly bypass the new contract.

## Scope

- Define one shared planning workspace-resolution path for:
  - `namba plan`
  - `namba harness`
  - `namba fix --command plan`
- Decide and implement the V1 behavior matrix for:
  - shared/root workspace invocation
  - already isolated dedicated worktree/branch invocation
  - dirty or ambiguous workspace invocation
  - explicit current-workspace override
- Introduce a deterministic "next SPEC id" strategy that works across the active local worktree set discovered from local git worktree state instead of only the current `.namba/specs/`.
- Define how worktree path and branch names are derived from the allocated SPEC id and description slug while respecting project git settings where applicable.
- Update operator guidance and generated docs/skills so the new behavior is visible in:
  - help/usage text
  - command-skill guidance
  - plan-review guidance where it references SPEC creation
- Add regression coverage for the isolation and numbering contract.

## Design Approach

### 1. Shared Planning Start Resolver

- Introduce one reusable planning-start helper that runs before `createSpecPackage(...)`.
- This helper should own:
  - workspace classification
  - SPEC id allocation
  - branch/worktree naming
  - any required worktree creation
  - explicit "scaffold here" override handling
- The planning commands should stop owning these decisions independently.

### 2. Active-Worktree SPEC Id Allocation

- Replace current branch-local numbering for planning entry points with a deterministic allocator that uses `git worktree list --porcelain` as the authoritative active-local-worktree inventory.
- V1 should scan every active local worktree in that inventory, including:
  - the root workspace
  - the current workspace
  - Namba-managed worktrees under `.namba/worktrees/`
  - user-managed local worktrees that live outside the repo tree but are still attached to the same repository
- The allocator should stay sequential and repository-local; this SPEC does not need a global network-backed reservation system.

### 3. Safe-By-Default Workspace Behavior

- Invoking planning from the shared root workspace or configured base branch should not silently write files in place.
- V1 should prefer creating or directing the operator into a dedicated worktree/branch when the current workspace is clearly the shared surface.
- V1 should reuse the current workspace only when all of these are true:
  - the current path is an attached git worktree that is not the shared root workspace
  - the working tree is clean enough for scaffolding
  - the current branch is not the configured base branch
  - the current branch matches the configured `spec_branch_prefix`
- A clean non-base branch by itself is still ambiguous; if the reuse rule above is not satisfied, the command should either create a fresh isolated worktree or require an explicit override.
- Dirty state should be handled explicitly:
  - do not auto-stash
  - do not move unrelated changes
  - either refuse with a clear reason or require an explicit override
- If the flow creates a worktree/branch before scaffolding and scaffolding then fails, V1 preserves the newly created worktree/branch and returns a clear resume-or-cleanup message. Do not silently delete partially prepared isolation state.

### 4. Explicit In-Place Escape Hatch

- Keep an intentional current-workspace mode for advanced users and tests.
- The override should be obvious in help text and command output so "I meant to scaffold here" is distinguishable from "Namba just wrote here by default."
- The override should be compatible with an already dedicated worktree/branch, but it must not weaken the safe default for the shared workspace.

### 5. Plan-Review Alignment

- `$namba-plan-review` and related generated review guidance should align with the same isolation contract.
- Bundled review flows do not need to re-implement worktree logic themselves, but they must not describe or imply a stale "always scaffold here" behavior.

### 6. Manual Worktree Command Relationship

- `namba worktree new` remains a lower-level manual command in V1.
- Planning auto-isolation should respect git-strategy base/prefix settings even if `namba worktree new` keeps its more direct semantics.
- Internal helper reuse is encouraged where it prevents divergence, but identical operator behavior between planning auto-isolation and the manual worktree command is not required in this SPEC.

### 7. Operator-Facing Result Contract

- Every planning-start path should emit one concise summary block that reports `SPEC id`, `workspace action`, `branch`, `worktree path`, and `next step`.
- The runtime contract should distinguish clearly between `reused current isolated workspace`, `created isolated workspace`, and `refused due to unsafe context`.
- Explicit current-workspace override output should be visibly exceptional so it cannot be mistaken for the safe default path.

## Implementation Priority

1. Lock the workspace-classification matrix and the current-workspace override contract.
2. Define the active-local-worktree SPEC id allocation rule before wiring any auto-worktree creation.
3. Introduce the shared planning-start helper and move `plan`, `harness`, and `fix --command plan` onto it.
4. Wire dedicated worktree/branch creation for shared-workspace entry.
5. Update command help, skills, and plan-review guidance to describe the new behavior accurately.
6. Add regression tests for safe default behavior, in-place allowed behavior, dirty-workspace refusal, and numbering collisions.
7. Run validation and refresh managed docs only after the runtime contract is stable.

## Initial Delivery Boundary

- First slice deliverables:
  - shared planning-start resolver
  - active-local-worktree SPEC id allocation
  - safe default behavior for shared/base workspace invocation
  - explicit current-workspace override
  - plan/harness/fix-plan integration
  - updated plan-review guidance
  - regression tests covering the new planning contract
- Follow-up slices may add:
  - richer workspace policy customization
  - branch naming policy per planning mode
  - automatic PR/bootstrap guidance immediately after scaffold creation
  - coordination across refs that are not currently checked out in a local worktree

## Non-Goals

- Do not reopen the already separated `regen` stale-cleanup bug in this SPEC.
- Do not change `namba run --parallel` fan-out behavior in this slice.
- Do not require a network-backed SPEC id reservation service.
- Do not auto-migrate or rewrite existing open SPEC directories across branches.
- Do not invent a second planning artifact model outside `.namba/specs/<SPEC>/`.

## Design Constraints

- Keep `.namba/specs/<SPEC>/` as the planning artifact model.
- Preserve sequential `SPEC-XXX` naming.
- Do not auto-stash, auto-commit, or relocate unrelated user changes.
- Make the operator-visible branch/worktree decision explicit.
- Keep the contract compatible with repo policy in `AGENTS.md` rather than requiring users to remember a separate unwritten workflow.
