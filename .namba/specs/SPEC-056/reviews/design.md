# Design Review

- Status: clear
- Last Reviewed: 2026-05-20
- Reviewer: Codex local review
- Command Skill: `$namba-plan-design-review`
- Recommended Role: `namba-designer`

## Focus

- Clarify art direction, palette discipline, anti-generic composition, purposeful motion, and visual risks before implementation starts.

- Evidence Status: not applicable; CLI report design only
- Gate Decision: clear
- Approved Direction: concise terminal Markdown and stable JSON
- Banned Patterns: decorative dashboard language, noisy tables without failure-first triage, color-dependent status semantics
- Negative-First Contract: failures and regressions appear before passing scenario detail
- Default Library Fit: plain text and JSON only
- Context-Specific Bans And Replacements: no browser UI, no generated images, no charting dependency
- Reference-Driven Asset Manifest: not applicable
- Generated Image Plan: not applicable
- Visual Grammar: summary, metric score, failure diagnostics, scenario appendix
- Generic-Section Proof: report sections are tied to eval failure analysis, not generic marketing copy
- Architecture Handoff: renderer split in `eval_render.go`
- Violation-Check Plan: command tests assert failed scenarios include expected versus actual fields
- Open Questions: none
- Unresolved Questions: none
- Design Review Axes: evidence, assets, alternatives, hierarchy, craft, functionality, differentiation
- Keep / Fix / Quick Wins: pending

## Review Checklist

- Art direction is clear and fits the task context.
- Palette temperature and undertone logic are coherent, saturation stays restrained, and the result does not collapse into washed-out gray minimalism.
- Semantic components and layout primitives match the content instead of defaulting to generic cards, border-heavy framing, or bento/grid fallback.
- The Do-Not Design Contract is complete enough to block generic fallback, including default-library fit, context-specific bans, allowed replacements, brand/category/trust reasoning, visual grammar, reference-driven asset manifest, generated-image plan, architecture handoff, and violation-check plan.
- Asset-led references define concrete image assets, generation prompts, output paths, and rendered usage evidence instead of allowing brand-color-only imitation or placeholder media.
- Motion, if proposed, has a concrete hierarchy, attention, or state-change purpose.
- The most generic section is redesigned when the task is page-, screen-, or section-scale; component-scale tasks call out the risk without gratuitous scope creep.
- Anti-overcorrection guardrails hold: no novelty for novelty's sake, no decorative asymmetry without payoff, and no loss of accessibility, design-system fit, or implementation realism.

## Findings

- This is a CLI and report-design feature, not a visual app surface.
- The important design requirement is failure-first information architecture: maintainers should see what regressed, why it failed, and which fixture owns the expectation.
- Markdown and JSON outputs should share the same result model to avoid two divergent report contracts.

## Decisions

- Do not add UI assets, browser flows, charting libraries, or generated imagery.
- Keep Markdown readable in plain terminals and GitHub logs.
- Keep JSON stable enough for CI and future automation.

## Follow-ups

- Add snapshot-light assertions for report section headings and failed-scenario fields rather than brittle full-output snapshots.

## Recommendation

- Clear for implementation.
