# SPEC-034 Plan

1. Lock the current-state evidence and measurement baseline:
   - inspect `internal/namba/project_analysis.go`, `internal/namba/namba.go`, `internal/namba/spec_review.go`, and `internal/namba/execution_evidence.go`
   - preserve the existing behavioral anchors in `sync_stability_test.go`, `internal/namba/sync_test.go`, and `internal/namba/project_analysis_inventory_test.go`
   - use `BenchmarkSpec027ProjectCommand` and `BenchmarkSpec027SyncCommand` as the initial benchmark guardrails
2. Introduce an analysis-side reusable index on top of `analysisInventory`:
   - fast path membership and candidate lookup derived from the inventory walk
   - selective on-demand memoized reads for README, `go.mod`, `package.json`, and small entry modules
   - no blanket "cache the whole repo" behavior
   - define the contract explicitly so helper code has one place to ask for membership, candidate selection, and lazy text reads
3. Refactor project-analysis helpers to consume the reusable index:
   - avoid secondary repo-wide discovery when the inventory already contains the needed candidate set
   - avoid repeated reads of the same small source file within one invocation
   - preserve current system detection, evidence ordering, conflict handling, and quality semantics
4. Refactor sync into staged output assembly:
   - build README outputs, analysis outputs, readiness outputs, and support docs first
   - keep one sync-owned manifest session even if the implementation uses a deliberately small fixed batch count
   - document cleanup authority for README-managed paths, project-analysis-managed paths, and explicit staged readiness/support outputs
   - batch manifest mutation and change detection so one sync run does not churn `.namba/manifest.json` once per readiness file when avoidable
   - keep stale managed-output cleanup and no-op modtime guarantees intact
5. Precompute the small shared sync support context once per invocation:
   - latest SPEC id
   - latest review-readiness summary
   - latest execution-proof summary
   - any other derived support-doc facts currently recomputed across builders
6. Add regression coverage for:
   - project-analysis inventory/cache reuse
   - batched readiness refresh across multiple SPECs
   - stable no-op sync behavior after the refactor
   - compatibility with current manifest ownership and stale-file cleanup semantics
   - deterministic counters or equivalent assertions for manifest writes, repeated readiness materialization, and shared support-context discovery reuse
7. Re-run benchmark coverage and compare the refactored `project` and `sync` paths against the current fixture with attention to allocations and avoidable filesystem work.
   - keep benchmarks as secondary evidence, not the only proof
8. Run validation commands.
9. Refresh the relevant review-readiness artifact if implementation details require it; otherwise keep product/engineering/design review tracks pending for follow-up review skills.
