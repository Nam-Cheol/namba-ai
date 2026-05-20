# Design Review

- Status: approved
- Last Reviewed: 2026-05-20
- Reviewer: Codex (`namba-designer`)
- Command Skill: `$namba-plan-design-review`
- Recommended Role: `namba-designer`

## Focus

- Clarify art direction, palette discipline, anti-generic composition, purposeful motion, and visual risks before implementation starts.

- Evidence Status: `spec.md`, `plan.md`, and `acceptance.md` reviewed; `SPEC-050` and `SPEC-052` summaries checked for boundary alignment.
- Gate Decision: advisory proceed with documentation and workflow-language guardrails.
- Approved Direction: Documentation and workflow UX clarity only; no visual UI redesign is required.
- Banned Patterns: readiness dashboards, maturity scorecards, decorative platform framing, and gate-style ceremonies in front of normal local NambaAI usage.
- Negative-First Contract: Do not imply plugin installation is required, remote-control is expected for normal use, Namba owns Codex mention resolution, or remote readiness status equals workflow failure.
- Default Library Fit: not applicable; no frontend component or visual system work is in scope.
- Context-Specific Bans And Replacements: replace feature-matrix-first explanation with local workflow first, command routing second, optional readiness third.
- Reference-Driven Asset Manifest: not applicable.
- Generated Image Plan: not applicable.
- Visual Grammar: documentation should use contrastive wording such as "Codex may expose X; Namba documents or observes X; Namba does not require X for normal local workflows."
- Generic-Section Proof: docs should include one compact routing/glossary note rather than repeated caveats.
- Architecture Handoff: align docs with the shared `codex_diagnostics` status strategy from engineering review.
- Violation-Check Plan: implementation review should scan README, Korean README, and generated docs for required/implied plugin or remote setup language.
- Open Questions: none blocking.
- Unresolved Questions: none blocking.
- Design Review Axes: evidence, assets, alternatives, hierarchy, craft, functionality, differentiation
- Keep / Fix / Quick Wins: keep local-first framing; fix concept order; add compact glossary; avoid ceremony.

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

1. Medium: The biggest UX risk is conceptual overload between explicit Namba
   command skills, Codex unified `@` mentions, plugin metadata, and file or
   directory references. Without a fixed explanation order, users may conclude
   that Namba itself parses `@` mentions or treats mention search as the primary
   workflow router.
2. Medium: Remote-control and remote-environment readiness can read like
   activation prerequisites unless neutral status language makes
   `unavailable`, `disabled`, `configured`, and `local_fallback` clearly
   non-failure states.
3. Medium: The acceptance set is technically strong, but docs need a simpler
   mental model: default local workflow, optional Codex platform readiness, and
   evidence-only status surfaces.
4. Low: This SPEC has no end-user visual UI. It should not compensate with
   dashboard, scorecard, or frontend-style gate framing.
5. Low: Plugin packaging and sharing language needs care so marketplace, share,
   and default-enabled hook language does not sound like required setup.

## Decisions

- Treat this as documentation information architecture and workflow
  comprehension work, not visual design work.
- Use the documentation order now recorded in `spec.md`: local command flow,
  explicit skill routing, unified mentions, optional plugin/share paths,
  optional remote status evidence, and SDK rename where relevant.
- Keep examples inspection-first and minimal.

## Follow-ups

- Add a compact glossary or command-routing note covering `@` mentions, Namba
  command skills, plugins, remote-control, remote environments, and evidence
  statuses.
- During implementation review, scan user-facing docs for sentences that imply
  plugin install, remote-control, Codex mention ownership, or readiness failure
  semantics.

## Recommendation

- Proceed. The SPEC is directionally sound after the documentation-order and
  negative-first workflow guardrails were added.
