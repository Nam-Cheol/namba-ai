# Design Review

- Status: approved
- Last Reviewed: 2026-04-01
- Reviewer: Codex
- Command Skill: `$namba-plan-design-review`
- Recommended Role: `namba-designer`

## Focus

- Clarify interaction quality, responsive states, accessibility, and visual direction before implementation starts.

## Findings

- The primary design surface in this SPEC is terminal UX plus generated documentation, not graphical UI. The important interaction question is whether users can immediately distinguish feature planning, bugfix planning, and direct repair.
- The `fix --command plan|run` model is easier to teach than moving bugfix planning into another command family. It keeps bug work discoverable under `fix` while still making planning-versus-execution explicit.
- The highest-value design requirement is copy clarity. Help text, redirect text, README bundles, and skill descriptions need to explain the same mental model with compact examples instead of internal jargon such as "harness-oriented repair entrypoint."
- Explaining why each user-facing repo-local skill exists is worthwhile, but the docs should stay example-led and scannable rather than becoming an encyclopedia of internal implementation details.

## Decisions

- Keep the user-facing mental model explicit in every surface: `plan` authors feature work, `fix --command plan` authors bugfix work, and `fix` or `fix --command run` performs a direct repair in the current workspace.
- Keep help output compact, non-interactive, and example-driven. The key examples are `namba plan "..."`, `namba fix --command plan "..."`, and `namba fix "..."`.
- Keep migration copy action-led. When users try the old mental model, the CLI should redirect them clearly to `namba fix --command plan` for reviewable SPEC creation.
- Keep README and generated-doc structure parallel across supported languages and across the user-facing skill list so intent, command mapping, and option details are easy to locate.

## Follow-ups

- During implementation, define one stable wording pattern for help, redirect, success, and failure states so the new contract is easy to scan in plain terminals.
- Explain each user-facing skill in terms of intent first, then command mapping, then options only where they materially change behavior.
- Avoid color-only or decoration-heavy terminal output; concise text lines are easier to read, test, and localize.

## Recommendation

- Advisory recommendation: approved. Proceed with implementation using concise, example-led CLI copy and aligned multilingual generated documentation.
