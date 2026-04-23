# SPEC-035 Plan

1. Refresh project context with `namba project`
2. Extract a shared Codex access settings layer so `namba init` and the new reconfiguration command use the same validation, labeling, and normalization rules
3. Add the top-level `namba codex access` subcommand contract with deterministic zero-argument inspect behavior, explicit flag-driven mutation behavior, project-root safety checks, and bounded managed-output regeneration
4. Upgrade the `namba init` wizard and help text so bootstrap users get the same preset language, consequence statements, and preview model as the post-init command
5. Update generated docs and user-facing guidance to describe both initial setup and later access reconfiguration from the repo root, including `init --help`, `codex access --help`, and getting-started guidance
6. Add regression tests for bootstrap flags, inspect-only reads, existing-repo access edits, generated config regeneration, no-op/session-refresh signaling, invalid-config rejection, and the no-clobber guarantee for non-managed files
7. Run validation commands
8. Sync artifacts with `namba sync`
