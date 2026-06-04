# Design Review

- Status: cleared
- Last Reviewed: 2026-06-04
- Reviewer: Codex default reviewer
- Command Skill: `$namba-plan-design-review`
- Recommended Role: `namba-designer`

## Focus

- Clarify art direction, palette discipline, anti-generic composition, purposeful motion, and visual risks before implementation starts.

- Evidence Status: not-applicable
- Gate Decision: not-applicable
- Approved Direction: restrained operational CLI TUI
- Banned Patterns: decorative dashboards, marketing hero language, noisy color gradients, hiding final evidence inside AltScreen only
- Negative-First Contract: avoid inventing new setup steps; avoid status claims that are not backed by command state
- Default Library Fit: Bubble Tea v2 and Lip Gloss v2 are appropriate for a Go terminal UI
- Context-Specific Bans And Replacements: use concise status rows, focused choice lists, and final summaries instead of card-heavy layouts
- Reference-Driven Asset Manifest: not applicable
- Generated Image Plan: not applicable
- Visual Grammar: compact command wizard, clear active row, visible key hints, restrained success/warning/error styling
- Generic-Section Proof: not applicable to browser sections
- Architecture Handoff: TUI rendering must stay separate from domain init/update/regen logic
- Violation-Check Plan: verify non-TTY output remains readable without ANSI-dependent state
- Open Questions: none
- Unresolved Questions: none
- Design Review Axes: evidence, hierarchy, craft, functionality, differentiation
- Keep / Fix / Quick Wins: keep current command semantics; fix log-only interaction; add post-exit evidence summary

## Review Checklist

- Art direction is clear and fits an operational CLI.
- Palette and styling should be restrained and legible.
- Components should represent command state directly instead of generic panels.
- Motion should be limited to meaningful progress or state changes.
- Accessibility and non-TTY readability remain required.

## Findings

- A full visual design system is unnecessary for this SPEC; the right bar is a polished, predictable terminal workflow.
- AltScreen must not be the only evidence surface because users need a durable summary after exit.

## Decisions

- Use Lip Gloss styles for active rows, muted hints, success, warning, and error states.
- Keep key hints visible but compact.

## Follow-ups

- During implementation, test narrow terminal rendering enough to avoid clipped labels or unreadable summaries.

## Recommendation

- Cleared for implementation.
