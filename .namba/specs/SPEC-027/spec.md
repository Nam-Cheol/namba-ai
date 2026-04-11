# SPEC-027

## Goal

Stabilize NambaAI's expanded runtime and planning surfaces by tightening the harness contract, reducing avoidable repository-scan and regeneration overhead, and refactoring oversized internals into clearer modules that are cheaper to evolve safely.

## Context

- Project: namba-ai
- Project type: existing
- Language: go
- Mode: tdd
- Work type: plan
- Planning surface: `namba plan "<description>"`
- Command choice rationale: this stays under the feature-planning surface because the outcome is a repo-wide stabilization program, not a new reusable harness-only scaffold or a direct repo-local skill/agent artifact.
- Verified local context as of 2026-04-09:
  - `internal/namba/runtime_harness.go` owns runtime config loading, add-dir resolution, and preflight checks, but execution semantics are still distributed across `internal/namba/execution.go`, `internal/namba/parallel_run.go`, and command orchestration in `internal/namba/namba.go`.
  - `internal/namba/execution.go` defines the core request/result/validation models, runner selection, and execution flow, so the standalone runtime contract is already substantial but not yet isolated behind one narrow boundary.
  - `internal/namba/parallel_run.go` mixes worktree setup, prompt emission, worker execution, merge policy, and cleanup policy in one path.
  - `internal/namba/project_analysis.go` performs repository-wide file collection via `filepath.WalkDir` and then renders the `.namba/project/*` outputs from that inventory.
  - `internal/namba/namba.go` is 4051 lines and still combines command parsing, SPEC scaffolding, repo detection, sync orchestration, manifest writes, discovery helpers, and release/PR flows in one file.
  - `internal/namba/readme.go` is 2113 lines and `internal/namba/templates.go` is 1550 lines, so generated documentation and instruction-surface rendering are also concentrated in very large source files.
  - `internal/namba/namba.go` still contains repeated repository scans in helpers such as `detectMethodology`, `treeContainsExtension`, `directoryContainsGo`, and several discovery/build paths, while `internal/namba/project_analysis.go` performs its own full-tree walk.
  - `internal/namba/namba.go` `runSync` rebuilds README outputs, reruns `namba project`, refreshes review readiness, and writes project support docs in one orchestration path, which is safe but increasingly expensive as the repo grows.
  - `.namba/project/quality-report.md` still warns that mismatch/conflict heuristics are intentionally narrow, so the current planning-quality signal is not yet strong enough to double as a broader harness-quality contract.
  - Relevant prior slices already exist and should be inherited rather than reopened:
    - `SPEC-016`: runtime harness redesign and execution contract expansion
    - `SPEC-022`: evidence-based `namba project` analysis foundation
    - `SPEC-023`: top-level help and argument contract hardening
    - `SPEC-025` and `SPEC-026`: `$namba-create` contract and generator ownership model
  - Self-hosting note: when implementation work in this repository needs generated docs or synced artifacts to match current source rather than an installed release, use a repo-source-aligned Namba entrypoint such as `go run ./cmd/namba <command>` or an equivalent verified local binary.

## Target Reader

- Namba maintainers who need the runtime and planning surfaces to stay coherent as the repository grows.
- Reviewers who need a safe, sliceable refactor program rather than another monolithic "cleanup" effort.
- Operators who want `namba project`, `namba sync`, and `namba run` to stay predictable and fast enough as more generated surfaces and workflow features accumulate.

## Problem

Recent SPECs increased NambaAI's capability, but they also stretched the repository across more runtime, planning, generation, and documentation surfaces. The remaining gap is no longer a single missing feature. It is a repo-wide stabilization program with three linked risks:

1. Harness contract drift
   - Runtime invariants for run modes, preflight, validation, session refresh, worktree cleanup/preservation, and managed output ownership are spread across multiple files and helper paths.
   - The behavior is mostly correct, but the contract is harder to review, test, and evolve as one unit.
2. Hidden performance costs
   - Core flows still rely on repeated full-tree scans and broad regeneration passes.
   - The current repository size keeps that acceptable for now, but the cost will rise as generated artifacts, specs, and supported surfaces grow.
