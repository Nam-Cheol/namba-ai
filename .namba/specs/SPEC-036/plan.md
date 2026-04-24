# SPEC-036 Plan

1. Refresh project context with `namba project` and map the current planning, review, run, and doc-generation surfaces that assume advisory-only readiness.
2. Define the `frontend-major` / `frontend-minor` classification model plus the five-gate evidence contract, and make `frontend-brief.md` the canonical parser-visible artifact for classification, classification rationale, and per-gate status.
3. Update planning scaffold generation so new frontend-touching planning seeds `frontend-brief.md`: `frontend-major` gets the full fixed-label gate block plus evidence sections, while `frontend-minor` gets explicit classification/rationale plus lightweight `not-applicable` gate defaults, without breaking historical or non-frontend SPEC creation.
4. Update review rendering and readiness aggregation so frontend gate status is tracked separately from the existing product/engineering/design advisory summary, including clear `missing` versus `insufficient` reporting, invalid-contract surfacing, and cross-artifact mismatch visibility.
5. Add selective execution gating to `namba run SPEC-XXX` so explicitly classified frontend-major work blocks with a clear "blocked for frontend synthesis" result when required evidence is missing, insufficient, or contradictory, while malformed `frontend-brief.md` contracts return deterministic remediation and non-frontend / valid frontend-minor paths stay advisory.
6. Update role cards, command skills, README/workflow guidance, and synced docs so research-before-implementation and revised frontend role ownership are stated consistently.
7. Add regression coverage for classification persistence, scaffold generation, design-review schema changes, readiness summary changes, invalid fixed-label parsing, cross-artifact contradiction handling, selective run blocking, blocked-run remediation messaging, parser defaults for historical SPECs, and mixed-scope planning guidance.
8. Run validation commands.
9. Sync artifacts with `namba sync`.
