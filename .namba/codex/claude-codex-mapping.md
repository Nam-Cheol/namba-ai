# Claude Code to Codex Mapping

This repository uses a Codex-adapted variant of the MoAI bootstrap model.

- `CLAUDE.md` -> `AGENTS.md`
- `.claude/skills/*` -> `.agents/skills/*` and `.codex/skills/*`
- `.claude/agents/*.md` -> `.codex/agents/*.toml` custom agents
- `.claude/hooks/*` -> explicit validation commands, structured run logs, and `namba sync`
- Claude slash-command-centric workflows -> built-in Codex slash commands plus `$namba` and `namba`

Why this is different:
- Claude Code has first-class hooks, subagents, and project slash-command workflows.
- Codex has AGENTS, repo-local skills, repo-local config, built-in slash commands, and experimental multi-agent delegation.
- NambaAI keeps the workflow semantics but ports the control surface into Codex-compatible assets.
