---
name: namba-init
description: Command-style entry point for project bootstrap with NambaAI.
---

Use this skill when the user explicitly says `$namba-init`, `namba init`, or asks to bootstrap a repository with NambaAI.

Behavior:
- Prefer running the installed `namba init` CLI when available because it writes the scaffold deterministically.
- Keep `.namba/config/sections/*.yaml` as the durable source of truth.
- Explain that repo skills live under `.agents/skills/` and Codex subagents live under `.codex/agents/*.toml`.
- Keep the selected human language aligned across Codex conversation, docs, PR content, and code comments.
