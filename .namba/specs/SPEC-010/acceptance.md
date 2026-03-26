# Acceptance

- [x] `namba pr "<title>"` is accepted by the CLI and fails clearly when run outside a git repository, on the configured base branch, or without GitHub CLI authentication.
- [x] `namba pr` runs `namba sync` and validation by default, then creates or reuses a PR into the configured base branch from the current work branch.
- [x] `namba pr` ensures the Codex review request comment exists without creating duplicates.
- [x] `namba land` is accepted by the CLI and can resolve the target PR from the current branch when a PR number is not passed.
- [x] `namba land` waits for required checks when requested, merges only when the PR merge state is clean, and reports blocking review or check failures clearly.
- [x] After a successful `namba land`, local `main` is updated safely without clobbering unrelated working tree changes.
- [x] Tests cover CLI parsing, PR reuse, comment idempotency, merge gating, and local main update behavior.
- [x] README and generated Codex guidance explain why `sync`, `pr`, and `land` are separate commands.
- [x] Validation commands pass.

Note: checklist synced after reviewing `pr_land_command.go`, `pr_land_command_test.go`, README/workflow-guide sync artifacts, and rerunning validation in this shell.
