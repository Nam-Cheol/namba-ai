# SPEC-028

## Goal

Add an event-driven progress model for parallel execution so Namba can persist and expose worker lifecycle state in real time today, while leaving a clean seam for future Codex background-agent or app-server streaming sources to emit the same event contract.

## Context

- Project: namba-ai
- Project type: existing
- Language: go
- Mode: tdd
- Work type: plan
- Planning surface: `namba plan "<description>"`
- Command choice rationale: this is a feature slice on Namba's runtime/observability surface, not a reusable harness-only scaffold or a direct repo-local skill/agent generator.
- Verified local context as of 2026-04-15:
  - `internal/namba/parallel_lifecycle.go` stages workers, launches goroutines, records worker outcomes only after execution returns, and writes the final `spec-xxx-parallel.json` report at the end of the run.
  - `internal/namba/parallel_run.go` defines the persisted parallel summary report, but that report is terminal-state oriented and does not represent mid-run progress.
  - `internal/namba/execution.go` already owns execution turns, validation attempts, and repair loops, which makes it the right place to emit execution/validation progress transitions once an observer boundary exists.
  - `internal/namba/namba.go` still runs `codex` through `cmd.CombinedOutput()`, so Namba cannot yet consume mid-command Codex progress directly from the installed CLI path.
  - `.namba/logs/runs/` already stores request, execution, validation, and preflight artifacts, so adding an append-only progress artifact extends an existing logging surface instead of inventing a second state model.
  - `namba status` is intentionally a static repository summary today and is not the right integration target for live run-state reporting.

## Target Reader

- Namba maintainers who need a practical live-progress foundation without immediately rewriting the Codex runner architecture.
- Reviewers who need a slice that improves observability now but does not trap the repository in a lifecycle-only event model.
- Future implementers of `namba run --watch`, `namba watch`, or Codex-stream-backed progress surfaces who need a durable event contract to build on.

## Problem

Parallel execution already has real lifecycle stages, but Namba currently persists only the end result:

1. Missing live visibility
   - Worker setup, execution, validation, merge, and cleanup all happen in distinct phases.
   - Users and maintainers cannot inspect those transitions while the run is active; they only see the final summary after completion.
2. Progress and summary are conflated
   - `spec-xxx-parallel.json` is a terminal report, not an append-only event stream.
   - That makes it useful for postmortems, but not for watchers, dashboards, or lightweight run-following commands.
3. Future Codex progress has nowhere clean to land
   - Codex `0.120.0` can stream background-agent progress while work is still running.
   - Namba's current subprocess model cannot consume that yet, and even if it could, there is no source-agnostic event contract for lifecycle events vs runner-native stream events to share.
4. UI work would be premature without a state contract
   - A `watch` command or live TUI can be added later, but it should read from a stable event source rather than custom ad hoc state stitched out of the final report.

## Desired Outcome

- Parallel runs produce an append-only progress artifact under `.namba/logs/runs/` that records lifecycle transitions as they happen.
- v1 is explicitly a machine-consumable and tail-friendly progress source, not a new operator-facing live UI on its own.
- The event model is explicitly source-aware:
  - lifecycle-originated events ship in this slice
  - future Codex-native streaming events can emit the same contract with a different source value
- Worker progress is visible at minimum through the following stages:
  - `queued`
  - `running`
  - `validating`
  - `merge_pending`
  - `merging`
  - `done`
  - `failed`
- The final parallel summary report remains in place and continues to be the terminal aggregate artifact.
- The first slice improves logging and architecture, not UI:
  - no requirement to ship a new watch command in this SPEC
  - no requirement to modify Codex's own TUI
- The resulting design gives a later slice one obvious extension point for Codex background-agent or app-server event ingestion.

## V1 Success Definition

