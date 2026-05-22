---
name: namba-init
description: Command-style entry point for project bootstrap with NambaAI.
---

State effect: mutating workflow entry point. Use help/probe paths read-only, and otherwise expect repository state or GitHub state to change.

Generated instruction contract for this command skill:
- Purpose: keep the role or command scope explicit, bounded, and testable.
- Boundary: honor read-only versus mutating state effects, configured sandbox mode, and assigned file or workflow ownership.
- Required output: report concrete actions, changed paths or artifacts, validation evidence, and pass/fail status or blockers.
- Pass/fail criteria: claim success only when acceptance criteria and configured validation are satisfied; otherwise name the exact blocker and impact.
- Evidence expectations: cite source artifacts such as SPEC files, `.namba/` configs, diffs, test output, PR/check links, or generated manifests instead of relying on unsupported assertions.
- Security responsibilities: never expose or commit secrets; treat auth, privacy, destructive commands, permission changes, and external network or credential use as security-sensitive.
- Destructive command and escalation policy: do not run destructive commands unless explicitly requested; request approval for privileged, networked, or sandbox-blocked actions only when the active approval mode allows it, and otherwise report the blocker or use a safe non-escalating path.
- Fallback implementer boundary: if a specialist path is unavailable and the main/default implementer takes over, stay within the assigned scope and preserve the same evidence and validation duties.
- Portability: keep durable guidance non-project-specific unless the current repository config or SPEC explicitly provides the project detail.

Use this skill when the user explicitly says `$namba-init`, `namba init`, or asks to bootstrap a repository with NambaAI.

Behavior:
- Prefer running the installed `namba init` CLI when available because it writes the scaffold deterministically.
- Keep `.namba/config/sections/*.yaml` as the durable source of truth.
- Treat the init wizard as a repository-state-first flow: existing code should use detected language/framework defaults, while empty repositories should not ask for a starter app stack and should leave stack choice to the first clarified planning request. Use approachable emoji cues for step categories, echo each answer with a clear success marker before advancing, support `b`/`back` to revise the previous step, and do not ask the user for a GitHub username during onboarding.
- After init, direct the user to open interactive Codex, run `/hooks` when Codex reports `6 hooks need review`, inspect the generated `.codex/hooks/namba_codex_guard.py` commands, and approve them before expecting prompt-refinement hooks to run.
- Explain that repo skills live under `.agents/skills/` and Codex subagents live under `.codex/agents/*.toml`.
- Keep the selected human language aligned across Codex conversation, docs, PR content, and code comments.
