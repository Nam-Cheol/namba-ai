# SPEC-030 Plan

1. Reproduce and confirm the cleanup scope bug:
   - inspect `runRegen`, `replaceManagedOutputs`, and manifest ownership matching
   - confirm which managed surfaces should be cleanup-eligible for regen and which should not
2. Narrow the regen cleanup decision:
   - keep path-based matching for files that `regen` writes directly
   - add an owner-aware matcher that is scoped only to the regen-managed namespace instead of every `namba-managed` manifest entry
3. Preserve existing cleanup behavior for stale regen outputs:
   - stale repo skills
   - stale Codex agents
   - legacy `.codex/skills/*` mirrors
   - `.namba/codex/*` outputs and repo Codex config
4. Add regression coverage for both sides of the contract:
   - stale regen-managed artifacts are removed
   - synced README/docs, `.namba/project/*`, and SPEC readiness outputs survive `regen`
5. Run validation commands and targeted regression checks
6. Sync artifacts with `namba sync`
