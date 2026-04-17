# Acceptance

- [ ] `namba regen` no longer deletes non-regen managed outputs such as synced README/docs, `.namba/project/*`, or SPEC review readiness files.
- [ ] `namba regen` still removes stale managed artifacts that belong to the regen-generated surface.
- [ ] The fix stays scoped to regen cleanup behavior and does not change unrelated `sync` or `project` responsibilities.
- [ ] Validation commands pass
- [ ] Existing behavior around the affected area is preserved
- [ ] A regression test covering the fix is present
