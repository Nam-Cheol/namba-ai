# Design Review

- Status: clear
- Last Reviewed: 2026-04-17
- Reviewer: codex
- Command Skill: `$namba-plan-design-review`
- Recommended Role: `namba-designer`

## Focus

- Clarify CLI interaction states, decision boundaries, and operator-facing guidance before implementation starts.

## Review Checklist

- The command makes the workspace decision legible before files are written.
- Automatic worktree creation and explicit in-place override are easy to distinguish in terminal output.
- Dirty or ambiguous states explain why execution stopped and what the operator must do next.
- The resulting SPEC id, branch, and worktree path are reported in a scannable way.
- Bundled plan-review guidance preserves the same mental model instead of describing stale behavior.

## Findings

- 2026-04-17: The core design surface is state communication, not visuals. This SPEC is strongest when treated as a three-state interaction model: `reuse current isolated workspace`, `create isolated workspace`, or `stop because the context is unsafe`. That structure is implied in `spec.md` and `acceptance.md`, but the output contract should name these states explicitly so the operator never has to infer what Namba decided.
- 2026-04-17: The highest UX risk is override ambiguity. "Scaffold here" is mentioned as an escape hatch, but the SPEC does not yet require a distinct acknowledgement pattern that differentiates intentional override from default behavior. The command output should visibly mark this path as an exception, not as a normal success variant.
- 2026-04-17: Auto-created worktree behavior needs a stronger next-step narrative. Reporting only the created path is not enough; the operator should see whether Namba already scaffolded in that worktree, whether they must `cd` first, and which branch now owns the new SPEC. Without that, the workflow remains technically correct but operationally fuzzy.
- 2026-04-17: Dirty-workspace refusal is correctly in scope, but the design requirement should also cover tone and resolution. The refusal message should separate `why blocked`, `what Namba did not do`, and `how to proceed` so it feels protective instead of arbitrary.
- 2026-04-17: Plan-review alignment is correctly called out, but the review bundle needs one shared phrase for the new mental model. If `namba plan` says "planning starts in an isolated workspace" while `$namba-plan-review` still reads like "create a SPEC here, then review it," users will keep mispredicting the flow even after the runtime behavior is fixed.

## Decisions

- No visual redesign is needed; this is a CLI/interaction-contract review.
- Treat workspace resolution as a first-class UX event with explicit operator messaging before and after scaffold creation.
- Keep the interaction terse, but require one consistent summary block that reports `SPEC id`, `workspace action`, `branch`, `path`, and `next step`.
- Preserve advisory tone for refusals and overrides: explain intent, avoid blame, and avoid hidden automation.

## Follow-ups

- Add concrete example outputs to implementation or help-text work before coding completes, especially for:
  - shared/base workspace invocation that creates a worktree
  - dedicated worktree reuse
  - explicit current-workspace override
  - dirty-workspace refusal
- Verify that help text and `$namba-plan-review` guidance use the same language for the isolation contract and do not imply in-place scaffolding as the default.
- During implementation review, check that success and refusal messages remain easy to scan in a narrow terminal without burying the decisive line below explanatory prose.

## Recommendation

- Clear to proceed from a design/UX communication perspective. No blocking design issue remains, provided the implementation treats workspace classification and override/refusal messaging as the primary user experience, not as secondary logging.
