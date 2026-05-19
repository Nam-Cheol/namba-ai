# Design Review

- Status: clear
- Last Reviewed: 2026-05-19
- Reviewer: namba-designer
- Command Skill: `$namba-plan-design-review`
- Recommended Role: `namba-designer`

## Focus

- Clarify art direction, palette discipline, anti-generic composition, purposeful motion, and visual risks before implementation starts.

- Evidence Status: spec, plan, and acceptance define enough operator-facing
  behavior and documentation scope to proceed.
- Gate Decision: advisory clear
- Approved Direction: installer output and docs should be concise, specific,
  and action-oriented: say what failed, what asset was checked, and whether
  installation stopped before extraction.
- Banned Patterns: vague "verification failed" messages, security theater copy,
  long README warnings above quick-start commands, hidden test-only knobs, and
  success copy that implies signing or provenance exists.
- Negative-First Contract: do not introduce a visual or marketing layer; this is
  a CLI/documentation trust-boundary change.
- Default Library Fit: not applicable; no UI component library is involved.
- Context-Specific Bans And Replacements: replace generic failure text with
  precise states such as missing checksum, checksum mismatch, unavailable hash
  tool, missing expected binary, or archive extraction failure.
- Reference-Driven Asset Manifest: not applicable.
- Generated Image Plan: not applicable.
- Visual Grammar: terminal and README text only; prefer short headings,
  monospace command names, and compact install notes.
- Generic-Section Proof: README quick start stays concise while checksum
  verification appears in install or release-integrity notes.
- Architecture Handoff: implementation should surface clear error strings that
  tests can assert without depending on decorative output.
- Violation-Check Plan: scan docs for stale language that implies unchecked
  installation or overclaims signing/provenance.
- Open Questions: none blocking.
- Unresolved Questions: none blocking.
- Design Review Axes: evidence, assets, alternatives, hierarchy, craft, functionality, differentiation
- Keep / Fix / Quick Wins: keep quick-start commands concise; fix failure copy
  precision; add manual checksum note without moving it above first-run install
  commands.

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

- This SPEC is mostly non-visual. The relevant design surface is operator
  comprehension in terminal output and install documentation.
- The failure-mode vocabulary is the main UX risk. Users should know whether
  the problem is a missing artifact, missing hash tool, mismatched digest, or
  invalid archive contents.
- The README direction avoids overcorrection: it improves trust communication
  without turning a quick-start flow into a security whitepaper.
- Explicitly disallowing skip flags prevents confusing product copy where users
  see a prominent bypass next to the new trust guarantee.

## Decisions

- Keep installer messaging short, direct, and state-specific.
- Keep quick-start commands prominent; place checksum explanation in nearby
  install/security notes.
- Do not imply binary signing, provenance, or package-manager assurance in docs
  for this SPEC.

## Follow-ups

- [non-blocking] During implementation, make tests assert at least one
  state-specific failure phrase for mismatch and missing checksum.
- [post-implementation] Consider a separate docs pass if future signing or
  provenance changes require a fuller trust model explanation.

## Recommendation

- Clear to proceed. No visual design blocker exists; the operator-facing copy
  guardrails are enough for implementation.
