# Design Review

- Status: clear
- Last Reviewed: 2026-05-22
- Reviewer: `namba-designer`
- Command Skill: `$namba-plan-design-review`
- Recommended Role: `namba-designer`

## Focus

- Clarify CLI UX hierarchy, advisory wording, noise control, and interaction quality before implementation starts.

- Evidence Status: not-applicable
- Gate Decision: not-applicable
- Approved Direction: concise terminal advisory in human-readable maintenance output, explicit refresh in doctor, no prompt in JSON or automation-safe flows
- Banned Patterns: auto-running updates, alarmist warnings for unknown state, repeated nag text, noisy CI output, JSON chatter, telemetry, ambiguous Codex update wording
- Negative-First Contract: not-applicable; this is CLI text behavior and does not require generated image assets
- Default Library Fit: use existing CLI output helpers and tests rather than adding a new terminal UI dependency
- Context-Specific Bans And Replacements: replace silent behind drift with one concise next-step line; replace implicit update action with consent-first language
- Reference-Driven Asset Manifest: not-applicable
- Generated Image Plan: not-applicable
- Visual Grammar: plain line-oriented terminal output with clear labels and no color-only or emoji-only meaning
- Generic-Section Proof: the generic risk is "new version available" nagging; the replacement is state-scoped advisory text only when behind is proven
- Architecture Handoff: centralize wording so doctor/status/regen/project/sync remain consistent
- Violation-Check Plan: tests assert no prompts for current, dev, stale, failed, JSON, CI, non-TTY, and automation-safe flows
- Open Questions: exact TTL duration and exact cache path naming
- Unresolved Questions: none blocking; TTL and path naming are implementation details covered by acceptance
- Design Review Axes: clarity, restraint, consent, automation safety, command fit, operator confidence
- Keep / Fix / Quick Wins: keep update and Codex update distinction; fix silent drift only in human surfaces; add explicit refresh discoverability

## Review Checklist

- Advisory text is concise enough to live below existing command output.
- The text makes user consent explicit before `namba update`.
- Unknown, stale, offline, dev, and parse-failed states do not produce update nags.
- `status --json` remains machine-readable only.
- `doctor --check-update` gives operators a clear intentional action when they want fresh metadata.
- No visual styling, emoji, or color is required for meaning.

## Findings

- The user experience should feel like a helpful maintenance hint, not a security warning or forced upgrade.
- `doctor` is the right place for explicit refresh because it already owns local readiness inspection.
- `status` should remain compact; behind state can be one line, while details belong in doctor or explicit refresh.
- `regen`, `project`, and `sync` should place update advice in next-step language after their normal success output.

## Decisions

- Use one centralized advisory phrase family across all human surfaces.
- Prefer quiet unknown behavior over noisy speculative prompts.
- Keep the next action framed as "ask the user before running `namba update`".

## Follow-ups

- Verify terminal snapshots or output assertions keep the advisory to one or two short lines.
- Verify no prompt appears when installed version is `dev`.
- Verify wording never suggests `namba update` updates Codex.

## Recommendation

- Cleared for implementation.
