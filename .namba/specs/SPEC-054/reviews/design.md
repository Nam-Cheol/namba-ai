# Design Review

- Status: approved
- Last Reviewed: 2026-05-20
- Reviewer: Codex (`namba-designer` lens)
- Command Skill: `$namba-plan-design-review`
- Recommended Role: `namba-designer`

## Focus

- Clarify art direction, palette discipline, anti-generic composition, purposeful motion, and visual risks before implementation starts.

- Evidence Status: `spec.md`, `plan.md`, `acceptance.md`, and current PR and release code surfaces reviewed.
- Gate Decision: advisory proceed with workflow-writing guardrails.
- Approved Direction: improve textual handoff UX; no visual UI or frontend design work is required.
- Banned Patterns: empty headings, placeholder bullets, "see checklist" as the whole summary, release bodies that only announce a tag, and unsolicited review-request text.
- Negative-First Contract: do not generate generic prose, do not hide validation behind vague words like "validated", do not imply Codex review was requested without explicit opt-in, and do not override configured language.
- Default Library Fit: not applicable.
- Context-Specific Bans And Replacements: replace checklist-only PR summaries with completed work, changes, evidence, and validation sections; replace generic release text with user-visible changes plus source evidence.
- Reference-Driven Asset Manifest: not applicable.
- Generated Image Plan: not applicable.
- Visual Grammar: concise engineering handoff prose with stable sections and evidence links.
- Generic-Section Proof: acceptance now requires non-empty, non-generic PR and release bodies with evidence-bearing sections.
- Architecture Handoff: align wording changes with `templates.go`, `buildPullRequestBody`, and `renderReleaseNotes`.
- Violation-Check Plan: scan generated artifacts in tests for required sections, evidence references, validation content, language selection, and absence of unrequested `@codex review`.
- Open Questions: none blocking.
- Unresolved Questions: none blocking.
- Design Review Axes: evidence, assets, alternatives, hierarchy, craft, functionality, differentiation
- Keep / Fix / Quick Wins: keep concise handoff structure; fix placeholder-style wording; add fixture-backed language and evidence checks.

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

1. Medium: This is not a visual surface, but it is still user experience.
   Reviewers need a PR body and release body that can be understood without
   opening every linked file.
2. Medium: The strongest anti-generic move is a stable writing structure:
   completed work, changes, evidence, and validation. The SPEC now encodes
   that structure.
3. Low: Release notes should remain release-facing, not become an internal
   execution log. Evidence should be concise and useful, such as SPEC IDs, PR
   numbers, short hashes, and validation summaries.

## Decisions

- Treat artifact prose as workflow UX.
- No frontend, image, palette, motion, or layout work is in scope.
- Use tests to prevent generic placeholder text from returning.

## Follow-ups

- Implementation review should inspect generated PR and release bodies as
  rendered text, not just individual template strings.

## Recommendation

- Proceed. The design risk is bounded to prose quality and evidence clarity,
  and the SPEC now gives implementers a clear anti-generic contract.
