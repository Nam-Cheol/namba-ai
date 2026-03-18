# SPEC-010 Plan

1. Add CLI surface and argument parsing for `namba pr` and `namba land`, including clear errors for unsupported states and missing GitHub CLI.
2. Implement `namba pr` so it can run sync and validation by default, create a commit from the current branch state, push the branch, create or reuse a PR into the configured base branch, and ensure the Codex review request comment exists without duplication.
3. Implement `namba land` so it can resolve the target PR, wait for required checks when requested, merge only from a clean merge state, and update local `main` safely after merge.
4. Cover PR reuse, review-comment idempotency, merge gating, and local main update behavior with tests.
5. Refresh generated docs and command-entry guidance so users understand the difference between `sync`, `pr`, and `land`.
6. Run validation commands and sync artifacts.
