# SPEC-034

## Goal

Reduce avoidable filesystem I/O in the `namba project` analysis pipeline and the `namba sync` path so larger repositories and SPEC-heavy workspaces stay lighter without regressing output fidelity, no-op stability, or end-to-end command latency.

## Command Choice Rationale

- This work changes Namba-owned core command behavior in `project`, `sync`, manifest updates, and review-readiness refresh paths.
- The request is a feature/runtime optimization inside Namba core, so `namba plan` is the correct planning surface.
- This is not a direct bug repair and not a reusable harness-only package.

## Context

- Project: namba-ai
- Project type: existing
- Language: go
- Mode: tdd
- Work type: plan
- Planning surface: `namba plan "<description>"`

## Verified Local Context

- `internal/namba/project_analysis.go` builds a repository inventory with `filepath.WalkDir(...)`, but analysis helpers still re-open selected files with `os.ReadFile(...)` for README/system summaries and related heuristics.
- `internal/namba/namba.go` runs `sync` in multiple write phases: README outputs, `runProject` plus review-readiness refresh, and then support docs.
- `internal/namba/spec_review.go` refreshes each SPEC readiness file independently, and each readiness refresh currently goes through `writeOutputs(...)`, which re-reads and may rewrite `.namba/manifest.json`.
- `internal/namba/execution_evidence.go` discovers the latest execution proof by globbing and reading evidence manifests, and sync-generated support docs currently ask for that information more than once per invocation.
- `sync_stability_test.go` already enforces the no-op sync contract: unchanged managed outputs must preserve both file content and modification times.
- `internal/namba/spec027_phase1_test.go` already contains `BenchmarkSpec027ProjectCommand` and `BenchmarkSpec027SyncCommand`, so this optimization can be benchmarked against an existing fixture instead of inventing a new one from scratch.

## Problem

The current implementation does more disk work than the operator-visible behavior requires.

### 1. Project Analysis Repeats Discovery Work

- `collectAnalysisFiles(...)` already performs the authoritative repository walk for one analysis run.
- After that walk, analysis still performs secondary file reads and, in some code paths, secondary discovery work for entrypoint or summary inference.
- The problem is not that analysis scans the repository once. The problem is that one invocation does not reuse that scan aggressively enough.

### 2. Sync Fragments Managed Writes

- `runSync(...)` materializes README outputs, project-analysis outputs, review readiness outputs, and support docs as separate phases.
- `writeOutputs(...)` loads and updates the manifest for each write batch.
- `refreshAllSpecReviewReadiness(...)` iterates every SPEC and calls `refreshSpecReviewReadiness(...)`, which currently performs its own `writeOutputs(...)` call per SPEC.
- On a repository with many SPEC packages, that turns one `namba sync` invocation into repeated manifest churn and repeated change-detection I/O.

### 3. Sync Support Docs Recompute The Same Derived Context

- Sync support docs need a small set of shared derived facts, such as latest SPEC id, review-readiness state, and latest execution-proof summary.
- Those facts are currently discovered through helper calls that can independently re-scan the same directories or evidence artifacts.
- The result is not catastrophic, but it is avoidable I/O in a path that should stay lightweight.

### 4. "Lighter" Must Not Mean "Less Safe"

- `namba sync` already has strong behavioral expectations around no-op stability, stale managed-file cleanup, and advisory review artifacts.
- A performance-only rewrite that breaks those guarantees would be a regression even if raw I/O counts improved.
- The implementation therefore needs explicit semantic and benchmark guardrails, not only a refactor narrative.

## Desired Outcome

- A single `namba project` invocation builds one authoritative analysis inventory and reuses it through the rest of the analysis/rendering pipeline.
- Analysis may add indexing or memoization, but it should do so selectively and on demand rather than loading the entire repository into memory.
- A single `namba sync` invocation stages its derived outputs more coherently so manifest work and file change detection are batched instead of repeated per sub-step or per SPEC when avoidable.
- Review-readiness refresh stays functionally the same, but its output materialization is aggregated enough that large SPEC sets do not create unnecessary manifest churn.
- Sync support docs consume shared precomputed state for latest SPEC/readiness/execution-proof facts instead of rediscovering the same information in multiple builders.
- Existing output semantics remain stable:
  - no-op sync still preserves file bodies and modification times
  - stale managed outputs still disappear when their source set shrinks
  - analysis findings, confidence labels, and quality gate behavior stay aligned with the same repository facts
