# Design Review

- Status: clear
- Last Reviewed: 2026-05-21
- Reviewer: Codex local design review
- Command Skill: `$namba-plan-design-review`
- Recommended Role: `namba-designer`

## Focus

- Clarify art direction, palette discipline, anti-generic composition, purposeful motion, and visual risks before implementation starts.

- Evidence Status: not-applicable
- Gate Decision: clear
- Approved Direction: CLI observability report with dense, scannable operator hierarchy.
- Banned Patterns: decorative dashboard language, marketing framing, noisy tables before decisions, and burying blockers below raw details.
- Negative-First Contract: not-applicable; this is a CLI/reporting surface, not a visual frontend.
- Default Library Fit: Go standard formatting and existing Namba CLI rendering patterns are sufficient.
- Context-Specific Bans And Replacements: avoid "all green" summaries when evidence is absent; replace with explicit `unknown`, `missing`, or `attention` states.
- Reference-Driven Asset Manifest: not-applicable.
- Generated Image Plan: not-applicable.
- Visual Grammar: lead with health, counts, blockers, and next action; keep detail sections grouped by runs, queue, specs, release, and diagnostics.
- Generic-Section Proof: the default report should not become a generic dump of JSON fields; it must explain why the workspace needs attention.
- Architecture Handoff: renderers should receive already-classified report data instead of parsing strings while rendering.
- Violation-Check Plan: tests should assert that human output includes top issues, blocked reasons, missing evidence, validation failures, review readiness, stale candidates, diagnostics, and next action.
- Open Questions: none blocking.
- Unresolved Questions: none.
- Design Review Axes: evidence, hierarchy, scanability, operator actionability, automation fit, accessibility through plain text.
- Keep / Fix / Quick Wins: keep report short by default; fix vague missing-state wording; quick win is a compact "Next" line that names the recovery command.

## Review Checklist

- Art direction is clear and fits the task context.
- Palette temperature and undertone logic are not applicable to this CLI surface.
- Semantic sections match the content instead of defaulting to a raw data dump.
- The Do-Not Design Contract is not applicable because this is not a visual implementation task.
- Asset-led references are not applicable.
- Motion is not applicable.
- The most generic section risk is addressed by requiring top issues and next action before detail.
- Anti-overcorrection guardrails hold: the report should stay utilitarian and not invent decorative dashboard concepts.

## Findings

- The user-facing design problem is information hierarchy, not visual styling.
- The report should answer "can I proceed?" and "what is blocking me?" before showing raw evidence paths.
- JSON output and human output need different shapes: JSON should be stable and complete; human output should be compact and diagnostic.

## Decisions

- Approve a compact operator report with health, top issues, counts, queue, evidence, readiness, diagnostics, and next action.
- Keep markdown/text output plain and terminal-friendly.
- Avoid in-app explanatory prose that teaches the entire system; the report should describe current state and recovery.

## Follow-ups

- During implementation, verify that the longest blocked reason and evidence paths remain readable in terminal output.
- Keep table usage optional; bullets are acceptable when they preserve scanability.

## Recommendation

- Proceed. Design risk is low and the CLI report hierarchy is sufficiently specified for implementation.
