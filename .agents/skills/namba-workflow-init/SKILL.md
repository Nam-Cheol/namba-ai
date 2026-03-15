---
name: namba-workflow-init
description: Codex-adapted init workflow that maps MoAI and Claude assets into NambaAI scaffold.
---

Use this skill when the user asks about `namba init`, project bootstrap, or Claude-to-Codex migration.

Core mapping:
- `CLAUDE.md` -> `AGENTS.md`
- `.claude/skills/*` -> `.agents/skills/*` with `.codex/skills/*` as a compatibility mirror
- `.claude/agents/*` -> `.codex/agents/*.md` role cards for Codex delegation
- `.claude/hooks/*` -> explicit validation pipeline and `namba` orchestration
- Claude custom slash-command workflows -> built-in Codex slash commands plus the `$namba` skill and `namba` CLI

When implementing init changes:
1. Keep `.namba/config/sections/*.yaml` as the durable source of truth.
2. Never write tokens or secrets into generated config files.
3. Prefer repo-local skills and agent role cards over provider-specific hidden state.
4. Keep generated assets readable so users can understand what `namba init .` changed.
