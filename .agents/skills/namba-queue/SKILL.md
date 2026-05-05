---
name: namba-queue
description: Command-style entry point for operating the existing-SPEC queue conveyor.
---

Use this skill when the user explicitly says `$namba-queue`, `namba queue`, or asks to process multiple existing SPEC packages in order.

Behavior:
- Prefer the installed `namba queue` CLI when available because queue state and Git/GitHub evidence are durable CLI-owned outputs.
- Only consume already-existing SPEC packages. Do not create new SPEC packages from this command surface.
- `namba queue start <SPEC-RANGE|SPEC-LIST>` accepts ranges such as `SPEC-001..SPEC-003` and explicit lists such as `SPEC-001 SPEC-004`, plus `--auto-land`, `--skip-codex-review`, and `--remote origin`.
- Use `namba queue status [--verbose]` to report active SPEC, durable state, blocker or wait reason, evidence path, PR link, and next safe command before deciding how to resume.
- Use `namba queue resume` only after checking or resolving the current wait/blocker state; use `pause` and `stop` as cooperative controls that preserve branches, PRs, and evidence.
- Treat `.namba/logs/queue/` as the durable queue state and report surface.
- Continue one active SPEC at a time through review, implementation, validation, active-SPEC-aware sync/PR, checks, optional land, and local main refresh.
- Block instead of skipping on failed validation, failed checks, non-mergeable PRs, dirty queue branches, GitHub auth failures, missing `gh`, diverged branches, ambiguous PR/check state, or unclear review readiness.
- Without `--auto-land`, stop in `waiting_for_land` after green and mergeable PR evidence so the operator can land intentionally.
- Keep the queue-scoped `--skip-codex-review` meaning narrow: skip creating a new `@codex review` marker comment, not review evidence or validation.
