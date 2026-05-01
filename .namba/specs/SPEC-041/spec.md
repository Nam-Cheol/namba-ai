# SPEC-041

## Problem

Codex CLI `0.124.0` through `0.128.0` introduced meaningful changes in permissions, sandboxing, subagents, plugin/MCP behavior, hooks, app-server surfaces, and CLI update flow. NambaAI is intentionally Codex-native, so the project needs an implementation-ready plan that answers:

- whether NambaAI can safely update its Codex assumptions to this release range
- which changes require NambaAI code, docs, tests, or config updates
- how to preserve the current 5-agent Codex workflow without confusing it with Namba worktree parallelism

## Goal

Prepare NambaAI for Codex CLI `0.124.0`-`0.128.0` by updating the compatibility contract, tests, and operator guidance around Codex execution, permissions, subagents, hooks, plugin/MCP loading, and update semantics.

The working conclusion is: **yes, NambaAI can update to this range**, and the preferred path is capability-based compatibility rather than hard version gating.

## Non-Goals

- Do not vendor or pin the Codex CLI binary inside this repository.
- Do not move user-specific model, auth, web-search, permission-profile, or platform sandbox settings into repo-managed `.codex/config.toml`.
- Do not adopt persisted `/goal` workflows as a required Namba execution primitive in this SPEC; evaluate them as an optional future orchestration surface.
- Do not automatically raise Namba worktree parallelism from 3 to 5 unless that is explicitly accepted as a separate implementation decision.

## Upstream Findings

Five parallel research agents split the range and repo impact:

- Agent 1: `0.124.0`/`0.125.0` release impact.
- Agent 2: `0.126.0`/`0.127.0` tag and compare impact.
- Agent 3: `0.128.0` endpoint and latest stable implications.
- Agent 4: NambaAI local Codex integration points.
- Agent 5: migration, validation matrix, and acceptance framing.

Key upstream changes:

- `0.124.0` adds stable hooks in `config.toml` and managed `requirements.toml`, hook observations for MCP tools, `apply_patch`, and long-running Bash, app-server multi-environment/per-turn cwd selection, remote plugin marketplace list/read, Bedrock provider support, Fast service-tier defaulting, permission drift fixes, MCP approval sync fixes, and `wait_agent` queue timing fixes.
- `0.125.0` adds app-server Unix socket transport, pagination-friendly resume/fork, sticky environments, remote thread config/store plumbing, permission-profile round-tripping across sessions/turns/MCP/shell escalation/app-server APIs, `codex exec --json` reasoning-token reporting, and stricter config/schema behavior around MultiAgentV2 and agent-role config paths.
- `0.126.0` and `0.127.0` are tag-heavy public releases with less expanded official release-page prose than `0.124.0`, `0.125.0`, and `0.128.0`; they should be validated primarily through binary help snapshots and GitHub tag/compare evidence. Relevant deltas include `/goal` groundwork and resume, `codex update`, TUI keymap work, plugin cache/path improvements, external-agent migration/import, hooks/list app-server surfaces, MCP approval policy fixes, Linux workspace metadata protections, and Windows/unified exec fixes.
- `0.128.0` is the stable endpoint for this plan. It adds persisted `/goal` workflows, `codex update`, configurable TUI keymaps, plan-mode nudges, expanded permission profiles and sandbox CLI profile selection, improved plugin marketplace/cache/uninstall/hook enablement/external-agent import flows, external agent session import, explicit MultiAgentV2 thread/wait/root-subagent configuration, managed-network hardening, Bedrock fixes, app-server release artifacts, and deprecates `--full-auto` in favor of explicit sandbox/profile flows.

Primary sources:

- https://developers.openai.com/codex/changelog
- https://developers.openai.com/codex/config-reference
- https://developers.openai.com/codex/subagents
- https://developers.openai.com/codex/cli/reference
- https://github.com/openai/codex/releases/tag/rust-v0.124.0
- https://github.com/openai/codex/releases/tag/rust-v0.125.0
- https://github.com/openai/codex/releases/tag/rust-v0.126.0
- https://github.com/openai/codex/releases/tag/rust-v0.127.0
- https://github.com/openai/codex/releases/tag/rust-v0.128.0

