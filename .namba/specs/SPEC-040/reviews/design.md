# Design Review

- Status: clear
- Last Reviewed: 2026-04-27
- Reviewer: Codex acting as `namba-designer`
- Command Skill: `$namba-plan-design-review`
- Recommended Role: `namba-designer`

## Focus

- Clarify art direction, palette discipline, anti-generic composition, purposeful motion, and visual risks before implementation starts.

- Evidence Status: complete
- Gate Decision: approved
- Approved Direction: Concise operator-facing command-entry guidance with progressive disclosure for long procedures and evidence-first validation reports.
- Banned Patterns: dense `SKILL.md` manuals, new standalone skills for imported ideas, install-flow instructions, generic frontend validation claims without rendered-state evidence, and decorative guidance unrelated to command execution.
- Open Questions: None blocking.
- Unresolved Questions: None.
- Design Review Axes: evidence, assets, alternatives, hierarchy, craft, functionality, differentiation
- Keep / Fix / Quick Wins: Keep the compact command-entry shape; fix any implementation diff that turns skills into long procedural manuals; quick win is adding clear evidence labels for PR check, review-thread, screenshot, DOM, and console outputs.

## Review Checklist

- Art direction is clear and fits the task context.
- Palette temperature and undertone logic are coherent, saturation stays restrained, and the result does not collapse into washed-out gray minimalism.
- Semantic components and layout primitives match the content instead of defaulting to generic cards, border-heavy framing, or bento/grid fallback.
- Motion, if proposed, has a concrete hierarchy, attention, or state-change purpose.
- The most generic section is redesigned when the task is page-, screen-, or section-scale; component-scale tasks call out the risk without gratuitous scope creep.
- Anti-overcorrection guardrails hold: no novelty for novelty's sake, no decorative asymmetry without payoff, and no loss of accessibility, design-system fit, or implementation realism.

## Findings

- This SPEC is not a visual UI slice, so the design risk is operator comprehension rather than palette, layout, or motion. The relevant design surface is the shape of command-entry instructions, PR/review evidence, and frontend validation reports.
- The progressive-disclosure direction is strong. It protects the first-load skill experience while still allowing detailed procedures to exist behind references or deterministic helpers.
- The proposed frontend validation evidence is useful and concrete: server lifecycle, rendered-state inspection, screenshots, DOM facts, and console logs. That is a better operator experience than relying on prose claims that a UI was "checked."
- The `namba pr` and `$namba-review-resolve` evidence requirements should improve review handoff clarity if implemented with compact labels and bounded snippets.
- The design checklist scaffold includes visual concerns that are mostly not applicable here; the SPEC correctly narrows actual acceptance around command UX and evidence presentation.

## Decisions

- Treat "design quality" for this slice as clarity, scanability, and evidence hierarchy in command guidance and validation output.
- Keep command-entry skills short and structured. Long procedures should become references or helpers only when they reduce repeated manual work.
- Frontend validation outputs should prefer named evidence artifacts or concise labels over free-form narrative.

## Follow-ups

- During implementation, check generated skill text for overlong bullets and repeated concepts across skills.
- If frontend helper guidance is added, ensure screenshot, DOM, and console evidence are named consistently enough to appear in PR summaries.
- Keep visual-design language out of non-frontend command-entry skills unless a command genuinely owns a UI or browser validation outcome.

## Recommendation

- Clear for implementation. No design blocker remains; the remaining quality bar is concise, evidence-first operator guidance.
