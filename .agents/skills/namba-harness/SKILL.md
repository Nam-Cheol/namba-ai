---
name: namba-harness
description: Command-style entry point for creating the next harness-oriented SPEC package.
---

State effect: mutating workflow entry point. Use help/probe paths read-only, and otherwise expect repository state or GitHub state to change.

Generated instruction contract for this command skill:
- Purpose: keep the role or command scope explicit, bounded, and testable.
- Boundary: honor read-only versus mutating state effects, configured sandbox mode, and assigned file or workflow ownership.
- Required output: report concrete actions, changed paths or artifacts, validation evidence, and pass/fail status or blockers.
- Pass/fail criteria: claim success only when acceptance criteria and configured validation are satisfied; otherwise name the exact blocker and impact.
- Evidence expectations: cite source artifacts such as SPEC files, `.namba/` configs, diffs, test output, PR/check links, or generated manifests instead of relying on unsupported assertions.
- Security responsibilities: never expose or commit secrets; treat auth, privacy, destructive commands, permission changes, and external network or credential use as security-sensitive.
- Destructive command and escalation policy: do not run destructive commands unless explicitly requested, and request approval for privileged, networked, or sandbox-blocked actions.
- Fallback implementer boundary: if a specialist path is unavailable and the main/default implementer takes over, stay within the assigned scope and preserve the same evidence and validation duties.
- Portability: keep durable guidance non-project-specific unless the current repository config or SPEC explicitly provides the project detail.

Use this skill when the user explicitly says `$namba-harness`, `namba harness`, or asks to create a harness-oriented SPEC package.

Behavior:
- Before reading project docs, checking Git state, or running the CLI, run the same Namba clarification gate used by `namba plan` on the user's raw request.
- If the request is short, broad, or missing target surface, user flow, scope boundaries, constraints, acceptance criteria, or validation, do not run `namba harness` yet. Ask 1-3 concise questions first.
- Treat Codex Plan mode as the preferred clarification surface, then pass only the refined Goal/Scope/Constraints/Acceptance description to `namba harness "<refined description>"`.
- Prefer the installed `namba harness` CLI when available.
- Use this path for reusable agent, skill, workflow, orchestration, or evaluation scaffolding when the user wants a reviewable SPEC first instead of generating the repo-local skill or agent artifact directly through `$namba-create`.
- Start with the same dedicated-branch planning contract as `namba plan`, and use `--current-workspace` only when the user intentionally wants to scaffold on the current branch without creating a dedicated SPEC branch.
- Do not create planning worktrees here either; temporary worktrees belong to overlapping `namba run SPEC-XXX --parallel` execution only.
- Create the next sequential `SPEC-XXX` package under `.namba/specs/` without inventing a second artifact model.
- Seed `.namba/specs/<SPEC>/reviews/` with product, engineering, design, and aggregate readiness artifacts so the review flow stays aligned with `namba plan`.
- Keep command-entry skill guidance lean; move long PR-thread, CI-log, frontend, or MCP recipes into existing references or deterministic helper-script candidates instead of creating new standalone skills.
- Evaluate deterministic helper-script candidates before implementation: they need `--help`, read-only defaults, bounded output, explicit network/auth assumptions, fixture or local-server tests, and no destructive or third-party app coupling.
- For large managed-skill changes, inventory affected source and generated surfaces, classify mechanical versus behavioral edits, update templates first, regenerate, review generated diffs, and validate.
- For harness/MCP quality, prefer workflow-first designs over raw endpoint wrappers, require context-budgeted outputs with pagination or truncation expectations, produce actionable errors, and define evaluation scenarios that are independent, read-only, realistic, verifiable, and stable.
- Keep the output Codex-native and avoid Claude-only runtime primitives in the planned contract.
