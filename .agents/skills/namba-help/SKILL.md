---
name: namba-help
description: Read-only entry point for explaining how to use NambaAI in the current repository.
---

Use this skill when the user explicitly says `$namba-help`, asks how to use NambaAI, wants to know which command or skill to use next, or needs a read-only walkthrough of the current repository workflow.

Behavior:
- Stay read-only. Do not mutate repository state, create a SPEC, run validators, or invoke workflow commands just to answer a usage question.
- Prefer `.namba/` and generated repository docs as the primary source of truth: `README*.md`, `docs/getting-started*.md`, `docs/workflow-guide*.md`, `.namba/codex/README.md`, and relevant repo skills under `.agents/skills/`.
- Explain the practical differences between `$namba-create`, `namba project`, `namba codex access`, `namba plan`, `namba harness`, `namba fix`, `namba run`, `namba sync`, `namba pr`, `namba land`, `namba regen`, `namba update`, and `namba doctor` when those distinctions matter to the user's goal.
- When the user describes an outcome, recommend the next Namba command or skill concretely instead of giving a vague taxonomy.
- If the user asks about pre-implementation review flow, explain when to use `$namba-plan-review` versus the individual review skills.
- Distinguish `$namba-create` from `namba plan` and `namba harness`: use create when the user wants repo-local skills or custom agents directly, and use plan or harness when the user needs a SPEC package first.
- If the user asks about a specific command, include concrete invocation examples and mention whether the path is read-only guidance, planning, execution, or GitHub handoff.
- If repository docs and executable/config evidence disagree, call out the mismatch instead of smoothing it over.