- A parallel run writes a durable append-only event log as the run progresses.
- Event records distinguish run-level and worker-level transitions.
- Event records remain trustworthy under concurrency because publication ordering is deterministic.
- The event contract is stable enough that a future consumer can follow progress without reading partially-mutated aggregate JSON.
- The execution path publishes lifecycle progress through a reusable boundary instead of hardcoding file writes directly into every callsite.
- The event contract carries enough stable detail for future plain-text watch surfaces to summarize failures, preserved workers, and merge-blocked states without a schema break.
- Existing final artifacts and current execution semantics still work after the change.
- Tests prove both the current lifecycle-originated event flow and the extension seam for future non-lifecycle event sources.

## Scope

- Define a parallel-progress event model for checked-in Go code, including:
  - event timestamp
  - spec/run identity
  - stable run identifier and monotonic event sequence
  - event source
  - scope (`run` or `worker`)
  - worker identity when applicable
  - current phase
  - status or outcome detail
  - stable human-facing summary/detail field or metadata convention for future watch consumers
  - optional metadata payload for future extension
- Add an append-only writer for parallel progress events under `.namba/logs/runs/`.
- Emit progress events from the existing lifecycle boundaries, at minimum for:
  - preflight completion or failure
  - worker staging / queue readiness
  - worker execution start
  - worker validation start
  - worker execution+validation completion
  - merge readiness
  - per-worker merge start/completion/failure
  - cleanup or preserved-on-failure outcome
  - overall run completion
- Introduce a reusable progress-observer boundary so `parallel_lifecycle.go` and `execution.go` can publish events without binding the design to one concrete file sink.
- Define sink failure handling for:
  - initialization failure before workers start
  - append failure while the run is active
  - final flush or close failure after worker completion
- Keep the final `spec-xxx-parallel.json` summary report and existing request/execution/validation artifacts intact.
- Add regression coverage for event ordering, failure-path durability, and source/phase semantics.

## Event Contract

- Primary artifact:
  - `.namba/logs/runs/<spec-id>-parallel.events.jsonl`
- Contract shape:
  - one JSON object per line
  - append-only writes in chronological order
  - each record stands alone and can be consumed without reading later lines
- Required dimensions:
  - `spec_id`
  - `run_id` or equivalent stable parallel-run identity
  - monotonic `sequence` or an equivalent deterministic single-writer ordering guarantee
  - `timestamp`
  - `source`
  - `scope`
  - `phase`
  - `status`
- Human-readable surface:
  - include a concise summary/detail field or metadata convention that later plain-text watch/readout surfaces can use without exposing raw internal phrasing directly
  - preserve a small set of stable, low-ambiguity fields so later operator surfaces do not depend on color, icons, or hidden aggregate state
- Recommended source values:
  - `lifecycle` for this slice
  - reserve space for future sources such as `codex-background-agent` or `codex-app-server`
- Recommended scopes:
  - `run`
  - `worker`
- Phase vocabulary:
  - the contract must support the required user-visible states above
  - the baseline watcher-facing phases are compatibility-sensitive and must stay stable; additional phases such as `prepared`, `cleanup`, or `preserved` are allowed only as additive extensions
  - `done` must be interpreted together with `scope`, `status`, and any stable summary/detail fields; consumers must not rely on hidden phase combinations to infer whether a worker finished execution, a worker finished merge, or the overall run completed
- The event schema must be designed so a future Codex stream adapter can publish finer-grained progress without forcing a breaking rewrite of the lifecycle-originated events.

## Design Approach

### 1. Source-Agnostic Progress Sink

- Introduce a narrow interface or helper boundary for progress publication.
- Use an explicit boundary such as `ProgressPublisher` or `ProgressSink` that is passed from the parallel lifecycle into `executeRun`, so execution turns, validation, and repair attempts can publish under the same contract.
- The first implementation can write to JSONL on disk, but the execution and lifecycle code should depend on the sink abstraction rather than the file format directly.
- This keeps the first slice practical while allowing a later Codex-stream consumer to reuse the same sink or fan out to multiple sinks.
- The publication path must be serialized or otherwise ordering-safe so concurrent worker goroutines do not race into nondeterministic JSONL ordering.

