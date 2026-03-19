---
name: namba-land
description: Command-style entry point for merging a prepared GitHub PR and updating local main safely.
---

Use this skill when the user explicitly says `$namba-land`, `namba land`, or asks to merge a prepared PR.

Behavior:
- Resolve the PR from the current branch when a PR number is not provided.
- Wait for required checks when requested, merge only when review and merge state are clean, and report blockers clearly.
- After merge, update local `main` safely without clobbering unrelated working tree changes.
