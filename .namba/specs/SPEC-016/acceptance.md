# Acceptance

- [ ] Standalone execution request artifacts capture the expanded Codex runtime contract, including resolved values for `model`, `profile`, `web_search`, `add_dirs`, and `session_mode` when those controls are in use.
- [ ] Standalone execution uses one shared Codex invocation builder across the primary runner path and any helper path, so hard-coded semantic drift such as a separate unconditional `--full-auto` path is removed or made an explicit contract choice.
- [ ] Effective Codex runtime metadata is persisted in run artifacts, including the effective model/profile, search setting, additional directories, delegation mode, retry count, and session identifier when applicable.
- [ ] `namba run SPEC-XXX --solo` remains the single-runner, single-workspace path and does not silently drift into same-workspace team orchestration or worktree parallelism.
- [ ] `namba run SPEC-XXX --team` means same-workspace multi-agent execution rather than worktree parallelism, uses real executable delegation semantics rather than prompt-only decoration, and selected role runtime profiles materially affect the execution path or child session settings.
- [ ] Execution results no longer report delegation as universally absent when team orchestration actually occurred; logs make it clear when delegation was or was not exercised.
- [ ] The standalone runner can perform a bounded implement/validate/repair/revalidate loop inside one continuous Codex session and stops with a clear terminal error when retries are exhausted.
- [ ] Validation remains available for `test`, `lint`, and `typecheck` and can additionally express configured steps such as build, migration dry-run, smoke start, or output-contract verification.
- [ ] Preflight checks fail fast with actionable messages when required runtime prerequisites such as project root, git/Codex availability, extra directory resolution, or required environment inputs are missing.
- [ ] `namba run SPEC-XXX --parallel` remains reserved for multi-worktree fan-out/fan-in, executes worker runs concurrently up to the configured limit, merges only after every worker passes execution plus validation, and still preserves failed worker worktrees and branches for inspection.
- [ ] Parallel execution logs contain enough timing or worker metadata to distinguish real overlap from sequential execution.
- [ ] When repo-generated instruction surfaces such as `AGENTS.md` or `.codex/agents/*.toml` change in a way that existing Codex sessions may not absorb automatically, Namba emits a clear session refresh requirement.
- [ ] `README.md`, localized README bundles, and generated workflow documentation clearly explain the differences between `--solo`, `--team`, and `--parallel`, along with repair behavior and any session refresh expectations introduced by this work.
- [ ] Regression tests covering execution-contract mapping, repair-loop behavior, delegation execution, parallel overlap, session-refresh signaling, and preflight/validation behavior are present.
- [ ] An opt-in live smoke path guarded by `CODEX_SMOKE=1` can exercise a real Codex CLI run in an isolated temporary repository without becoming mandatory for default local or CI test runs.
- [ ] Validation commands pass
