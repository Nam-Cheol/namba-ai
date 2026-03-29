# Design Review

- Status: approved
- Last Reviewed: 2026-03-29
- Reviewer: Codex
- Command Skill: `$namba-plan-design-review`
- Recommended Role: `namba-designer`

## Focus

- Clarify interaction quality, responsive states, accessibility, and visual direction before implementation starts.

## Findings

- The most important design surface here is terminal UX, not graphical UI. `namba --version` should stay one-line, terse, and copyable so it works well for both humans and scripts.
- `namba update` needs staged, high-signal console messaging. Users should be able to scan for target version, platform or asset context, success or failure state, and the next action without parsing long paragraphs.
- Uninstall guidance should sit near install and update content in every managed language so the lifecycle feels complete and discoverable rather than buried in one language or one guide only.

## Decisions

- Keep `namba --version` output intentionally minimal: one concise line with the resolved version label.
- Keep `namba update` non-interactive and text-first. Prefer short phase-oriented lines over banners, prompts, or noisy progress UI that would be harder to test and less stable across shells.
- Keep documentation structure parallel across English, Korean, Japanese, and Chinese so install, update, and uninstall information is easy to find in the same relative place.

## Follow-ups

- Make Windows-specific restart guidance visually obvious in update success copy because replacement can be deferred there.
- Keep failure copy actionable and compact; avoid generic "something went wrong" wording when release, asset, checksum, or network causes can be named directly.
- Ensure uninstall docs cover both binary removal and PATH cleanup expectations in each supported language set.

## Recommendation

- Advisory recommendation: approved. Proceed with implementation using concise, structured terminal copy and parallel multilingual documentation updates.
