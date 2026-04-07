# Product Review

- Status: clear
- Last Reviewed: 2026-04-07
- Reviewer: namba-product-manager
- Command Skill: `$namba-plan-pm-review`
- Recommended Role: `namba-product-manager`

## Focus

- Challenge the problem framing, scope, user value, and acceptance bar before implementation starts.

## Findings

- The SPEC frames the user problem correctly: this is not only a parser cleanup but a trust issue where help probing can unexpectedly mutate repository state.
- The target user definition is strong because it covers both human operators and Codex sessions, which is the real source of repeated accidental command probing.
- The acceptance bar is concrete enough to protect the high-risk commands (`project`, `sync`, `pr`, `land`, `release`, `update`, `run`, `worktree`, `init`) instead of limiting the fix to only planning commands.
- The delimiter requirement for `plan`, `harness`, and `fix` preserves an important user need: planning or repairing help-related bugs without the parser eating the literal flag text.

## Decisions

- Treat this as a CLI-wide contract hardening feature, not a one-off `plan`/`fix` patch.
- Keep doc updates secondary to behavior and regression-test hardening; user-facing docs should describe the contract after the implementation is made true.

## Follow-ups

- During implementation, keep the first delivery focused on predictability and safety rather than making usage text overly elaborate.
- Confirm whether `namba help <command>` should mirror the exact same text as `<command> --help` or merely the same semantics; either is acceptable if it is consistent and testable.

## Recommendation

- Clear to proceed from a product perspective. The scope, user value, and acceptance bar are strong enough for implementation.
