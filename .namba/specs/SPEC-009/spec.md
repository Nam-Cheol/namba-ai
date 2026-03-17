# SPEC-009

## Goal

Add `namba run SPEC-XXX --solo` and `namba run SPEC-XXX --team` so the standalone runner can explicitly target Codex subagent workflows without overloading the existing worktree-based `--parallel` mode.

## Context

- Project: namba-ai
- Project type: existing
- Language: go
- Mode: tdd
- Work type: plan
- Current `run` behavior supports only the default execution path and `--parallel`, where `--parallel` means git worktree fan-out/fan-in with merge gates and cleanup policy.
- Current generated Codex guidance still describes repo-local `.codex/agents/*.md` role cards, while the latest official Codex subagent documentation describes built-in subagents and project-scoped custom subagents under `.codex/agents/*.toml`.

## Problem

NambaAI does not currently expose an explicit CLI mode for Codex subagent execution. Users who want a single delegated subagent flow or a multi-subagent team flow have no first-class `run` option, and `--parallel` cannot be repurposed because it already means worktree orchestration with branch merge semantics.

## Desired Outcome

- `--solo` means a standalone `namba run` execution request that explicitly asks Codex to use a single subagent-oriented workflow inside one workspace.
- `--team` means a standalone `namba run` execution request that explicitly asks Codex to coordinate multiple subagents in one run.
- `--parallel` keeps its current meaning: multiple git worktrees, worker request fan-out, merge only after all workers and validators pass, and preserve failures for inspection.
- User-facing docs clearly separate subagent orchestration from worktree parallelism.

## Scope

- Add CLI flag parsing and execution-mode selection for `--solo` and `--team`.
- Update execution request construction so Codex receives mode-specific instructions for default, solo, and team runs.
- Define and enforce invalid flag combinations, especially conflicts with `--parallel`.
- Update generated docs and guidance to reflect current Codex subagent terminology and how Namba maps to it.
- Add regression tests for flag parsing, execution request generation, and documentation output.

## Non-Goals

- Do not redesign or weaken the existing `--parallel` worktree architecture in this SPEC.
- Do not attempt to recreate Claude Agent Teams lifecycle events inside Codex.
- Do not require a full migration of every existing `.codex/agents/*.md` asset unless it is necessary for the new `run` modes to be correct and documented.

## Design Constraints

- Treat the latest official Codex subagent documentation as the product-level source of truth for subagent semantics.
- Keep `--solo` and `--team` explicit and opt-in; default `namba run SPEC-XXX` should keep current behavior.
- Prefer a narrow first version: incompatible combinations should fail fast with clear errors rather than silently picking one mode.
