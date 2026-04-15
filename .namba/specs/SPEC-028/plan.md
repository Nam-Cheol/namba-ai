# SPEC-028 Plan

1. Refresh project context and verify the current parallel-runtime evidence from:
   - `.namba/project/*`
   - `internal/namba/parallel_lifecycle.go`
   - `internal/namba/parallel_run.go`
   - `internal/namba/execution.go`
   - `internal/namba/namba.go`
   - `internal/namba/parallel_run_lifecycle_test.go`
2. Lock the delivery boundary before implementation starts:
   - lifecycle-originated progress events ship now
   - future Codex-native stream ingestion is enabled by design, not implemented end-to-end in this slice
   - `namba status` remains out of scope
3. Define the event contract for parallel progress:
   - append-only JSONL artifact path
   - required event fields
   - source and scope vocabulary
   - stable minimum phase vocabulary for watcher consumers
   - stable `run_id`, ordering, and detail-summary semantics for later plain-text watch surfaces
4. Introduce a reusable progress publication boundary:
   - narrow event sink or observer interface
   - concrete file-backed writer for `.namba/logs/runs/`
   - explicit publication path from `parallel_lifecycle.go` into `executeRun`
   - deterministic serialization or monotonic sequence assignment for concurrent workers
   - explicit behavior for initialization, append, and final flush failures
5. Wire lifecycle-originated progress publication from `parallel_lifecycle.go`:
   - preflight result
   - worker staging / queue readiness
   - worker start and completion
   - merge pending / merge start / merge completion or failure
   - cleanup / preserved outcome
   - overall run completion
6. Extend `execution.go` to publish execution-owned progress through the same boundary:
   - execution start
   - validation start and completion
   - repair attempt boundaries when applicable
   - metadata/detail emission without expanding the baseline watcher phase vocabulary prematurely
7. Preserve the existing aggregate report model:
   - keep `spec-xxx-parallel.json`
   - keep request/execution/validation artifacts
   - only add fields or references when they help connect the final report to the event log without changing the report's role
8. Add regression coverage for:
   - ordered append-only event persistence
   - success and failure paths
   - sink failure policy behavior
   - sequence or publication ordering guarantees under concurrent workers
   - merge-blocked and preserved-worker paths
   - compatibility of lifecycle events with future non-lifecycle sources
9. Run the relevant review passes under `.namba/specs/SPEC-028/reviews/` and refresh the readiness summary when the design stabilizes.
10. Run validation commands.
11. Sync artifacts with `namba sync` or an equivalent source-aligned invocation when implementation changes make managed outputs stale.
