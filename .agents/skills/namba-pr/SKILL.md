---
name: namba-pr
description: Command-style entry point for preparing the current branch for GitHub review.
---

Use this skill when the user explicitly says `$namba-pr`, `namba pr`, or asks to hand off the current branch for review.

Behavior:
- Use the configured PR base branch, PR language, and Codex review marker from `.namba/config/sections/git-strategy.yaml`.
- Run `namba sync` and validation by default before creating review artifacts.
- Include the latest SPEC review-readiness artifact in the PR summary/checklist when `.namba/specs/<SPEC>/reviews/readiness.md` exists.
- Commit and push the current work branch, create or reuse the GitHub PR, and ensure the Codex review marker exists without duplication.
- Collaboration defaults: PRs target `main`, PR content is written in Korean, and `@codex review` is the review marker.
