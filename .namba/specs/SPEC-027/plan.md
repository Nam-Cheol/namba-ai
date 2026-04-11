# SPEC-027 Plan

1. Refresh project context and confirm the current repo signals from:
   - `.namba/project/*`
   - `internal/namba/runtime_harness.go`
   - `internal/namba/execution.go`
   - `internal/namba/parallel_run.go`
   - `internal/namba/project_analysis.go`
   - `internal/namba/namba.go`
   - `internal/namba/readme.go`
   - `internal/namba/templates.go`
2. Lock the program boundary before implementation starts:
   - treat this as a repo-wide stabilization/refactor program, not a one-off cleanup
   - inherit `SPEC-016`, `SPEC-022`, `SPEC-023`, `SPEC-025`, and `SPEC-026` where their contracts already stand
   - preserve current public CLI semantics unless a contract conflict makes change unavoidable
3. Define the explicit harness contract to protect during the work:
   - run-mode invariants
   - preflight behavior
   - validation pipeline behavior
   - session-refresh signaling
   - managed output ownership and write behavior
   - worktree cleanup/preservation rules
4. Narrow phase-1 before implementation starts:
   - primary extraction seam: runtime contract and preflight helpers shared by `runtime_harness.go`, `execution.go`, and the execution request setup path in `namba.go`
   - single measured optimization target: one duplicated repository-scan hotspot in the current helper path, preferably around `detectMethodology`, `treeContainsExtension`, or `directoryContainsGo`
   - explicit phase-1 deferrals: broad `project_analysis.go` decomposition, broad `parallel_run.go` redesign, and large renderer splits stay out of scope
5. Capture baseline measurements for the highest-value command paths, at minimum:
   - `namba project`
   - `namba sync`
   - `namba run` preflight/execution setup
6. Inventory repeated repository walks, renderer passes, manifest I/O, and orchestration duplication so optimization work is evidence-based, but only select one hotspot for phase-1.
7. Implement the phase-1 extraction seam:
   - extract the shared runtime contract/preflight helpers into a narrower internal module or file boundary
   - adopt the extracted seam in the existing execution path without changing public CLI semantics
8. Implement the phase-1 measured optimization:
   - reduce one duplicated repository-scan path
   - record before/after evidence in a durable checked-in or reproducible form
9. Design the structural extraction map for the current monoliths, especially:
   - `internal/namba/namba.go`
   - `internal/namba/project_analysis.go`
   - `internal/namba/readme.go`
   - `internal/namba/templates.go`
   - define concrete target seams for runtime contract/preflight, execution orchestration, analysis inventory, renderers, and manifest/output helpers
10. Keep an explicit inherit-vs-reopen checklist for the existing contracts from `SPEC-016`, `SPEC-022`, `SPEC-023`, `SPEC-025`, and `SPEC-026` so later slices do not re-open settled behavior accidentally.
11. When generated docs or sync outputs must match the current repository source, use a repo-source-aligned Namba invocation instead of assuming the installed CLI is current.
12. Add or strengthen regression coverage for:
   - harness contract invariants
   - repeated-scan or regeneration regressions
   - refactor-safe behavior preservation
   - generated-output ownership and session-refresh safety
13. Exit phase-1 only when all of the following are true:
   - the shared runtime contract/preflight seam is real in code
   - one measured hotspot improvement is landed
   - deferred work is written down explicitly
   - validation still passes
14. Run the relevant review passes under `.namba/specs/SPEC-027/reviews/` and refresh the readiness summary when the implementation design stabilizes.
15. Run validation commands.
16. Sync artifacts with `namba sync` or an equivalent source-aligned invocation when self-hosting drift would otherwise make the outputs stale.
