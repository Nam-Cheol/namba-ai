# SPEC-033 Plan

1. Confirm the current split between planning-side harness evidence and runtime-side structured artifacts:
   - inspect `internal/namba/harness_contract.go`
   - inspect `internal/namba/spec_review.go`
   - inspect `internal/namba/execution.go`
   - inspect `internal/namba/parallel_progress.go`
2. Design the typed execution-evidence manifest:
   - choose the persisted artifact path under `.namba/logs/runs/`
   - define identity, status, and provenance fields
   - define required base artifact references
   - define optional extension slots for browser and selected runtime evidence
   - define one canonical manifest emission/finalization rule across success and failure paths
3. Keep the phase model explicit:
   - preserve current pre-implementation readiness semantics
   - define how post-run execution proof is surfaced without pretending it is the same signal
   - keep `reviews/readiness.md` plan-only and make `namba sync` the primary v1 execution-proof consumer
4. Implement manifest emission from the runtime paths that already write structured artifacts.
5. Wire parallel progress references into the manifest without redefining the `SPEC-028` event schema.
6. Define browser-evidence attachment rules:
   - optional
   - capability-scoped
   - `not_applicable` when browser verification is irrelevant
7. Update readiness/sync summary logic so the latest execution-evidence layer can be surfaced separately from plan readiness.
8. Update stable documentation and generated guidance where the new contract becomes user-visible.
9. Add regression coverage for:
   - base manifest generation
   - preflight-failure manifest coverage
   - execution failure before validation
   - validation failure after retries
   - optional-extension behavior
   - browser `not_applicable` vs present cases
   - parallel progress linkage
   - legacy compatibility when the manifest is absent
10. Run validation commands.
11. Refresh the relevant review artifacts and readiness summary.
