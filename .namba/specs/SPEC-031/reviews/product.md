# Product Review

- Status: clear
- Last Reviewed: 2026-04-17
- Reviewer: namba-product-manager
- Command Skill: `$namba-plan-pm-review`
- Recommended Role: `namba-product-manager`

## Focus

- Challenge the problem framing, scope, user value, and acceptance bar before implementation starts.

## Findings

- The core problem framing is product-valid. This SPEC is fixing a trust break in the planning entrypoints, not just an internal git workflow mismatch. Today users can follow documented branch/worktree policy and still have `namba plan`, `namba harness`, or `namba fix --command plan` write into the wrong workspace by accident.
- The V1 boundary is mostly tight enough. Keeping this slice focused on planning-start preflight, deterministic local SPEC allocation, and plan-review alignment is the right product cut; it avoids dragging broader git automation, PR bootstrap, or artifact migration into the first delivery.
- One operator-experience gap still needs to stay explicit during implementation: when planning starts from the shared workspace and Namba creates a new worktree, the user will still be standing in the old shell location afterward. The product contract therefore cannot stop at "worktree created"; it must also make the handoff unmistakable by reporting the allocated SPEC id, branch, exact worktree path, and the immediate next command or action.
- The explicit in-place override is necessary, but it is also the main UX risk. If the override is named inconsistently across `plan`, `harness`, and `fix --command plan`, or explained as an implementation detail rather than an intentional escape hatch, users will relearn the same ambiguity this SPEC is trying to remove.
- The local-only numbering boundary is acceptable for V1, but only if docs and output stay honest about it. The current spec correctly scopes the allocator to active local worktrees rather than promising a repo-global reservation system, and that limitation should remain explicit in user-facing guidance.

## Decisions

- Proceed with this as one product slice rather than splitting worktree isolation and SPEC-id allocation into separate SPECs. Users experience them as one planning-start contract.
- Treat operator handoff messaging as part of the shipped behavior, not post-implementation polish. Safe creation in another worktree is only useful if the next step is obvious without reading source or tribal docs.
- Keep the override surface singular and memorable across all planning entrypoints. The product goal is "safe by default, explicit by exception," not a matrix of mode-specific escape hatches.

## Follow-ups

- In implementation and acceptance, require a concrete handoff contract for shared-workspace invocation: what was created, where it lives, whether scaffolding already happened there, and what the operator should do next.
- Keep dirty-workspace refusal copy action-oriented. Users should be told why Namba refused, whether the workspace was considered shared vs. ambiguous, and which escape hatch or cleanup step is valid.
- Make help/readme/skill updates show one short example per path: shared workspace auto-isolation, already-dedicated worktree reuse, and intentional current-workspace override.
- Preserve the V1 boundary around active local worktrees. If future demand appears for cross-clone or team-wide SPEC reservation, treat that as a follow-up SPEC rather than silently stretching this one.

## Recommendation

- Clear to proceed from a product perspective. The value, scope, and acceptance bar are strong enough for implementation, provided execution keeps the user-facing handoff after auto-isolation and the explicit in-place override contract as first-class product behavior.
