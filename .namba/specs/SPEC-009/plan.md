# SPEC-009 Plan

1. Refresh project context with `namba project` and confirm the latest Codex subagent guidance that affects CLI behavior and generated docs.
2. Introduce an explicit `run` execution mode model that distinguishes default, `--solo`, `--team`, and worktree `--parallel`.
3. Update execution request generation so `codex exec` receives mode-specific instructions for subagent use without changing existing validation flow.
4. Define and test invalid flag combinations such as `--solo --team`, `--solo --parallel`, and `--team --parallel`.
5. Update README and generated Codex guidance so users understand the difference between subagent orchestration and worktree parallelism.
6. Run validation commands and sync artifacts with `namba sync`.
