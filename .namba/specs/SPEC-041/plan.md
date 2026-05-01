# SPEC-041 Plan

1. Capture the Codex `0.124.0`-`0.128.0` compatibility baseline.
   - Store the upstream findings in SPEC evidence and update `docs/codex-upstream-reference.md` only if the repository needs a durable summary beyond this SPEC.
   - Treat official Codex developer docs as the primary source and `openai/codex` release/tag pages as implementation evidence.

2. Add capability-regression coverage.
   - Add fixtures or table-driven tests for representative `codex --version`, `codex exec --help`, and `codex exec resume --help` outputs across `0.124.0`-`0.128.0`.
   - Assert `parseCodexCommandCapabilities` detects current supported flags and tolerates additive help text or JSON output fields.
   - Before adding `codex exec --json` regressions, identify the Namba-owned consumer that parses Codex JSON output. If no owned consumer exists, replace the JSON regression with a documented compatibility note and keep tests focused on help/version parsing.

3. Verify and document concurrency semantics.
   - Keep `.codex/config.toml` same-workspace subagent capacity at `max_threads = 5`.
   - Keep `.namba/config/sections/workflow.yaml` at `max_parallel_workers: 3` unless implementation explicitly accepts raising worktree fan-out.
   - Add docs/tests that distinguish Codex subagent threads from Namba-managed git worktree workers.
   - Add a concrete proof step for `[agents] max_threads = 5`: either validate the generated `.codex/config.toml` against the target Codex config schema/local `0.128.0` config surface, or add a fixture-based test that protects the schema shape Namba emits.

4. Update generated and operator-facing guidance.
   - Clarify `codex update` versus `namba update`.
   - Remove or discourage `--full-auto` guidance and prefer explicit sandbox/approval/profile wording.
   - Preserve the repo-safe config boundary: permission profiles stay user-owned unless deliberately widened.
   - Mention persisted `/goal` workflows as future opportunity, not current dependency.

5. Revalidate hook, plugin, and MCP behavior.
   - Check Namba hook event names and typed observation contract against Codex stable hooks for MCP, `apply_patch`, and long-running Bash.
   - Scope MCP and plugin validation to Namba-owned integration boundaries: repo-managed MCP presets, approval persistence assumptions, custom metadata boundaries, and hook enablement guidance. Do not attempt to certify upstream marketplace/cache/import behavior end-to-end.
   - Confirm no `.codex/skills` mirror or unmanaged skill path is reintroduced.

6. Optional implementation branch: raise worktree parallelism to 5 only if accepted.
   - Update `.namba/config/sections/workflow.yaml`, `renderWorkflowConfig()`, and tests expecting `max_parallel_workers: 3`.
   - Expand parallel lifecycle/report/cleanup tests for five workers, merge-block behavior, and preserved worktree handling.
   - Keep this as a separable commit or follow-up SPEC if it broadens the risk profile.

7. Validate and sync.
   - Run `codex --version`, `codex exec --help`, and `codex exec resume --help` on local `0.128.0` where available.
   - Run `go test ./...`.
   - Run `gofmt -l "cmd" "internal" "namba_test.go"`.
   - Run `go vet ./...`.
   - Run `namba sync` after implementation if generated docs, README surfaces, or project artifacts changed.
