# SPEC-003

## Goal

Harden NambaAI parallel execution so worktree fan-out/fan-in behaves predictably on success and failure, with explicit merge gates and cleanup rules.

## Context

- Project: namba-ai
- Language: go
- Mode: tdd
- Execution mode: `namba run SPEC-XXX --parallel`
- Current gap: worker failures abort execution, but merge, cleanup, and partial-result policy are still implicit

## Requirements

- Record per-worker execution outcomes before any merge attempt.
- Prevent fan-in merges when any worker fails runner execution or validation.
- Distinguish worker execution failure, validation failure, merge failure, and cleanup failure in logs.
- Add an explicit cleanup policy for temporary worktrees and branches after success and failure.
- Keep `--dry-run` behavior unchanged.
- Preserve the existing worktree-based parallel architecture; do not redesign it into a different orchestration model in this SPEC.