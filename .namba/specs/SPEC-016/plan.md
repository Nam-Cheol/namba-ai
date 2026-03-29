# SPEC-016 Plan

1. Inspect the current standalone runner, helper shortcut, delegation prompt builder, parallel worktree path, and validation pipeline to lock the exact runtime gaps this SPEC is fixing.
2. Define the expanded execution contract, config precedence, and logging schema so repo defaults, per-run overrides, and effective Codex runtime metadata are explicit and testable.
3. Refactor standalone Codex invocation behind one shared execution builder so the main runner path and any helper execution path cannot drift semantically.
4. Implement the new Codex CLI argument mapping for runtime controls such as `model`, `profile`, `web_search`, `add_dirs`, and `session_mode`, including request/result log updates.
5. Design and implement a stateful run controller that can keep one Codex session alive across implementation, validation, repair, and revalidation cycles with bounded retry policy.
6. Make the mode contract explicit and truthful: `--solo` stays single-runner, `--team` becomes real same-workspace multi-agent execution, and `--parallel` stays reserved for multi-worktree execution.
7. Rework `--team` execution so selected role runtime profiles affect the actual same-workspace execution path instead of only prompt text.
8. Rework `--parallel` execution so worker runs overlap with bounded concurrency, then preserve today's merge-after-pass and preserve-on-failure behavior.
9. Add instruction-surface refresh detection for commands that regenerate `AGENTS.md`, `.agents/skills/*`, or `.codex/agents/*.toml`, and surface a clear session refresh requirement when applicable.
10. Add preflight checks for runtime prerequisites and generalize validation into a configurable pipeline that can express more than only test/lint/typecheck.
11. Expand observability so request/result/validation artifacts record effective runtime metadata, delegation execution details, session identity, retry counts, and parallel worker timing.
12. Update `README.md`, localized README bundles, and workflow docs so users can understand the revised `--solo`/`--team`/`--parallel` contract without reading source code.
13. Add regression coverage for execution-contract mapping, helper/runner unification, repair-loop behavior, executable team delegation, real parallel overlap, session-refresh signaling, preflight/validation behavior, and user-facing docs where stable assertions are appropriate.
14. Add an opt-in live Codex smoke suite for runtime compatibility, then run the standard validation flow and any gated smoke checks needed to verify the final harness.
