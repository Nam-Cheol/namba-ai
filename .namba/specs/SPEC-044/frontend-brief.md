# Frontend Brief

Task Classification: frontend-minor
Classification Rationale: Harness-only workflow planning that mentions frontend gates but does not implement end-user UI in this SPEC package.
Frontend Gate Status: not-applicable
Problem Gate: not-applicable
Reference Gate: not-applicable
Critique Gate: not-applicable
Decision Gate: not-applicable
Prototype Gate: not-applicable
Prototype Evidence: n/a

## Current Pattern

- `SPEC-036` blocks `frontend-major` work on missing, insufficient, invalid, or mismatched synthesis evidence.
- The existing gate does not yet require a negative-first Do-Not Design Contract or post-implementation banned-pattern checks.

## Intended Change

- Extend the workflow contract for future `frontend-major` UI work while keeping this SPEC itself in the harness/core implementation path.
- Implementation details live in `spec.md`, `plan.md`, and `acceptance.md`.

## Notes

- The current classifier scaffolded this sidecar because the request contains explicit frontend terms; the persisted classification clarifies that this SPEC is workflow/harness work, not a UI implementation task.