### 2. Lifecycle Events First

- This slice should publish the progress Namba already knows today from its own orchestration boundaries.
- That means the implementation does not wait on Codex protocol integration before improving observability.
- `parallel_lifecycle.go` owns worker staging, merge, cleanup, and preservation transitions.
- `execution.go` owns execution, validation, and repair-related transitions.
- Repair attempts should remain watcher-compatible by using stable phases plus metadata, rather than multiplying the public phase vocabulary prematurely.

### 3. Final Summary Stays Separate

- `spec-xxx-parallel.json` remains the aggregate summary artifact.
- The event log supplements that summary; it does not replace it in v1.
- The final report may optionally include a pointer to the event log path, but watcher logic should treat the JSONL log as the source of progress truth.

### 4. Extensibility For Codex Streaming

- This slice does not rewrite `runBinary` or the Codex subprocess model.
- Instead, it creates the data contract and observer seam that a later slice can plug Codex-native stream events into.
- The future Codex integration should be able to emit:
  - the same `run` or `worker` scoped events when appropriate
  - richer metadata without breaking lifecycle-only consumers

### 5. Failure Policy Must Be Explicit

- Sink initialization failure before worker execution must be classified explicitly as either fatal or a documented degraded mode before implementation starts.
- Append failure during active execution must have a defined response that preserves trust in the remaining artifacts and does not silently continue as if full progress durability still exists.
- Final flush or close failure must likewise have a defined response, including whether the run is marked degraded, failed, or preserved for inspection.
- The failure policy should align with existing Namba behavior for runtime artifact durability instead of creating a special silent-failure exception for progress logs.

## Implementation Priority

1. Define the event schema and durable append-only writer.
2. Lock the failure policy, ordering guarantee, and publication boundary before code movement starts.
3. Wire lifecycle-originated progress publication through `parallel_lifecycle.go`.
4. Extend `execution.go` so execution and validation stages can publish through the same boundary.
5. Preserve and, where useful, enrich the final summary report without changing its role.
6. Add tests proving success paths, failure paths, ordering guarantees, and event-seam extensibility.
7. Defer UI/watch surfaces and Codex-native stream ingestion to follow-up slices.

## Initial Delivery Boundary

- First slice deliverables:
  - a checked-in event model and writer
  - an explicit failure policy for sink initialization, append, and close semantics
  - a deterministic ordering strategy with stable `run_id` and sequence semantics
  - an explicit `parallel_lifecycle.go` -> `executeRun` publication seam
  - lifecycle and execution hooks that emit progress records for parallel runs
  - durable event log output under `.namba/logs/runs/`
  - regression tests for event ordering and failure semantics
  - no new interactive UI requirement
- Follow-up slices may add:
  - `namba run --watch`
  - `namba watch`
  - summary rendering over the event log
  - Codex app-server or background-agent event ingestion

## Non-Goals

- Do not redesign `namba status` into a live-run dashboard in this SPEC.
- Do not require a new watch-mode CLI surface in this slice.
- Do not rewrite the Codex runner from `CombinedOutput()` to a streaming transport in this slice.
- Do not remove or replace the existing final parallel summary report.
- Do not expand this work into team-mode or solo-mode live progress unless the shared event boundary makes that adoption trivial and non-disruptive.

## Design Constraints

- Keep `.namba/logs/runs/` as the repository-local runtime log surface.
- Prefer append-only event persistence over rewriting mutable shared state.
- Preserve current execution semantics, cleanup rules, and failure-preservation behavior.
- Keep the event schema source-aware so lifecycle events and future Codex-native events can coexist.
- Treat the event boundary as the primary extensibility seam; do not couple future UI work directly to internal goroutine structure.
- Keep the slice reviewable and bounded: progress logging first, UI and Codex streaming later.
