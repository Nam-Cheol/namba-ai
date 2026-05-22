---
name: namba-workflow-init
description: Codex-adapted init workflow that maps MoAI and Claude assets into NambaAI scaffold.
---

Use this skill when the user asks about `namba init`, project bootstrap, or Claude-to-Codex migration.

State effect: mutating scaffold workflow when applied. Read-only when used only to explain migration or init behavior.

Generated instruction contract for this repo skill:
- Purpose: keep the role or command scope explicit, bounded, and testable.
- Boundary: honor read-only versus mutating state effects, configured sandbox mode, and assigned file or workflow ownership.
- Required output: report concrete actions, changed paths or artifacts, validation evidence, and pass/fail status or blockers.
- Pass/fail criteria: claim success only when acceptance criteria and configured validation are satisfied; otherwise name the exact blocker and impact.
- Evidence expectations: cite source artifacts such as SPEC files, `.namba/` configs, diffs, test output, PR/check links, or generated manifests instead of relying on unsupported assertions.
- Security responsibilities: never expose or commit secrets; treat auth, privacy, destructive commands, permission changes, and external network or credential use as security-sensitive.
- Destructive command and escalation policy: do not run destructive commands unless explicitly requested, and request approval for privileged, networked, or sandbox-blocked actions.
- Fallback implementer boundary: if a specialist path is unavailable and the main/default implementer takes over, stay within the assigned scope and preserve the same evidence and validation duties.
- Portability: keep durable guidance non-project-specific unless the current repository config or SPEC explicitly provides the project detail.

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
