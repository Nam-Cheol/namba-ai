# SPEC-055 Plan

1. Refresh project context with `namba project`
2. Inspect `.github/workflows/release.yml` for `windows-latest` and release action versions that still rely on Node.js 20.
3. Run the relevant review passes under `.namba/specs/SPEC-055/reviews/` and refresh the readiness summary.
4. Implement the smallest safe workflow-only fix:
   - Pin the Windows runner to `windows-2022`.
   - Upgrade or replace only the release workflow action versions required to remove Node.js 20 deprecation warnings.
5. Run YAML/action static validation for `.github/workflows/release.yml`.
6. Confirm the workflow through `workflow_dispatch` or a PR check path that does not publish a real GitHub Release; do not use tag-push validation.
7. Sync artifacts with `namba sync`.
