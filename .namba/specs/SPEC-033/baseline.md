# SPEC-033 Baseline Evidence

## Purpose

Capture the current repository state before introducing a typed execution-evidence bridge for core harness work.

## Structured Today

| Surface | What already exists | Evidence |
| --- | --- | --- |
| Harness classification | typed `harness-request.json` model, route selection, required evidence/reviews | `internal/namba/harness_contract.go` |
| Plan readiness | advisory review readiness plus harness evidence summary | `internal/namba/spec_review.go` |
| Runtime artifacts | request, preflight, execution, validation JSON artifacts | `internal/namba/execution.go` |
| Parallel runtime observability | append-only progress JSONL and final parallel report | `internal/namba/parallel_progress.go`, `internal/namba/parallel_run.go` |
| Browser capability preset | Playwright MCP is available as a repo-managed preset when relevant | `.namba/config/sections/codex.yaml`, `.codex/config.toml` |

## Missing Bridge

| Gap | Why it matters | Evidence |
| --- | --- | --- |
| No canonical execution-evidence manifest | consumers must know artifact-specific file names and semantics | `internal/namba/execution.go`, `internal/namba/parallel_progress.go` |
| No typed optional slot for browser artifacts | browser proof cannot be referenced consistently when it exists | `internal/namba/harness_contract.go` evidence enum, current generated guidance |
| No explicit relationship between plan readiness and post-run proof | current readiness is pre-implementation while runtime artifacts are post-run | `internal/namba/spec_review.go` |
| No single advisory surface for latest execution proof | sync/readiness can show planning evidence, but not one canonical run-proof index | current `.namba/project/*` docs and readiness rendering |

## Planning Implication

The right next slice is not "make browser evidence universally required."

The right next slice is to define one typed execution-evidence contract that:

- reuses current runtime artifacts
- keeps browser evidence optional
- keeps readiness and execution-proof phases distinct
