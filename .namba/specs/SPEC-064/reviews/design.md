# Design Review

- Status: clear
- Last Reviewed: 2026-05-22
- Reviewer: `namba-designer`
- Command Skill: `$namba-plan-design-review`
- Recommended Role: `namba-designer`

## Focus

- Clarify art direction, palette discipline, anti-generic composition, purposeful motion, and visual risks before implementation starts.

- Evidence Status: not-applicable
- Gate Decision: not-applicable
- Approved Direction: language-first entry, compact progress rail, repo-intelligence header after language selection, purpose-first setup paths, Codex access consequence preview, Git guidance, and ASCII/plain fallback as first-class output
- Banned Patterns: country flags, full-screen TUI, alternate-screen takeover, box-drawing dashboards, wide bordered cards, color-only meaning, emoji-only meaning, long marketing copy before first choice, multi-column terminal layout, raw-key-only interaction
- Negative-First Contract: not-applicable; this is CLI terminal UX and does not require generated image assets
- Default Library Fit: use existing CLI prompt helpers and small rendering helpers rather than introducing a new TUI framework
- Context-Specific Bans And Replacements: replace country flags with text-first language codes; replace visual-only state with textual markers; replace fullscreen framing with compact progress and handoff sections
- Reference-Driven Asset Manifest: not-applicable for terminal text UI
- Generated Image Plan: not-applicable
- Visual Grammar: compact line-oriented terminal hierarchy, optional safe emoji decoration, text-complete semantics, no mandatory full-screen behavior
- Generic-Section Proof: the most generic surface is the current banner/step header; replacement is a purpose-first language entry plus repo-intelligence header and final handoff card
- Architecture Handoff: implement as CLI rendering/copy helpers under existing init wizard code, with tests for styled and plain output
- Violation-Check Plan: assert no country flag emoji in language labels, plain output has no ANSI reliance, and raw-key failure falls back to line prompts
- Open Questions: exact progress-rail text format and conservative terminal-capability cutoff
- Unresolved Questions: none blocking; both open questions are implementation-level choices covered by acceptance
- Design Review Axes: evidence, terminal resilience, alternatives, hierarchy, craft, functionality, differentiation
- Keep / Fix / Quick Wins: keep Codex access preview semantics; fix pre-language comprehension and flag labels; add compact progress and final next-step handoff

## Review Checklist

- Art direction is clear and fits the task context.
- Palette and saturation concerns are not applicable to plain CLI output; color must remain optional.
- Semantic terminal primitives match the content instead of defaulting to decorative framing.
- The design contract blocks country flags, full-screen takeover, color-only meaning, and raw-key-only interaction.
- Asset-led image references are not needed because this is not a browser or bitmap UI.
- Motion is not needed; progress should be structural text, not animation.
- The generic current banner/step header is replaced by a language-first entry and repo-aware handoff structure.
- Anti-overcorrection guardrails hold: no novelty for novelty's sake, no loss of accessibility, and no new TUI dependency unless clearly justified.

## Findings

- The current flow starts with Korean copy before language selection.
- Current language labels use country-flag emoji.
- Repo-intelligence is valuable after language selection, not before the user understands the wizard.
- Plain/ASCII mode must be treated as an equal output path rather than a degraded fallback.
- Progress should communicate orientation without becoming a full-screen terminal app.

## Decisions

- Keep this as a disciplined CLI wizard, not a terminal application.
- Require every visual meaning to have a text equivalent.
- Use recommended badges only when they reflect actual repo-state defaults.

## Follow-ups

- Verify styled and plain outputs preserve the same information architecture.
- Verify labels like `[ko] Korean`, `[en] English`, `[ja] Japanese`, and `[zh] Simplified Chinese`.
- Verify compact header fields only: path, repo state, detected stack, methodology, and defaults.
- Verify no fullscreen redraw, decorative chrome, or country flag emoji.

## Recommendation

- Cleared for implementation.
