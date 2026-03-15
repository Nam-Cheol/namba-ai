# SPEC-002

## Goal

Wire `.namba/config/sections/system.yaml` approval and sandbox settings into actual `codex exec` invocations so runtime behavior matches NambaAI configuration.

## Context

- Project: namba-ai
- Language: go
- Mode: tdd
- Runner target: codex
- Scope: Codex execution flags and config validation

## Requirements

- Stop using `--full-auto` for non-interactive execution.
- Pass approval mode with `-a <mode>` and sandbox mode with `-s <mode>`.
- Keep defaults aligned with NambaAI templates: `on-request` and `workspace-write`.
- Reject unsupported `approval_mode` and `sandbox_mode` values with explicit errors.
- Record the effective approval and sandbox values in execution logs.
- Reuse the same behavior for serial and parallel execution paths.