## Local Evidence

- `.codex/config.toml` already uses repo-safe defaults: `approval_policy = "on-request"`, `sandbox_mode = "workspace-write"`, and `[agents] max_threads = 5`.
- `.codex/config.toml` explicitly keeps user-specific models, auth, web search, permission profiles, and platform sandbox choices in user-level Codex config.
- `.namba/config/sections/workflow.yaml` currently sets `max_parallel_workers: 3`; this is Namba-managed git worktree fan-out, not same-workspace Codex subagent capacity.
- `internal/namba/codex_capability.go` already probes `codex --version`, `codex exec --help`, and `codex exec resume --help` instead of relying on a fixed Codex version.
- `internal/namba/hook_runtime.go` already models typed hook observations for `after_patch`, `after_bash`, and `after_mcp_tool`.
- Local binary check during planning reported `codex-cli 0.128.0`.

## Scope

1. Strengthen the Codex capability matrix.
   - Add versioned help fixtures or equivalent regression coverage for `0.124.0`, `0.125.0`, `0.126.0`, `0.127.0`, and `0.128.0`.
   - Cover `codex exec`, `codex exec resume`, `--json`, `--ephemeral`, `-c`, `-m`, `-p`, `-s`, `--add-dir`, approval, sandbox, and config override behavior.
   - Treat `codex exec --json` as a guard only for Namba-owned JSON consumers if implementation identifies one; otherwise narrow the work to documenting additive JSON output tolerance and keep parser tests focused on the owned help/version surfaces.

2. Update operator-facing Codex guidance.
   - Clarify that `codex update` updates the upstream Codex CLI, while `namba update` updates the NambaAI CLI.
   - Replace or explicitly discourage any lingering `--full-auto` guidance in favor of `--sandbox workspace-write`, `approval_policy`, and explicit permission/profile settings.
   - Document that permission profiles remain user-owned unless Namba deliberately widens repo-managed config.

3. Preserve the 5-agent Codex plan.
   - Keep same-workspace Codex subagent capacity at 5 via `.codex/config.toml`.
   - Document that Namba worktree parallelism remains 3 unless a separate acceptance decision raises `max_parallel_workers`.
   - Add tests or docs that prevent these two concurrency knobs from being conflated.
   - Prove the repo-managed `[agents] max_threads = 5` setting remains compatible with the target Codex config schema or local `0.128.0` binary help/config validation surface, not only with Namba template rendering.

4. Validate hooks, plugins, and MCP surfaces.
   - Re-check Namba-owned hook runtime assumptions against Codex stable hook event names and observation payload shape for MCP, `apply_patch`, and long-running Bash.
   - Re-check repo-managed MCP preset boundaries against Codex plugin/MCP cache/path and approval-policy behavior where Namba config actually integrates; do not expand this SPEC into end-to-end marketplace, plugin import, or external-agent validation.

5. Keep `/goal` as future-facing.
   - Record persisted `/goal` workflows as a potential future Namba orchestration candidate.
   - Do not make `/goal` required for `namba run`, `namba plan`, or `namba pr` in this SPEC.

## Risks

- Permission profiles can easily become user-specific policy; adding them to repo defaults too early would make NambaAI override local operator intent.
- MultiAgentV2 configuration changed across the range; tests should prove the current `[agents] max_threads = 5` config is still valid under the target Codex version.
- `0.126.0` and `0.127.0` need binary-help and tag/compare verification because their public release pages are less descriptive.
- Raising worktree parallelism from 3 to 5 would increase merge conflicts, validation time, cleanup failures, and preserved-worktree handling.
- `codex update` and `namba update` are semantically close names; docs and help output must avoid ambiguity.

## Implementation Readiness

This SPEC is ready for implementation after advisory review. The first execution slice should be docs/tests/capability fixtures. Any change to Namba worktree worker count should be split or explicitly accepted inside implementation review before editing templates and defaults.
