# Acceptance

- [ ] Parallel execution records worker results before any merge decisions are made.
- [ ] If any worker fails execution or validation, no worker branch is merged into the base branch.
- [ ] Successful parallel runs merge only after every worker passes the run and validation gates.
- [ ] Parallel execution reports merge failures and cleanup failures explicitly.
- [ ] Temporary worktrees and branches follow a documented cleanup policy after success and failure.
- [ ] `run --parallel --dry-run` still avoids runner execution, merges, and cleanup side effects.
- [ ] Validation commands pass.
- [ ] Tests covering the new behavior are present.