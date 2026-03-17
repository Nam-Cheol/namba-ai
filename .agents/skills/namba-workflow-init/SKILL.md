---
name: namba-workflow-init
description: Codex-adapted init workflow that maps MoAI and Claude assets into NambaAI scaffold.
---

Use this skill when the user asks about `namba init`, project bootstrap, or Claude-to-Codex migration.

Core mapping:
- `CLAUDE.md` -> `AGENTS.md`
- `.claude/skills/*` -> `.agents/skills/*` with `.codex/skills/*` as a compatibility mirror
- `.claude/agents/*` -> `.codex/agents/*.toml` custom agents with `.md` role-card mirrors
- `.claude/hooks/*` -> explicit validation pipeline and `namba` orchestration
- Claude custom slash-command workflows -> built-in Codex slash commands plus the `$namba` skill and `namba` CLI

When implementing init changes:
1. Keep `.namba/config/sections/*.yaml` as the durable source of truth.
2. Never write tokens or secrets into generated config files.
3. Prefer repo-local skills and `.toml` custom agents while keeping `.md` files as readable mirrors.
4. Keep one selected human language aligned across Codex conversation, docs, PR content, and code comments unless the user explicitly overrides it.
5. Keep generated assets readable so users can understand what `namba init .` changed.
