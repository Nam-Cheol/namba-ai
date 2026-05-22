---
name: namba-help
description: Read-only entry point for explaining how to use NambaAI in the current repository.
---

State effect: read-only guidance. Do not mutate repository state.

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

Use this skill when the user explicitly says `$namba-help`, asks how to use NambaAI, wants to know which command or skill to use next, or needs a read-only walkthrough of the current repository workflow.

Behavior:
- Stay read-only. Do not mutate repository state, create a SPEC, run validators, or invoke workflow commands just to answer a usage question.
- Prefer `.namba/` and generated repository docs as the primary source of truth: `README*.md`, `docs/getting-started*.md`, `docs/workflow-guide*.md`, `.namba/codex/README.md`, and relevant repo skills under `.agents/skills/`.
- Explain the practical differences between `$namba-create`, `namba project`, `namba codex access`, `namba plan`, `namba harness`, `namba fix`, `namba run`, `namba queue`, `namba sync`, `namba pr`, `namba land`, `namba release`, `namba regen`, `namba update`, `$namba-review-resolve`, `$namba-release`, and `namba doctor` when those distinctions matter to the user's goal.
- When the user describes an outcome, recommend the next Namba command or skill concretely instead of giving a vague taxonomy.
- If the user asks about pre-implementation review flow, explain when to use `$namba-plan-review` versus the individual review skills.
- Distinguish `$namba-create` from `namba plan` and `namba harness`: use create when the user wants repo-local skills or custom agents directly, and use plan or harness when the user needs a SPEC package first.
- If the user asks about a specific command, include concrete invocation examples and mention whether the path is read-only guidance, planning, execution, or GitHub handoff.
- If repository docs and executable/config evidence disagree, call out the mismatch instead of smoothing it over.
