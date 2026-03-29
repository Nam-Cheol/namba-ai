# SPEC-017

## Problem

Codex CLI capability mismatch causes `namba run --team` and `namba run --parallel` to fail late because `codex exec` and `codex exec resume` support different option surfaces on the installed CLI version.

## Goal

Apply the smallest safe fix that makes Namba detect Codex runtime capability up front, map supported runtime controls safely, and fail before team turns or parallel worktree fan-out when the local Codex CLI cannot represent the requested contract.

## Context

- Project: namba-ai
- Project type: existing
- Language: go
- Mode: tdd
- Work type: fix
- Affected area: `internal/namba/execution.go`, `internal/namba/runtime_harness.go`, `internal/namba/parallel_run.go`
- Local evidence:
  - `codex exec` accepts `-c`, `-m`, `-s`, `-p`, `--add-dir`
  - `codex exec` rejects `-a` and `--search`
  - `codex exec resume` accepts `-c` and `-m`
  - `codex exec resume` rejects `-s`, `-p`, and `--add-dir`
