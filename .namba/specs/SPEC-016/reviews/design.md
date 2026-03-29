# Design Review

- Status: approved
- Last Reviewed: 2026-03-29
- Reviewer: Codex
- Command Skill: `$namba-plan-design-review`
- Recommended Role: `namba-designer`

## Focus

- Clarify interaction quality, responsive states, accessibility, and visual direction before implementation starts.

## Findings

- The primary design surface in this SPEC is terminal UX plus generated documentation, not graphical UI. The design risk is therefore ambiguity, verbosity, and inconsistent message hierarchy rather than layout polish.
- Truthful mode semantics are a design requirement. If `--solo`, `--team`, `--parallel`, repair retries, or session refresh states are hard to distinguish from one another in terminal output and docs, users will still experience the runtime as unreliable even if the implementation is technically correct.
- Session refresh messaging is part of interaction quality. When Namba regenerates `AGENTS.md` or `.codex/agents/*.toml`, the product needs a short, explicit message that tells the user what changed and what action is needed, instead of assuming they will infer stale-session behavior.
- Expanded runtime metadata is valuable for debugging, but it must stay layered. The default terminal flow should stay scan-friendly, while deeper effective metadata belongs in logs and structured artifacts.

## Decisions

- Keep `namba --version` intentionally minimal: one concise, script-safe line in the form `namba <version>`.
- Keep run/update messaging phase-oriented and compact. Prefer short lines that answer four questions in order: what mode is running, what Codex/runtime context is effective, whether repair or delegation occurred, and what the next action is.
- Keep session refresh notices explicit and action-led. The preferred interaction is a direct line such as "session refresh required" plus the affected instruction surface, not a buried advisory paragraph.
- Keep documentation structure parallel across English, Korean, Japanese, and Chinese so users can find `run`, `--solo`, `--team`, `--parallel`, session-refresh, update, and uninstall guidance in the same relative place regardless of language.
- Keep terminal copy non-interactive by default. Do not add prompts, banners, or decorative progress output that would reduce shell readability or make behavior harder to test.

## Follow-ups

- During implementation, define one stable wording pattern for success, degraded fallback, and failure states so users can quickly distinguish "repaired in-session", "re-tried with degraded continuity", and "manual retry required".
- Make parallel worker summaries readable at a glance: worker identity, overlap/timing evidence, pass/fail state, and preservation status should be easy to scan without opening JSON first.
- Keep accessibility in the CLI sense: avoid color-only distinctions and avoid requiring users to compare long free-form paragraphs to understand what happened.

## Recommendation

- Advisory recommendation: approved. Proceed with implementation using compact, truthful terminal messaging and parallel multilingual documentation updates, with logs carrying the heavier runtime detail.
