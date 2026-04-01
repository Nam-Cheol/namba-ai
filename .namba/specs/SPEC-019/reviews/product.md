# Product Review

- Status: approved
- Last Reviewed: 2026-04-01
- Reviewer: Codex
- Command Skill: `$namba-plan-pm-review`
- Recommended Role: `namba-product-manager`

## Focus

- Challenge the problem framing, scope, user value, and acceptance bar before implementation starts.

## Findings

- The revised contract is more intuitive for bug work. Users can stay inside the `fix` family for both reviewable bugfix planning and direct repair, instead of being forced to infer that bugfix planning lives under an unrelated command family.
- The `fix --command plan|run` split gives Namba one clear bug-work story: plan when the bug needs a reviewable SPEC, run when the user wants the harness to repair the issue directly in the current workspace.
- Expanding the README and generated workflow docs is product-relevant, not scope creep. The repo-local skill surface is part of the product experience, and users need to understand why each command-entry skill exists plus what its important options mean without reading source code.
- The acceptance bar now protects the main user-facing trust points: help flows are read-only, `fix` behavior is explicit, direct repair stays in the current workspace, and docs or skills must explain the same contract everywhere.

## Decisions

- Approve the `fix` command-family model where `namba fix --command plan` handles authored bugfix planning and plain `namba fix` or `namba fix --command run` handles direct repair.
- Keep `namba plan` focused on feature planning so the top-level product mental model stays simple instead of turning `plan` into a mixed feature and bugfix switchboard.
- Treat README, generated workflow docs, and skill descriptions as product deliverables that must explain intent, command mapping, and relevant options for each user-facing repo-local Namba skill.
- Keep GitHub handoff outside `fix`; users should continue to reach for `namba pr` and `namba land` after a successful repair flow.

## Follow-ups

- Make the migration path obvious for existing users who remember `namba fix` as a planning shortcut. Help text and error or redirect copy should explicitly point them to `namba fix --command plan`.
- Keep the option surface small and memorable. If future expansion is needed, preserve the current mental model instead of growing a large matrix of hidden modes.
- Use concrete examples in README and help output for feature planning, bugfix planning, and direct repair so the contract is learnable from examples, not just prose.

## Recommendation

- Advisory recommendation: approved. Proceed with implementation, keeping the delivery centered on bug-work discoverability, explicit command meaning, and user-facing documentation clarity.
