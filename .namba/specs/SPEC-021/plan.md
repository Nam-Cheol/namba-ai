# SPEC-021 Plan

1. Refresh project context with `namba project` and audit the current gap between renderer source and checked-in README/workflow docs.
2. Lock the implementation boundary before rewriting copy: keep `internal/namba/readme.go` as source of truth, keep root README and workflow guide role-separated, and treat generated outputs as `namba sync` artifacts only.
3. Redesign the README information architecture for first-time users so the basic journey and command-choice matrix are obvious at a glance.
4. Rewrite the command and skill explanation sections to explain when to use each surface, especially `namba project`, `namba plan`, `namba harness`, `namba fix`, `namba run`, `namba sync`, `namba pr`, and `namba land`.
5. Update workflow guide structure so lifecycle commands, planning commands, execution modes, review readiness, and merge flow are clearly separated without duplicating the README.
6. Sync localized generated docs from the renderer rather than patching checked-in README files by hand.
7. Add or tighten renderer/doc regression tests for the new information architecture, command/skill guidance, and stale-doc drift risk.
8. Run the relevant review passes under `.namba/specs/SPEC-021/reviews/` and refresh the readiness summary.
9. Run validation commands.
10. Sync artifacts with `namba sync`.
