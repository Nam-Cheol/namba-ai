# SPEC-027 Phase-1 Contract

## Purpose

Phase-1 keeps SPEC-027 bounded to two concrete outcomes:

1. one shared runtime/preflight contract seam for `namba run`
2. one measured repository-scan optimization in the init helper path

This document is the human-readable contract anchor for that slice. The code and tests remain authoritative.

## Preserved Runtime Contract

- Run-mode normalization stays stable:
  - `default` keeps `stateful`
  - `solo` upgrades `stateful` to `solo`
  - `team` upgrades `stateful` to `team`
  - `parallel` upgrades `stateful` to `parallel-worker`
- Preflight still validates:
  - project root availability
  - `codex` availability and CLI contract compatibility
  - `git` availability for git-backed or parallel runs
  - `add_dirs` resolution
  - required environment presence
  - declared network requirement signaling
- Validation remains part of the runtime contract:
  - request, preflight, execution, and validation artifacts are still written under `.namba/logs/runs/`
  - repair-loop retry count still follows `repair_attempts`
- Session refresh signaling remains source-of-truth in `.namba/logs/session-refresh-required.json`
- Managed-output ownership remains source-of-truth in the existing `writeOutputs` and `replaceManagedOutputs` flow
- Worktree cleanup and preserve-on-failure behavior remain inherited from the existing parallel-run contract

## Landed Code Boundary

- [`internal/namba/runtime_contract.go`](/mnt/c/study/mo-ai/namba-ai/internal/namba/runtime_contract.go)
  - runtime request/config models
  - mode normalization
  - approval/sandbox normalization
  - add-dir resolution
  - system/codex/workflow config loading
  - session refresh and preflight report types
- [`internal/namba/runtime_harness.go`](/mnt/c/study/mo-ai/namba-ai/internal/namba/runtime_harness.go)
  - preflight execution only
- [`internal/namba/execution.go`](/mnt/c/study/mo-ai/namba-ai/internal/namba/execution.go)
  - runner execution, repair loop, validation, and artifact emission
- [`internal/namba/init_scan.go`](/mnt/c/study/mo-ai/namba-ai/internal/namba/init_scan.go)
  - single-walk init repository scan for language/methodology/gofmt target detection

## Regression Anchors

- [`internal/namba/runtime_contract_test.go`](/mnt/c/study/mo-ai/namba-ai/internal/namba/runtime_contract_test.go)
- [`internal/namba/init_scan_test.go`](/mnt/c/study/mo-ai/namba-ai/internal/namba/init_scan_test.go)
- [`internal/namba/execution_test.go`](/mnt/c/study/mo-ai/namba-ai/internal/namba/execution_test.go)
- [`internal/namba/parallel_run_test.go`](/mnt/c/study/mo-ai/namba-ai/internal/namba/parallel_run_test.go)
- [`internal/namba/parallel_test.go`](/mnt/c/study/mo-ai/namba-ai/internal/namba/parallel_test.go)
- [`sync_stability_test.go`](/mnt/c/study/mo-ai/namba-ai/sync_stability_test.go)

## Explicit Non-Goals For This Slice

- No broad `project_analysis.go` decomposition
- No broad `parallel_run.go` redesign
- No mandatory `readme.go` or `templates.go` split
- No user-facing CLI contract changes
