# Acceptance

- [ ] `namba pr "<title>"` is accepted by the CLI and fails clearly when run outside a git repository, on the configured base branch, or without GitHub CLI authentication.
- [ ] `namba pr` runs `namba sync` and validation by default, then creates or reuses a PR into the configured base branch from the current work branch.
- [ ] `namba pr` ensures the Codex review request comment exists without creating duplicates.
- [ ] `namba land` is accepted by the CLI and can resolve the target PR from the current branch when a PR number is not passed.
- [ ] `namba land` waits for required checks when requested, merges only when the PR merge state is clean, and reports blocking review or check failures clearly.
- [ ] After a successful `namba land`, local `main` is updated safely without clobbering unrelated working tree changes.
- [ ] Tests cover CLI parsing, PR reuse, comment idempotency, merge gating, and local main update behavior.
- [ ] README and generated Codex guidance explain why `sync`, `pr`, and `land` are separate commands.
- [ ] Validation commands pass.
