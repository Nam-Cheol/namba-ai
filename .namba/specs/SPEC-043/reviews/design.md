# Design Review

- Status: clear
- Last Reviewed: 2026-05-13
- Reviewer: Codex
- Command Skill: `$namba-plan-design-review`
- Recommended Role: `namba-designer`

## Focus

- Clarify art direction, palette discipline, anti-generic composition, purposeful motion, and visual risks before implementation starts.

- Evidence Status: not-applicable
- Gate Decision: not-applicable
- Approved Direction: Reusable documentation blueprint per manual family, borrowing Ouroboros-style readability patterns while keeping NambaAI practical and workflow-oriented.
- Banned Patterns: Decorative-only badges; duplicated one-note card grids; unsupported product claims; one-locale-only polish; GitHub-hostile HTML; dense emoji that hurts scanning.
- Open Questions: Whether supporting reference docs need a full visual pass or only top navigation; whether README output should change through sync-owned sources.
- Unresolved Questions: None blocking implementation.
- Design Review Axes: evidence, assets, alternatives, hierarchy, craft, functionality, differentiation
- Keep / Fix / Quick Wins: pending

## Review Checklist

- Art direction is clear and fits the task context.
- Palette temperature and undertone logic are coherent, saturation stays restrained, and the result does not collapse into washed-out gray minimalism.
- Semantic components and layout primitives match the content instead of defaulting to generic cards, border-heavy framing, or bento/grid fallback.
- Motion, if proposed, has a concrete hierarchy, attention, or state-change purpose.
- The most generic section is redesigned when the task is page-, screen-, or section-scale; component-scale tasks call out the risk without gratuitous scope creep.
- Anti-overcorrection guardrails hold: no novelty for novelty's sake, no decorative asymmetry without payoff, and no loss of accessibility, design-system fit, or implementation realism.

## Findings

- The reference provides a clear readability vocabulary: language links, first-screen identity, compact navigation, quick-start blocks, command tables, and strong section breaks.
- NambaAI should adapt the structure rather than the personality. The docs should feel like an operator manual with a polished command map, not a manifesto page.
- GitHub Markdown constraints matter more than visual novelty. Tables, concise callouts, and simple HTML are enough.
- The biggest design failure mode is adding visual furniture that does not answer a reader's next decision.

## Decisions

- Use a "command cockpit" information architecture: orient, start, choose command, understand workflow, continue deeper.
- Keep badges/status links sparse and factual.
- Use equivalent section order across locales so readers can switch languages without losing place.

## Follow-ups

- During implementation, preview or inspect rendered Markdown for table width, anchor behavior, and language link correctness.
- Keep Japanese and Chinese headings natural rather than forcing English table labels too literally.

## Recommendation

- Proceed with the blueprint direction. Design review should be revisited only if implementation introduces heavy HTML, new imagery, or a substantially different document structure.
