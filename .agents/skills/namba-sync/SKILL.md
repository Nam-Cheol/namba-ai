---
name: namba-sync
description: Command-style entry point for refreshing Namba project artifacts after implementation.
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

Use this skill when the user explicitly says `$namba-sync`, `namba sync`, or asks to refresh PR-ready Namba artifacts after changes.

Behavior:
- Refresh `.namba/project/*` docs, release notes/checklists, codemaps, advisory review readiness under `.namba/specs/<SPEC>/reviews/`, and any README bundles enabled by `.namba/config/sections/docs.yaml` after implementation.
- Keep PR creation and merge automation in `namba pr` and `namba land` so `sync` stays a local artifact refresh command.
- Use `namba regen` separately when template-generated scaffold assets changed.
- Run validation first when code changed and the quality config requires it.
