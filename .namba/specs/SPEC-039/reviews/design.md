# Design Review

- Status: advisory-pass
- Last Reviewed: 2026-04-27
- Reviewer: namba-designer
- Command Skill: `$namba-plan-design-review`
- Recommended Role: `namba-designer`

## Focus

- Clarify the README visual language: emoji should be semantic scan markers for Namba workflow concepts, not decorative filler.
- Keep the generated README professional for a CLI/tooling audience while making command routing and lifecycle stages easier to skim.
- Check multilingual headings so emoji placement does not make English, Korean, Japanese, or Chinese output feel inconsistent or cluttered.
- Preserve accessibility and copy clarity: command names and links must remain text-first and screen-reader-friendly.

- Evidence Status: reviewed against `.namba/specs/SPEC-039/spec.md`, `plan.md`, `acceptance.md`, and current generated `README*.md` outputs.
- Gate Decision: pass with guardrails
- Approved Direction: use a small, repeated emoji vocabulary at section-heading level plus a few high-value bullets for lifecycle cues; keep command names, links, and code blocks emoji-free.
- Banned Patterns: emoji on every bullet, emoji replacing section nouns, decorative marker changes that differ by locale, badge/link lines mixed with emoji, and acceptance tests that snapshot full translated files instead of checking durable section anchors.
- Open Questions: whether the same emoji set should be shared across all locales verbatim, and whether `$namba-review-resolve` / `$namba-release` should receive dedicated section markers or inherit existing workflow markers.
- Unresolved Questions: none blocking implementation if the guardrails below are adopted before renderer changes.
- Design Review Axes: evidence, assets, alternatives, hierarchy, craft, functionality, differentiation
- Keep / Fix / Quick Wins: keep the current text-first README structure and hook-runtime precedent; fix the lack of explicit emoji density limits; quick win is to define stable heading anchors before implementation and test them across all four README outputs.

## Review Checklist

- Art direction is clear and fits the task context.
- Palette temperature and undertone logic are coherent, saturation stays restrained, and the result does not collapse into washed-out gray minimalism.
- Semantic components and layout primitives match the content instead of defaulting to generic cards, border-heavy framing, or bento/grid fallback.
- Motion, if proposed, has a concrete hierarchy, attention, or state-change purpose.
- The most generic section is redesigned when the task is page-, screen-, or section-scale; component-scale tasks call out the risk without gratuitous scope creep.
- Anti-overcorrection guardrails hold: no novelty for novelty's sake, no decorative asymmetry without payoff, and no loss of accessibility, design-system fit, or implementation realism.

## Findings

- The current generated README is professional and readable because most sections are plain-language, command-first, and visually restrained. That is the right baseline to preserve. The only existing high-density marker zone is `Hook Runtime`, which works because the emoji vocabulary is semantic and scoped to a single explanatory section.
- The SPEC goal is directionally correct, but it currently leaves too much room for emoji spread. Without an explicit density rule, implementers could decorate command lists, link rows, and routine bullets, which would make the CLI README feel less like tooling documentation and more like marketing copy.
- Multilingual readability is the main design risk. English, Korean, Japanese, and Chinese all have different line rhythm and punctuation density; if emoji are inserted inconsistently or attached to inline commands, the translated READMEs will lose parallel structure and become harder to scan side-by-side.
- Accessibility needs to be treated as a content-system rule, not just a copy preference. Screen readers and skim readers both benefit when emoji stay at predictable heading or cue positions and do not interrupt command names, file paths, or badge/link clusters.
- The acceptance criteria already move in the right direction by calling for stable anchors instead of brittle snapshots. That should be tightened into a visual acceptance rule: headings and command-skill sections need durable semantic landmarks that survive wording changes and locale differences.

## Decisions

- Approve the README visual upgrade only if emoji usage stays constrained to section headings and a short list of repeated workflow cues such as command choice, planning, execution, review, release, install/update, and technical snapshot.
- Preserve the current CLI-tooling tone. The README should read like operational guidance with clearer hierarchy, not like a promotional landing page.
- Require locale-stable structure: the same sections should receive the same semantic treatment in English, Korean, Japanese, and Chinese even when the wording differs.
- Treat command literals, language links, release/CI/security links, and shell snippets as text-only zones.
- Use acceptance/tests to pin section-level anchors, not exact bullet phrasing or full-file emoji layout.

## Follow-ups

- Define a small emoji vocabulary before implementation, for example command choice, planning, execution, review, release, install/update, docs, and technical snapshot.
- Decide the density cap explicitly in the renderer or review notes: headings by default, selected bullets only when they describe lifecycle state or caution.
- Verify generated README output visually after `namba sync`, especially `README.md`, `README.ko.md`, `README.ja.md`, and `README.zh.md`, to confirm parallel section hierarchy.
- Add acceptance evidence for stable visual anchors such as `Which Command`, `Quick Start`, `Hook Runtime`, `Command Skills In Codex`, and `Technical Snapshot` rather than broad "visual upgrade" wording alone.

## Recommendation

- Proceed with implementation. The design direction is sound if the team treats emoji as restrained information scent, preserves text-first CLI professionalism, and validates multilingual heading consistency with stable section anchors.
