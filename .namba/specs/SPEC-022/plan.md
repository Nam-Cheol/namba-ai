# SPEC-022 Plan

1. Inspect the current planning-context path across `namba project`, `.namba/project/*`, and `$namba-plan` so the redesign is anchored to the existing generator and skill boundaries rather than bolted on top of them.
2. Define the foundation-release contract:
   - default include/exclude scoping
   - repository/app-boundary detection
   - source-priority rules
   - evidence/confidence schema
   - mismatch-report contract
   - thin-output quality-gate rules
3. Define output compatibility and document hierarchy so current `.namba/project/*` entry points remain stable while richer artifacts are added.
4. Introduce a configurable analyzer core that classifies files, groups repository systems, attaches evidence, and evaluates output quality independently from stack-specific parsing.
5. Split framework-specific reasoning into adapter layers, but keep v1 implementation to a Go-first path plus a generic extension seam; document broader adapter rollout as follow-up work.
6. Redesign `.namba/project/*` output so the generated docs cover purpose/user flow, runtime/deploy topology, modules, interfaces, data/state, security, test map, mismatch report, and appendix-style structure output with a fixed reading order.
7. Update `$namba-plan` and `$namba-project` guidance so planning consumes evidence-backed artifacts, preserves conflicts, and respects multi-app boundaries instead of flattening the repository.
8. Run the relevant review passes under `.namba/specs/SPEC-022/reviews/` and refresh the readiness summary before implementation if the final design materially affects architecture or user-facing workflow guidance.
9. Add regression tests for scoping, semantic extraction, adapter routing, mismatch reporting, output compatibility, and thin-output quality gates.
10. Run validation commands and sync any generated artifacts that the implementation changes require.
