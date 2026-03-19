# Acceptance

- [x] `namba run SPEC-XXX --solo` is accepted by the CLI and produces a standalone Codex execution request that explicitly targets a single-subagent workflow.
- [x] `namba run SPEC-XXX --team` is accepted by the CLI and produces a standalone Codex execution request that explicitly targets a multi-subagent workflow.
- [x] `namba run SPEC-XXX --parallel` keeps its current worktree fan-out/fan-in behavior and is not silently reinterpreted as subagent parallelism.
- [x] Invalid flag combinations fail with clear errors, including at minimum `--solo --team`, `--solo --parallel`, and `--team --parallel`.
- [x] Tests cover CLI parsing and execution request generation for default, solo, team, and conflicting mode combinations.
- [x] README and generated Codex guidance explain the difference between default runs, subagent runs, and worktree parallel runs in terms a user can act on.
- [x] Validation commands pass.
Note: checklist synced after adding explicit `--solo` / `--team` run modes, refreshing generated Codex guidance, and rerunning validation.
