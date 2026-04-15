# Engineering Review

- Status: clear
- Last Reviewed: 2026-04-15
- Reviewer: namba-planner
- Command Skill: `$namba-plan-eng-review`
- Recommended Role: `namba-planner`

## Focus

- Lock architecture, sequencing, failure modes, trust boundaries, and validation strategy before execution starts.

## Findings

- The previously raised engineering gaps are now addressed at the planning-contract level. `spec.md`, `plan.md`, and `acceptance.md` all treat sink failure handling, deterministic ordering, and the `parallel_lifecycle.go` -> `executeRun` publication seam as first-class deliverables rather than implicit implementation details.
- Sink failure policy is now explicit enough for implementation planning even though the exact behavior matrix is intentionally left to the implementation slice. The core artifacts now require defined behavior for initialization failure, append failure during active execution, and final flush/close failure, and they require tests for those paths. That removes the earlier ambiguity about whether sink durability is optional; it is not optional, and any degraded or fatal behavior now has to be chosen deliberately and exercised in tests.
- Ordering and correlation are now concrete enough to prevent a concurrency-shaped redesign later. The contract requires a stable `run_id` plus either a monotonic `sequence` or an equivalent serialized publication guarantee, and the plan calls out deterministic serialization for concurrent workers. That is the right engineering boundary for watcher trust without overcommitting to one internal implementation shape too early.
- The publication seam into `executeRun` is now explicit. The planning artifacts call for a reusable sink/observer boundary, name the `parallel_lifecycle.go` -> `executeRun` handoff directly, and require `execution.go` to publish execution, validation, and repair-attempt transitions through the same contract. That is sufficient to start implementation without hardcoding JSONL writes across the runtime.
- One implementation note remains: the docs now require the right decisions, but they do not yet pick the exact sink API or failure-mode matrix. That is acceptable at this stage because the gap has moved from "missing design constraint" to "first code change must lock the concrete rule." I do not consider that a planning blocker anymore.

## Decisions

- Keep v1 scoped to lifecycle-owned and execution-owned events only. Do not rewrite the `CombinedOutput()`-based Codex runner in this SPEC.
- Keep the final `spec-xxx-parallel.json` artifact as the terminal summary and treat the JSONL event stream as the append-only source for future watchers.
- Require the v1 contract to be source-aware and additive so later Codex-native streaming can publish into the same schema without changing the watcher-facing minimum phases.
- Treat the three earlier review concerns as resolved for planning purposes: failure semantics are now mandatory design input, ordering guarantees are now contractually explicit, and the `executeRun` publication boundary is now part of scope rather than an implementation guess.

## Follow-ups

- Make the first implementation step a concrete sink behavior matrix covering initialization failure, append failure, and close failure so code review can judge the runtime trust model against one explicit rule set.
- Choose and document one publication mechanism early: either monotonic `sequence` assignment, a serialized single-writer path, or an equivalent design that preserves deterministic replay for watchers.
- Keep repair-attempt details in metadata or summary/detail fields unless a new watcher-facing phase is truly required; the baseline phase vocabulary should stay compatibility-sensitive.
- Add tests for preflight failure, append failure, repair-attempt ordering, merge-blocked preservation, and cleanup completion as part of the first implementation pass, not as follow-on hardening.

## Recommendation

- Advisory recommendation: engineering is now clear enough to proceed. The remaining work is to lock the concrete sink API and failure-mode matrix at implementation start, not to reopen the SPEC design.
