# SPEC-032 Plan

1. Refresh project context with `namba project` and inventory the current harness-related command, skill, runtime, and review surfaces.
2. Record the current structural versus prose-only state in `baseline.md`, citing the concrete repo files that own each behavior today.
3. Define the canonical request classes, minimal typed metadata, and routing precedence in `contract.md`.
4. Lock the transport and persistence rule:
   - `.namba/specs/<SPEC>/harness-request.json` for SPEC routes
   - transient JSON through `namba __create preview|apply` for direct routes
5. Define the standard evidence pack for harness-classified work, including the exact condition that makes `harness-map.md` mandatory.
6. Rewrite the role boundaries for `namba plan`, `namba harness`, and `$namba-create` around the typed decision contract rather than around loose wording, and add operator-facing route examples.
7. Define the validator and eval fixture matrix in `eval-plan.md`, including should-route, should-not-route, direct-route escalation, evidence-completeness checks, and legacy-readiness compatibility.
8. Ensure the proposed model layers onto the existing planning-start, review-readiness, `namba sync`, and runtime contracts without inventing a second planning package type or reopening `namba run` mode semantics.
9. Run the relevant review passes under `.namba/specs/SPEC-032/reviews/` and refresh the readiness summary when the contract stabilizes.
10. Implement the minimal shared classifier, metadata surfacing, readiness-validator hook, and evidence requirements.
11. Add regression coverage for route selection, evidence completeness, transport split, legacy readiness refresh, and role-boundary preservation.
12. Run validation commands.
13. Sync artifacts with `namba sync`.
