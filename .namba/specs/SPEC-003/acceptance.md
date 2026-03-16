# Acceptance

- [x] Parallel execution records worker results before any merge decisions are made.
- [x] If any worker fails execution or validation, no worker branch is merged into the base branch.
- [x] Successful parallel runs merge only after every worker passes the run and validation gates.
- [x] Parallel execution reports merge failures and cleanup failures explicitly.
- [x] Temporary worktrees and branches follow a documented cleanup policy after success and failure.
- [x] `run --parallel --dry-run` still avoids runner execution, merges, and cleanup side effects.
- [x] Validation commands pass.
- [x] Tests covering the new behavior are present.

Note: checklist synced after implementing the parallel run report, documented cleanup policy, and regression tests.
