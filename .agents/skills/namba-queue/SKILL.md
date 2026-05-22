---
name: namba-queue
description: Command-style entry point for operating the existing-SPEC queue conveyor.
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

Use this skill when the user explicitly says `$namba-queue`, `namba queue`, or asks to process multiple existing SPEC packages in order.

Behavior:
- Prefer the installed `namba queue` CLI when available because queue state and Git/GitHub evidence are durable CLI-owned outputs.
- Only consume already-existing SPEC packages. Do not create new SPEC packages from this command surface.
- `namba queue start <SPEC-RANGE|SPEC-LIST>` accepts ranges such as `SPEC-001..SPEC-003` and explicit lists such as `SPEC-001 SPEC-004`, plus `--auto-land`, `--review`, deprecated `--skip-codex-review`, and `--remote origin`.
- Use `namba queue status [--verbose]` to report active SPEC, durable state, blocker or wait reason, evidence path, PR link, and next safe command before deciding how to resume.
- Use `namba queue resume` only after checking or resolving the current wait/blocker state; use `pause` and `stop` as cooperative controls that preserve branches, PRs, and evidence.
- Treat `.namba/logs/queue/` as the durable queue state and report surface.
- Continue one active SPEC at a time through review, implementation, validation, active-SPEC-aware sync/PR, checks, optional land, and local main refresh.
- Block instead of skipping on failed validation, failed checks, non-mergeable PRs, dirty queue branches, GitHub auth failures, missing `gh`, diverged branches, ambiguous PR/check state, or unclear review readiness.
- Without `--auto-land`, stop in `waiting_for_land` after green and mergeable PR evidence so the operator can land intentionally.
- Queue PR handoff does not request Codex review unless `--review` is present; keep `--skip-codex-review` as deprecated compatibility syntax only, and never combine it with `--review`.
