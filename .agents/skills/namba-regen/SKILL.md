---
name: namba-regen
description: Command-style entry point for regenerating Namba scaffold assets from config.
---

State effect: mutating workflow entry point. Use help/probe paths read-only, and otherwise expect repository state or GitHub state to change.

Use this skill when the user explicitly says `$namba-regen`, `namba regen`, or asks to re-render generated Namba assets from configuration.

Behavior:
- Regenerate `AGENTS.md`, repo skills under `.agents/skills/`, `.codex/agents/*.toml`, readable `.md` role cards, `.namba/codex/*`, and `.codex/config.toml` from `.namba/config/sections/*.yaml`.
- Do not recreate `.codex/skills/`; that mirror causes duplicate skill discovery in Codex.
- Remove obsolete generated skill files when the template set changes.
