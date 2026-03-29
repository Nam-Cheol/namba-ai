# SPEC-016

## Goal

Redesign the standalone `namba run` Codex runtime harness so Namba uses Codex as a configurable, stateful, executable collaboration runtime instead of a thin one-shot command wrapper.

## Context

- Project: namba-ai
- Project type: existing
- Language: go
- Mode: tdd
- Work type: plan
- Verified local context as of 2026-03-29:
  - `internal/namba/execution.go` defines `executionRequest` with only `runner`, `approval_policy`, `sandbox_mode`, and `delegation_plan` beyond prompt/workdir identity.
  - `buildCodexExecArgs` currently reduces every standalone execution to `codex exec -a <approval> -s <sandbox> <prompt>`.
  - `internal/namba/namba.go` still has a separate `runCodexExec` helper that hard-codes `codex exec --full-auto <prompt>`, creating a second execution meaning outside the main runner contract.
  - `internal/namba/agent_runtime.go` records role-specific `model` and `model_reasoning_effort`, but the standalone runner only renders those values into prompt text; it does not pass them into the Codex CLI invocation.
  - `codexRunner.Execute` records `DelegationObserved: false` for every run, so current `--team` semantics are advisory rather than executable runtime delegation.
  - `internal/namba/parallel_run.go` fans work into worktrees but executes each worker sequentially in a `for` loop, even though the mode is exposed as `--parallel`.
  - `runValidationReport` currently supports only `test`, `lint`, and `typecheck`.
  - Repo-generated Codex assets already distinguish repo-safe defaults under `.codex/config.toml` from user-specific settings, but the execution adapter does not yet expose the broader Codex CLI control surface needed to exploit that separation during a run.

## Problem

Namba is directionally Codex-native, but the current runtime harness is still too shallow in four ways:

1. Execution contract drift: `.codex/config.toml`, `.namba/config/sections/system.yaml`, helper shortcuts, and direct CLI flags do not resolve through one explicit runtime contract, so effective execution behavior is harder to reason about and debug.
2. Stateless recovery: the standalone runner executes once, validates once, and exits, so failed implementation runs cannot repair and re-validate inside one continuous Codex context.
3. Non-executable delegation: team routing heuristics and runtime role profiles exist mostly as prompt decoration, while actual runner behavior remains single-process and reports no delegation.
4. Fake parallelism: `--parallel` creates worktree fan-out but still performs execution serially, so the operational contract and runtime behavior diverge.

This leaves Namba underusing Codex as an actual collaboration runtime and makes failures, retries, observability, and configuration semantics harder than they need to be.

## Desired Outcome

- Namba has one explicit execution contract for standalone Codex runs, with repo defaults plus per-run overrides mapped onto real Codex CLI capabilities.
- Stateful execution can keep one Codex session alive across implement, validate, repair, and revalidate cycles, with bounded retry behavior and durable logs.
- `--solo` remains the single-runner, single-workspace standalone path.
- `--team` clearly means same-workspace multi-agent execution with real executable delegation semantics rather than only prompt guidance.
- `--parallel` clearly remains the multi-worktree fan-out/fan-in path and runs worker executions concurrently up to a configured limit while preserving today's safe merge and failure preservation behavior.
- Session-bound instruction changes are surfaced clearly, so users know when regenerated `AGENTS.md` or `.codex/agents/*.toml` content requires a fresh Codex session.
- Preflight, validation, and observability cover runtime environment issues and effective Codex execution metadata, not only test/lint/typecheck output.
- Opt-in live smoke coverage can catch compatibility regressions against the real Codex CLI without making the default test suite network-dependent.

## Scope

- Expand the standalone execution contract to include the Codex runtime controls Namba needs per run, including at minimum `model`, `profile`, `web_search`, `add_dirs`, and `session_mode`, plus any supporting metadata required for logging and retry control.
- Define clear precedence so `.codex/config.toml` stays the repository-owned baseline while Namba contributes only explicit per-run overrides and runtime metadata.
- Remove the split-brain between the main runner path and helper shortcuts by routing both through one shared Codex execution builder and one shared execution meaning.
- Design and implement a stateful run controller that can execute `implement -> validate -> repair -> revalidate` inside one continuous Codex session, with bounded retries and persisted session metadata.
- Preserve a simple user-facing mode model: `--solo` means one runner in one workspace, `--team` means same-workspace multi-agent execution, and `--parallel` means multi-worktree execution.
- Make standalone `--team` materially affect execution through real same-workspace child execution semantics, runtime prompts, and role-specific Codex settings rather than prompt-only heuristics.
- Preserve the sandbox and approval semantics inherited from the parent session, while shaping role behavior through model/profile selection, prompts, and work partitioning instead of imaginary per-role permission changes.
- Rework `--parallel` so worktree workers execute concurrently with a configurable cap, then merge only after every worker execution and validation pass; retain today's preserve-on-failure behavior.
- Detect when commands such as `namba regen` or `namba sync` mutate instruction-surface files that existing Codex sessions may not automatically absorb, and emit an explicit session refresh signal.
- Add preflight checks for runtime prerequisites such as project root, git/Codex availability, safe working directory expectations, required env vars, extra directories, and declared network needs before execution starts.
- Generalize validation into a pipeline that still supports `test`, `lint`, and `typecheck` but can also express extra steps such as build, migration dry-run, smoke start, and output-contract verification where configured.
- Enrich run logs and JSON artifacts with effective runtime metadata such as resolved model/profile, search mode, additional directories, retry count, delegation mode, session identifier, and validation pipeline details.
- Add opt-in live smoke coverage, guarded by an environment switch such as `CODEX_SMOKE=1`, that runs a real Codex execution in a temporary repository to catch runtime compatibility drift.
- Update `README.md`, localized README bundles, and workflow guidance anywhere the runtime semantics for `--solo`, `--team`, `--parallel`, repair loops, or session refresh requirements become user-visible contract.

## Non-Goals

- Do not generalize Namba into a multi-runner abstraction beyond Codex at this stage.
- Do not spend this SPEC on improving keyword routing accuracy beyond what is necessary to support executable delegation.
- Do not redesign the entire Namba planning, PR, or landing workflow outside the execution/runtime surfaces touched by this work.
- Do not weaken the current safety property that failed parallel workers keep their worktrees and branches for inspection.
- Do not make live Codex smoke tests mandatory for default local or CI validation.

## Design Constraints

- Keep `.namba/` as the source of truth for repo-owned workflow state and generated artifacts.
- Keep `.codex/config.toml` limited to repository-safe defaults; user-specific auth, personal preferences, and global Codex behavior remain outside repo-managed state.
- Make runtime overrides explicit and serializable so execution logs explain the effective Codex behavior for a run.
- Keep parent approval and sandbox semantics authoritative when delegating through Codex child sessions or subagents.
- Ensure headless/non-interactive execution fails clearly when an action would require approval that cannot be granted in that mode.
- Keep offline/default tests deterministic; any real Codex invocation must remain opt-in.
- Preserve backward compatibility where reasonable for current `namba run`, `--solo`, `--team`, and `--parallel` entrypoints while tightening the semantics behind them and making their differences explicit in user-facing docs.
