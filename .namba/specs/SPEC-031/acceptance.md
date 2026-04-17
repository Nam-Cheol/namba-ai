# Acceptance

- [ ] `namba plan`, `namba harness`, and `namba fix --command plan` all use a shared planning-start resolver instead of scaffolding directly into the current workspace.
- [ ] The planning-start resolver can distinguish the shared/base workspace, an already dedicated worktree/branch, and an unsafe dirty or ambiguous workspace.
- [ ] Invoking planning from the shared/base workspace does not silently scaffold into place; V1 either creates a dedicated worktree/branch or stops with an explicit next-step contract.
- [ ] Invoking planning from an already dedicated worktree/branch can scaffold in place without creating a nested worktree.
- [ ] The operator has an explicit, documented way to opt into current-workspace scaffolding when that behavior is intentional.
- [ ] Unsafe dirty-workspace behavior is explicit and tested; the implementation does not auto-stash, auto-commit, or move unrelated changes.
- [ ] SPEC id allocation for planning commands no longer depends only on the current workspace's `.namba/specs/`; active local worktrees discovered from local git worktree state do not reuse the same next sequential `SPEC-XXX`.
- [ ] Branch and worktree naming are derived predictably from the allocated SPEC id and description slug, while honoring configured branch-base/prefix settings where applicable.
- [ ] Reuse of the current workspace is allowed only for a clean, non-root attached worktree whose branch matches the configured `spec_branch_prefix`; a clean non-base branch alone does not count as already isolated.
- [ ] If planning creates a worktree/branch before scaffolding and scaffolding then fails, the implementation preserves that isolation state and tells the operator how to resume or clean it up.
- [ ] Operator-facing output tells the user which SPEC id was allocated, which workspace action was taken, which branch/worktree was chosen or created, and what to do next.
- [ ] Success, override, and dirty-workspace refusal outputs are distinguishable enough that the operator can tell whether Namba reused the current isolated workspace, created a new one, or stopped without mutation.
- [ ] Plan-review guidance and related planning docs reflect the same isolation contract instead of implying "always scaffold here."
- [ ] The relationship between planning auto-isolation and the lower-level `namba worktree new` command is documented so naming/base expectations do not silently diverge.
- [ ] Regression coverage proves shared-workspace safety, dedicated-worktree reuse, current-workspace override behavior, dirty-workspace refusal, and duplicate-id avoidance.
- [ ] Validation commands pass
- [ ] Tests covering the new behavior are present
