# SPEC-001

## Goal

Harden the NambaAI runner core so `namba run` uses an explicit runner abstraction, produces structured execution and validation logs, and separates runner failures from validation failures.

## Context

- Project: namba-ai
- Language: go
- Mode: tdd
- Runner target: codex
- Scope: Runner Core only

## Requirements

- Add a runner abstraction and keep `codex` as the default runner.
- Load runner metadata from `.namba/config/sections/system.yaml`.
- Write structured execution logs and validation reports under `.namba/logs/runs/`.
- Keep `--dry-run` behavior unchanged.
- Reuse the same execution path for serial and parallel runs.
- Do not redesign worktree merge policy in this SPEC.