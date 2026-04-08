---
name: namba-create
description: Skill-first entry point for creating a repo-local skill, a project-scoped custom agent, or both.
---

Use this skill when the user explicitly says `$namba-create` or asks to create a repo-local skill, a project-scoped custom agent, or both through Namba.

Behavior:
- Phase 1 is skill-first. Do not introduce a new `namba create` Go CLI command as part of this flow.
- Run the interaction as a staged generator: `unresolved` -> `narrowed` -> `confirmed`.
- Keep each turn stateful and visible: summarize the current candidate target and the remaining unresolved items before asking the next clarifying question.
- If the user explicitly says `skill`, `agent`, or `both`, treat that directive as authoritative over any heuristic classification.
- Use `sequential-thinking` when decomposition or clarification planning is non-trivial, use `context7` only when targeted external library or framework guidance materially helps the generated instructions, and use `playwright` only when browser verification is actually relevant.
- Before any write, present a non-mutating preview that includes the chosen output type, slug or name, intended file paths, validation plan, whether a fresh Codex session will likely be required, and the planned five-role analysis or verification record.
- Do not write files until the target, slug, paths, and overwrite decisions are explicit and the user has confirmed the preview.
- Normalize names into a safe slug, constrain writes to `.agents/skills/<slug>/SKILL.md`, `.codex/agents/<slug>.toml`, and `.codex/agents/<slug>.md`, and reject path traversal, invalid slugs, silent overwrites, or incomplete agent mirror pairs.
- Reject durable instructions that preserve raw unnormalized user prose, stale Claude-only primitives, or repository-policy violations.
- Record at least five independent role outputs across planning or verification when the flow advances to generation, while degrading safely if the effective same-workspace thread limit is lower than the repo default.
- Keep user-authored outputs distinct from Namba-managed built-ins so `namba regen` preserves them, and surface fresh-session guidance clearly when instruction surfaces change.
