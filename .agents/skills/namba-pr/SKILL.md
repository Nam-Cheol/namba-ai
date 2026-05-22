---
name: namba-pr
description: Command-style entry point for preparing the current branch for GitHub review.
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

Use this skill when the user explicitly says `$namba-pr`, `namba pr`, or asks to hand off the current branch for review.

Behavior:
- Use the configured PR base branch, PR language, and Codex review command from `.namba/config/sections/git-strategy.yaml`; let an explicit `--language <en|ko|ja|zh>` request override the configured PR language for that handoff.
- Run `namba sync` and validation by default before creating review artifacts.
- Build the PR body from concrete completed work, changed areas or files, evidence sources, validation results, and the latest SPEC review-readiness artifact when `.namba/specs/<SPEC>/reviews/readiness.md` exists.
- PR bodies must avoid empty placeholder prose; cite source artifacts such as `.namba/project/change-summary.md`, `.namba/project/pr-checklist.md`, SPEC files, validation output, commits, or PR metadata when available.
- Inspect current PR check status before review handoff; for failing GitHub Actions checks, capture run URLs and bounded GitHub Actions failure snippets, and report external checks by status and details URL only.
- Commit and push the current work branch, create or reuse the GitHub PR, and do not request Codex review unless the user explicitly asked for `--review`.
- When `--review` is present, ensure the configured Codex review marker exists exactly once.
- Collaboration defaults: PRs target `main`, PR content is written in Korean, and `@codex review` is the explicit review request command.
