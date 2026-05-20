---
name: namba-pr
description: Command-style entry point for preparing the current branch for GitHub review.
---

State effect: mutating workflow entry point. Use help/probe paths read-only, and otherwise expect repository state or GitHub state to change.

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