3. Structural refactor debt
   - `namba.go`, `project_analysis.go`, `readme.go`, and `templates.go` have become high-churn hotspots.
   - Safe changes now require editing very large files that mix orchestration, discovery, rendering, and policy.
4. Delivery-shape mismatch
   - The next step should not be framed as one oversized implementation PR.
   - Namba needs an explicit multi-workstream plan that defines contract boundaries, performance baselines, refactor seams, and validation strategy before more debt accumulates.

## Desired Outcome

- A documented and testable harness contract exists for the repo's runtime and sync surfaces, explicitly covering:
  - `namba run` mode invariants
  - preflight expectations
  - validation pipeline behavior
  - session-refresh signaling
  - managed artifact ownership and write behavior
  - worktree cleanup/preservation rules
- Shared internal models and utilities reduce duplicated contract logic across `execution.go`, `runtime_harness.go`, `parallel_run.go`, `project_analysis.go`, and `namba.go`.
- Performance baselines are captured for at least:
  - `namba project`
  - `namba sync`
  - the preflight/execution setup path of `namba run`
- Optimization work is tied to measured hotspots rather than intuition, with the highest-value redundant scans and regeneration paths addressed first.
- Oversized sources are decomposed into narrower modules so command parsing, SPEC scaffolding, runtime harnessing, project analysis, docs rendering, and manifest/output handling can evolve with lower regression risk.
- The program is executable in bounded slices, so maintainers can land the work across follow-up phases or PRs without losing contract clarity or repository safety.

## V1 Success Definition

- The first delivery does not try to "finish the whole refactor."
- The first delivery is considered successful when it leaves the repository with:
  - one explicit harness/stability contract and matching regression anchors
  - reproducible baseline measurements for `namba project`, `namba sync`, and `namba run` setup cost
  - a documented extraction map for the major monolith files
  - at least one low-risk module extraction landed behind preserved CLI behavior
  - at least one measured hotspot improvement landed with before/after evidence
- Broader decomposition, broader optimization passes, or further contract cleanup can then proceed as follow-up slices without re-litigating the first delivery boundary.

## Scope

- Define a v2 harness/stability program for:
  - runtime contract normalization across `internal/namba/execution.go`, `internal/namba/runtime_harness.go`, `internal/namba/parallel_run.go`, and command orchestration in `internal/namba/namba.go`
  - performance baselining and targeted optimization for repo scanning, project analysis, sync regeneration, and execution preparation
  - modular refactor of monolithic internal packages and renderer sources
  - regression and observability gates that prevent contract drift after the refactor
- Produce an implementation decomposition that identifies:
  - which behaviors are global contracts versus local helpers
  - which hotspots need measurement before change
  - which large files can be split first with the lowest regression risk
  - which follow-up slices can run in parallel and which remain on the critical path
- Define the first-pass extraction seams explicitly, at minimum for:
  - runtime contract and preflight policy
  - execution orchestration and run-mode behavior
  - analysis inventory and system detection
  - renderer and docs-generation surfaces
  - manifest/output sync and SPEC scaffolding helpers
- Update generated documentation or workflow guidance only as a consequence of true source changes, not by hand-editing generated outputs.

## Inherited Contracts

- `SPEC-016` remains the source of truth for current standalone runtime-harness semantics such as run modes, repair loops, session refresh signaling, and parallel worker expectations.
- `SPEC-022` remains the source of truth for evidence-backed project analysis, mismatch reporting, and analysis quality gates.
- `SPEC-023` remains the source of truth for top-level help and argument safety.
- `SPEC-025` and `SPEC-026` remain the source of truth for `$namba-create`, manifest ownership, and user-authored output preservation.
- `SPEC-027` is therefore a consolidation and stabilization program. It may tighten seams, remove duplication, and add measurement or extraction boundaries, but it should not casually relitigate those accepted user-facing contracts unless a contradiction is documented and intentionally resolved.

## Workstreams

### 1. Harness Contract Hardening

- Normalize and document invariants for run modes, preflight, validation, session refresh, cleanup/preservation, and managed artifact ownership.
- Remove or isolate behavior drift where similar concepts are encoded in multiple files with different assumptions.
- Expand regression anchors so contract changes fail loudly.

### 2. Performance Foundation