- Final validation evidence makes the operator value explicit:
  - which workspace shape improved (`many SPECs`, `large inventory`, or both)
  - which redundant work was removed (duplicate discovery reads, repeated readiness writes, repeated manifest churn, or repeated derived-context scans)
- Benchmark and regression coverage make it obvious if the optimization accidentally makes `project` or `sync` slower or more allocation-heavy.

## Scope

- Introduce an analysis-side reusable inventory/index layer for path membership, candidate lookup, and selective on-demand content reads.
- Refactor project-analysis helpers to consume that layer instead of launching avoidable secondary discovery or duplicate file reads.
- Refactor sync output assembly so README, analysis docs, review readiness files, and support docs can be written through fewer manifest/update cycles.
- Reduce repeated artifact discovery for sync-generated summaries by precomputing the small shared context that multiple support-doc builders need.
- Add or update regression tests and benchmark coverage for the optimized paths.

## Non-Goals

- Do not redesign the analysis document format, evidence wording, or review-readiness product semantics beyond what is necessary for the I/O refactor.
- Do not introduce a long-lived daemon, background indexer, or cross-command persistent cache.
- Do not trade filesystem savings for an unbounded "read every file into memory" approach.
- Do not weaken managed-output cleanup, no-op sync stability, validation requirements, or manifest ownership semantics.
- Do not change `namba run`, harness classification, or unrelated execution-evidence semantics in this slice.

## Design Constraints

- Keep `.namba/` as the source of truth for generated project state and managed outputs.
- Prefer one-command-local indexing and memoization over persistent cache invalidation complexity.
- Any content cache must stay bounded and selective:
  - only for files that the active analysis/sync pass actually needs
  - no blanket buffering of the entire repository
- Preserve deterministic output ordering so generated docs and tests remain stable.
- Keep performance verification repository-local and deterministic by reusing the existing benchmark fixtures where possible.

### Analysis Index Boundary

- One `namba project` invocation gets one analysis-owned reusable index derived from the authoritative inventory walk.
- That index owns:
  - normalized path membership checks
  - candidate lookup by exact path, basename, and system-root scope
  - bounded lazy text reads for only the small source files the current analysis pass actually needs
- Eligible memoized reads are limited to lightweight text inputs such as README files, `go.mod`, `package.json`, and small entry modules already present in the collected inventory.
- Analysis helpers that currently perform ad hoc lookup or repeated `os.ReadFile(...)` calls should consume this boundary instead of inventing parallel caches or rescans.

### Sync Staged-Write Boundary

- `namba sync` should stage all derived outputs before mutation rather than interleaving discovery and writes per helper.
- The sync path should own one manifest session for the command:
  - load manifest once
  - compute stale removals once
  - materialize staged outputs
  - flush manifest once at the end of the sync-owned write path
- Cleanup authority stays explicit:
  - README cleanup remains governed by the README-managed path set
  - project-analysis cleanup remains governed by the project-analysis-managed path set
  - readiness and support docs participate as explicit staged outputs, not as separately rediscovered per-SPEC write calls
- A deliberately small fixed batch count is acceptable only if one manifest session is preserved and each batch has a documented ownership reason.

### Deterministic I/O Verification

- Benchmarks remain secondary guardrails.
- Primary proof should come from deterministic regression tests that can count or otherwise assert reduced repeated work using existing `App` seam points such as file-read overrides, manifest-write overrides, or equivalent narrow instrumentation.
- The SPEC should prove, at minimum:
  - repeated analysis helper reads are collapsed within one invocation
  - multi-SPEC readiness refresh is not materialized through one manifest/update cycle per SPEC
  - shared sync support-doc context is discovered once per sync invocation rather than once per builder

## Implementation Priority

1. Lock the current I/O map and identify the avoidable extra scans/manifest updates in `project` and `sync`.
2. Introduce the reusable analysis inventory/index contract before touching analysis heuristics.
3. Batch sync output materialization and manifest updates without breaking stale-file cleanup or no-op stability.
4. Precompute shared sync support-doc context instead of rediscovering it in multiple builders.
5. Add regression and benchmark coverage before calling the refactor complete.
