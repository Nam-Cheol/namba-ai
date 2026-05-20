# Design Review

- Status: approved
- Last Reviewed: 2026-05-20
- Reviewer: Codex (`namba-designer`)
- Command Skill: `$namba-plan-design-review`
- Recommended Role: `namba-designer`

## Focus

- Review the SPEC from a documentation UX and CLI evidence presentation
  perspective before implementation.

- Evidence Status: advisory review complete
- Gate Decision: proceed with implementation after wording and presentation guardrails below are adopted
- Approved Direction: expose Codex diagnostics as inspectable evidence, not as product-health alarms
- Banned Patterns: red danger framing for optional diagnostics, ambiguous "failed" summaries without scope, README examples that hide where logs live, KR/EN terminology drift
- Negative-First Contract: warn only when user action is plausibly needed; otherwise label missing Codex or mismatches as advisory/unavailable/non-blocking
- Default Library Fit: CLI tables, JSON snippets, and README callouts are appropriate if they stay compact and evidence-first
- Context-Specific Bans And Replacements: ban "broken / invalid environment" copy for absence or mismatch; replace with "not detected", "outside configured roots", or "advisory mismatch"
- Reference-Driven Asset Manifest: none required; use deterministic text examples instead of decorative assets
- Generated Image Plan: not applicable
- Visual Grammar: concise status labels, progressive disclosure from summary -> paths -> raw evidence, consistent field names across CLI, JSON, README, and generated contract docs
- Generic-Section Proof: the riskiest generic section is README evidence guidance; it should show exact inspection commands and sample file paths instead of narrative-only explanation
- Architecture Handoff: implementation should centralize status vocabulary so README/CLI/generated docs do not fork in tone
- Violation-Check Plan: verify every surfaced status against three questions: is it blocking, who should act, and where is the evidence file
- Open Questions: none blocking
- Unresolved Questions: exact final status taxonomy remains for implementation to settle
- Design Review Axes: evidence clarity, inspectability, warning tone, bilingual parity, status semantics, fallback comprehension
- Keep / Fix / Quick Wins: keep optional/non-blocking direction; fix wording severity; add exact evidence inspection paths and commands early in docs

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

1. Warning severity is under-specified for an optional feature. The SPEC and
   acceptance correctly say Codex evidence is non-blocking, but the plan leaves
   room for CLI or README copy that could overstate missing Codex, doctor
   failures, or workspace-root mismatches as health errors. The implementation
   should reserve alarming language for real execution blockers and keep all
   other statuses explicitly advisory.
2. README inspectability needs stronger emphasis than "explain how to inspect".
   For this feature, users will need exact paths, command examples, and a clear
   sequence for moving from summary status to persisted doctor logs and evidence
   JSON. Without that, the feature will feel opaque even if the data exists.
3. Korean/English documentation parity is a real UX risk because status terms
   such as "missing", "unavailable", "mismatch", "timeout", and "advisory"
   carry different severity if translated loosely. The SPEC should be
   implemented with a stable status vocabulary that both READMEs reuse.
4. Workspace-root mismatch messaging is especially easy to over-alarm. A
   mismatch should explain what was compared and why it matters, but it should
   not imply corruption or unsafe state when the workflow is intentionally
   non-blocking.

## Decisions

- Approve the evidence-first direction.
- Treat this as information architecture work, not visual design work.
- Require summary-first presentation:
  1. status
  2. why it has that status
  3. where the evidence lives
  4. what action, if any, is recommended
- Prefer neutral status words such as `detected`, `not detected`,
  `unavailable`, `timed out`, `advisory mismatch`, and `redacted`.
- Keep README examples compact and inspection-oriented; do not bury file paths
  or commands in long explanatory prose.

## Follow-ups

- Add an implementation note to keep one shared status taxonomy for CLI output,
  evidence JSON, README.md, README.ko.md, and generated contract docs.
- In both READMEs, include at least one exact example of:
  - the summary surface users will see
  - the evidence JSON location
  - the persisted doctor log path
  - the command or workflow a user should run next when they want details
- Ensure timeout/failure copy distinguishes between:
  - workflow success with partial diagnostics
  - workflow failure caused by something else
- For workspace-root mismatch text, include both compared values when safe and
  end with explicit non-blocking guidance.

## Recommendation

- Recommendation: approve with advisory adjustments. This SPEC is ready for
  implementation if documentation and CLI surfaces are written to minimize false
  alarm, maximize evidence inspectability, and preserve Korean/English wording
  parity.