- Establish reproducible baseline measurements for representative commands and repository states.
- Inventory repeated filesystem walks, manifest reads/writes, and renderer passes.
- Optimize the highest-value hotspots first, especially repo scans reused across `project`, `sync`, and discovery paths.

### 3. Structural Refactor

- Decompose monolithic files into narrower modules without changing public CLI semantics.
- Separate renderer concerns from orchestration concerns.
- Make internal models reusable across commands instead of continuing to grow `namba.go`.

### 4. Delivery Safety

- Keep generated artifact ownership, session refresh signaling, and validation behavior stable during the refactor.
- Keep each slice reviewable, rollback-safe, and shippable on its own.

## Implementation Priority

1. Lock the harness/stability contract and its regression anchors.
2. Capture baselines and hotspot evidence before optimization.
3. Extract low-risk modules from monolithic files.
4. Land targeted performance improvements behind the new seams.
5. Refresh docs, reviews, and synced outputs after behavior becomes true.

## Initial Delivery Boundary

- First slice deliverables:
  - contract definitions and regression anchors
  - baseline measurement path and saved evidence
  - extraction map for `namba.go`, `project_analysis.go`, `readme.go`, and `templates.go`
  - one low-risk extraction proving the refactor seam
  - one measured optimization proving the performance path
- Follow-up slices may expand extraction breadth or optimization depth, but they should not reopen the v1 contract unless a validated contradiction appears.

## Phase-1 Focus

- Phase-1 is intentionally narrower than the full stabilization program.
- Phase-1 code movement is limited to one primary extraction seam:
  - runtime contract and preflight helpers currently spread across `internal/namba/runtime_harness.go`, `internal/namba/execution.go`, and the execution request setup path in `internal/namba/namba.go`
  - target candidates include `codexConfig`, `workflowConfig`, `resolveCodexRuntimeForMode`, `resolveRuntimeAddDirs`, `runPreflight`, `newExecutionRequest`, and the request/preflight models directly coupled to them
- Phase-1 optimization is limited to one measured repository-scan hotspot:
  - prefer the smallest safe reduction of duplicated filesystem walks in the current helper path around `detectMethodology`, `treeContainsExtension`, `directoryContainsGo`, or an equivalent shared-inventory seam in `internal/namba/namba.go`
  - the optimization must come with before/after evidence, not just a claimed cleanup
- Phase-1 must not expand into broad architectural movement outside those seams unless a concrete contradiction blocks the extraction.
- Phase-1 still produces the longer-term extraction map, but it does not have to fully decompose every large file.

## Deferred From Phase-1

- Do not fully decompose `internal/namba/project_analysis.go` in phase-1.
- Do not broadly rewrite `internal/namba/parallel_run.go` in phase-1 beyond adopting shared runtime-contract or preflight helpers where needed.
- Do not treat `internal/namba/readme.go` or `internal/namba/templates.go` as mandatory code-movement targets in phase-1; document their future extraction seams instead.
- Do not redesign the overall `runSync` pipeline in phase-1 beyond the minimum source-aligned behavior and any shared helper adoption required by the chosen extraction seam.

## Non-Goals

- Do not attempt a one-PR rewrite of the whole repository.
- Do not redesign user-facing command taxonomy such as `plan` vs `harness` vs `fix` unless a concrete contract conflict requires it.
- Do not weaken existing safety properties such as preserve-on-failure worktrees, manifest ownership separation, or session-refresh signaling.
- Do not optimize based only on perceived slowness without baseline evidence.
- Do not hand-edit generated README or workflow outputs as a substitute for fixing renderer sources.
- Do not treat the first implementation slice as a promise to finish every optimization and refactor target in one pass.

## Design Constraints

- Keep `.namba/` as the source of truth for generated workflow state and review artifacts.
- Preserve existing CLI semantics while tightening the internal contract behind them.
- Treat measured evidence and executable code as stronger than stale docs or intuition.
- Prefer extraction and reuse over continued helper duplication.
- When this repository self-hosts the Namba CLI during implementation, keep generated outputs source-aligned rather than trusting an unrelated installed release blindly.
- Keep the refactor reviewable: each slice should leave the repository in a validated, shippable state.
