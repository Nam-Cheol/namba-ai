# Engineering Review

- Status: approved
- Last Reviewed: 2026-03-29
- Reviewer: Codex
- Command Skill: `$namba-plan-eng-review`
- Recommended Role: `namba-planner`

## Focus

- Lock architecture, sequencing, failure modes, trust boundaries, and validation strategy before execution starts.

## Findings

- The runtime direction is correct, but config ownership must follow the repository's current truth model. In this repo, `.namba/config/sections/*.yaml` is authoritative and `.codex/config.toml` is a generated Codex-facing baseline, so the implementation must not treat both as independent runtime inputs.
- The repair-loop requirement is valuable, but the SPEC should not assume that plain `codex exec` automatically provides native multi-turn session continuity. The implementation needs a session-controller boundary that can log whether continuity is native, degraded, or unavailable.
- The repo already has two different concurrency concepts: `.codex/config.toml` exposes `[agents].max_threads`, while the scaffolded workflow config defines `max_parallel_workers`. The SPEC should keep Codex child-session/subagent concurrency separate from git worktree parallelism.
- The preflight and validation expansion is justified, but it should remain a pipeline extension of today's `test`/`lint`/`typecheck` defaults instead of becoming a broad release-automation rewrite.

## Decisions

- Keep `.namba/config/sections/*.yaml` as the repo-owned source of truth. Resolve repo defaults from `.namba/config/sections/system.yaml` and `.namba/config/sections/codex.yaml`, emit `.codex/config.toml` as the generated baseline, and persist the fully resolved per-run overlay in execution request/result artifacts.
- Keep the mode contract intuitive: `--solo` remains one runner in one workspace, standalone `--team` becomes the same-workspace orchestrated team path, and `--parallel` stays the multi-worktree path. The interactive multi-agent path remains the Codex-native in-session workflow defined by `AGENTS.md` and repo skills, not a renamed `--team`.
- Introduce a session-controller abstraction and land one implementation backend in this SPEC: standalone Codex orchestration with bounded implement/validate/repair/revalidate retries. Future MCP-native continuation can plug into the same boundary later without changing the outer contract.
- Keep concurrency ownership separate. Use workflow configuration for worktree `--parallel` worker caps, and keep `.codex` agent thread settings scoped to Codex child-session or subagent concurrency only.
- Keep validation pipeline growth incremental. Preserve current default validation behavior, then add opt-in extra steps such as build, migration dry-run, smoke start, or output-contract verification through configuration rather than hard-coded global phases.

## Follow-ups

- Extend config loading so workflow parallel caps and Codex runtime defaults are available through one resolved execution contract before implementation begins.
- Make run logs and docs explicit about whether a repair loop used native continuous session semantics or a degraded orchestration fallback.
- Update request/result JSON fixtures and documentation carefully so existing tooling that reads run artifacts does not silently break.
- Update README and workflow-guide generation together with the runtime contract so CLI naming does not drift from implementation again.

## Recommendation

- Advisory recommendation: approved. Proceed with implementation using the decisions above; avoid re-opening config precedence, session-controller boundaries, or concurrency ownership during the first implementation pass.
