# Acceptance

- [ ] A single explicit harness/stability contract is encoded in the implementation and regression coverage for the repo's runtime and sync surfaces, including run modes, preflight, validation, session refresh, managed artifact ownership, and cleanup/preservation behavior.
- [ ] Shared internal models or utilities reduce duplicated contract logic across the current runtime-harness, execution, parallel-run, and sync/orchestration paths.
- [ ] The implementation explicitly inherits and preserves the accepted user-facing contracts from `SPEC-016`, `SPEC-022`, `SPEC-023`, `SPEC-025`, and `SPEC-026` unless a contradiction is documented and intentionally resolved.
- [ ] Phase-1 code movement stays bounded to the runtime contract/preflight seam plus one measured repository-scan optimization target; broader decomposition is deferred explicitly.
- [ ] Baseline measurements are captured for `namba project`, `namba sync`, and the preflight/execution setup path of `namba run`, and the optimization work is tied to those measured hotspots.
- [ ] The highest-value redundant repository scans, renderer passes, or broad regeneration paths are reduced without weakening analysis quality, safety checks, or generated-output correctness.
- [ ] Phase-1 extracts the shared runtime contract/preflight seam into a narrower boundary used by the existing execution path, with preserved CLI behavior and generated-output compatibility.
- [ ] The implementation defines explicit extraction seams for runtime contract/preflight, execution orchestration, analysis inventory, renderer surfaces, and manifest/output helpers before broader code movement proceeds.
- [ ] `internal/namba/project_analysis.go`, `internal/namba/parallel_run.go`, `internal/namba/readme.go`, and `internal/namba/templates.go` may remain largely intact in phase-1 as long as their deferred extraction seams are documented explicitly.
- [ ] The implementation lands in bounded, reviewable slices with explicit checkpoints rather than depending on a single all-at-once rewrite.
- [ ] The first delivery boundary is explicit and testable: contract anchors, baseline evidence, an extraction map, at least one low-risk module extraction, and at least one measured hotspot improvement are all present before broader follow-up slices begin.
- [ ] Generated docs and workflow guidance remain renderer-driven, stay source-aligned during self-hosting implementation work in this repository, and describe the hardened contract after the source changes are made true.
- [ ] Regression coverage proves contract invariants, refactor safety, and the targeted performance-sensitive behavior.
- [ ] Validation commands pass
- [ ] Tests covering the new behavior are present
