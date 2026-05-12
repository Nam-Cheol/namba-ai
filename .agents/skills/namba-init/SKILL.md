---
name: namba-init
description: Command-style entry point for project bootstrap with NambaAI.
---

Use this skill when the user explicitly says `$namba-init`, `namba init`, or asks to bootstrap a repository with NambaAI.

Behavior:
- Prefer running the installed `namba init` CLI when available because it writes the scaffold deterministically.
- Keep `.namba/config/sections/*.yaml` as the durable source of truth.
- Treat the init wizard as a repository-state-first flow: existing code should use detected language/framework defaults, while empty repositories should not ask for a starter app stack and should leave stack choice to the first clarified planning request. Use approachable emoji cues for step categories, echo each answer with a clear success marker before advancing, support `b`/`back` to revise the previous step, and do not ask the user for a GitHub username during onboarding.
- After init, direct the user to open interactive Codex, run `/hooks` when Codex reports `6 hooks need review`, inspect the generated `.codex/hooks/namba_codex_guard.py` commands, and approve them before expecting prompt-refinement hooks to run.
- Explain that repo skills live under `.agents/skills/` and Codex subagents live under `.codex/agents/*.toml`.
- Keep the selected human language aligned across Codex conversation, docs, PR content, and code comments.
