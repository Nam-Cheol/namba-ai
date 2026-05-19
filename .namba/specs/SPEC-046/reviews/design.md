# Design Review

- Status: approved
- Last Reviewed: 2026-05-19
- Reviewer: Codex as `namba-designer`
- Command Skill: `$namba-plan-design-review`
- Recommended Role: `namba-designer`

## Focus

- Clarify art direction, palette discipline, anti-generic composition, purposeful motion, and visual risks before implementation starts.

- Evidence Status: not-applicable
- Gate Decision: not-applicable
- Approved Direction: CLI and documentation language only; no frontend visual surface is in scope.
- Banned Patterns: Do not describe normal PR or queue handoff as automatically requesting Codex review.
- Negative-First Contract: not-applicable
- Default Library Fit: not-applicable
- Context-Specific Bans And Replacements: Replace default-review wording with explicit opt-in wording.
- Reference-Driven Asset Manifest: not-applicable
- Generated Image Plan: not-applicable
- Visual Grammar: not-applicable
- Generic-Section Proof: not-applicable
- Architecture Handoff: Documentation and help text should use parallel phrasing for direct PR and queue paths.
- Violation-Check Plan: Search generated docs and skills for stale automatic review language.
- Open Questions: none
- Unresolved Questions: none
- Design Review Axes: evidence, assets, alternatives, hierarchy, craft, functionality, differentiation
- Keep / Fix / Quick Wins: Keep concise CLI wording; fix stale review-marker promises.

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

- This is not a frontend or visual-design SPEC. The only design-sensitive area is command and documentation wording.
- The current generated language over-promises review marker creation during normal handoff. That should become explicit opt-in language.
- Queue and PR docs should mirror each other so users do not have to learn two review request contracts.

## Decisions

- No image, visual asset, screen, or prototype work is required.
- Keep wording terse and operational: normal handoff versus explicit Codex review request.

## Follow-ups

- During implementation, scan localized docs for stale wording that says `namba pr` guarantees a Codex review marker.

## Recommendation

- Approved. Frontend gate is not applicable.
