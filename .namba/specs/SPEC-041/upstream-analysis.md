# Codex 0.124.0-0.128.0 Upstream Analysis

## Planning Verdict

NambaAI can update its Codex assumptions to the `0.124.0`-`0.128.0` range. The safest implementation path is to strengthen capability probing, fixtures, and generated guidance rather than pinning to a single Codex CLI version.

## Version Matrix

| Version | Date / Evidence | NambaAI Impact |
| --- | --- | --- |
| `0.124.0` | OpenAI changelog and `rust-v0.124.0` release, 2026-04-23 | Stable hooks, app-server multi-environment and per-turn cwd selection, remote plugin marketplace list/read, Bedrock support, Fast service-tier defaulting, permission drift fixes, MCP approval sync, `wait_agent` queue timing fix. Validate hook runtime and multi-workspace assumptions. |
| `0.125.0` | OpenAI changelog and `rust-v0.125.0` release, 2026-04-24 | Unix socket app-server transport, pagination-friendly resume/fork, sticky environments, permission-profile round-trip, `codex exec --json` reasoning-token usage, stricter agent config behavior. Validate capability parser and JSON consumers. |
| `0.126.0` | GitHub tag `rust-v0.126.0`, 2026-04-29 | Public page is tag-centric. Use binary help snapshots and compare/tag evidence. Watch `codex update`, `/goal` groundwork, keymaps, plugin/external-agent migration, `--full-auto` deprecation movement, and release packaging changes. |
| `0.127.0` | GitHub tag `rust-v0.127.0`, 2026-04-30 | Public page is tag-centric. Validate `/goal resume`, remote installed plugin cache for skills/MCP, hooks/list app-server RPC, app MCP path override, MCP approval policy fixes, workspace metadata protections, Windows/unified exec fixes. |
| `0.128.0` | OpenAI changelog and `rust-v0.128.0` release, 2026-04-30 | Stable endpoint. Adds persisted `/goal`, `codex update`, keymaps, plan-mode nudges, expanded permission profiles and sandbox profile selection, plugin marketplace/cache/uninstall/hook enablement, external session import, explicit MultiAgentV2 controls, network hardening, app-server artifacts, and `--full-auto` deprecation guidance. |

## 5-Agent Synthesis

- Release-analysis agents agree the upgrade is feasible if NambaAI validates CLI surfaces by capability instead of version.
- Repository-analysis found the current `.codex/config.toml` already supports 5 same-workspace Codex agent threads.
- Implementation-analysis identified a separate Namba worktree worker cap of 3, which should not be silently changed by a Codex update SPEC.
- The strongest immediate implementation work is tests, fixtures, and generated documentation.
- `/goal` is promising but should not be adopted as a required runtime dependency until Namba has a separate harness for persisted workflows.

## Local Impact Map

- `.codex/config.toml`: preserve `approval_policy = "on-request"`, `sandbox_mode = "workspace-write"`, and `[agents] max_threads = 5`; keep user-specific permission profiles outside repo config.
- `.namba/config/sections/workflow.yaml`: current `max_parallel_workers: 3` controls worktree fan-out, not Codex subagent concurrency.
- `internal/namba/codex_capability.go`: strengthen existing capability probing and parser tests.
- `internal/namba/hook_runtime.go`: validate against stable hook observations for MCP, `apply_patch`, and long-running Bash.
- `docs/codex-upstream-reference.md`, README, workflow guides, and generated Codex docs: clarify update semantics, permission profile boundaries, and `--full-auto` deprecation.

## Sources

- https://developers.openai.com/codex/changelog
- https://developers.openai.com/codex/config-reference
- https://developers.openai.com/codex/subagents
- https://developers.openai.com/codex/cli/reference
- https://github.com/openai/codex/releases/tag/rust-v0.124.0
- https://github.com/openai/codex/releases/tag/rust-v0.125.0
- https://github.com/openai/codex/releases/tag/rust-v0.126.0
- https://github.com/openai/codex/releases/tag/rust-v0.127.0
- https://github.com/openai/codex/releases/tag/rust-v0.128.0
