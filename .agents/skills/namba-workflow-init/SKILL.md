---
name: namba-workflow-init
description: Codex-adapted init workflow that maps MoAI and Claude assets into NambaAI scaffold.
---

Use this skill when the user asks about `namba init`, project bootstrap, or Claude-to-Codex migration.

Core mapping:
- `CLAUDE.md` -> `AGENTS.md`
- `.claude/skills/*` -> `.agents/skills/*`
- `.claude/commands/*` -> command-entry repo skills such as `.agents/skills/namba-create/SKILL.md` and `.agents/skills/namba-run/SKILL.md`
- `.claude/agents/*` -> `.codex/agents/*.toml` custom agents with `.md` role-card mirrors
- `.claude/hooks/*` -> explicit validation pipeline and `namba` orchestration
- Claude custom slash-command workflows -> built-in Codex slash commands plus repo skills such as `$namba-create`, `$namba-run`, `$namba-pr`, `$namba-land`, `$namba-plan`, `$namba-sync`, and the `namba` CLI

When implementing init changes:
1. Keep `.namba/config/sections/*.yaml` as the durable source of truth.
2. Never write tokens or secrets into generated config files.
3. Prefer repo-local skills and `.toml` custom agents while keeping `.md` files as readable mirrors.
4. Keep one selected human language aligned across Codex conversation, docs, PR content, and code comments unless the user explicitly overrides it.
5. Keep generated assets readable so users can understand what `namba init .` changed.
