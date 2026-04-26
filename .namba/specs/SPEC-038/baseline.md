# SPEC-038 Baseline

## Current Skill Surfaces

Existing generated command-entry and guidance skills include `$namba`, `$namba-help`, `$namba-create`, `$namba-plan`, `$namba-plan-review`, `$namba-harness`, `$namba-fix`, `$namba-run`, `$namba-sync`, `$namba-pr`, `$namba-land`, `$namba-regen`, and `$namba-update`.

Evidence:

- `internal/namba/codex.go` defines `managedCodexSkillNames()` and `codexSkillTemplates()`.
- `internal/namba/templates.go` renders `AGENTS.md`, `$namba`, `$namba-help`, `$namba-create`, `$namba-plan`, `$namba-harness`, `$namba-run`, and Codex usage docs.
- `internal/namba/readme.go` renders README and workflow guide command-skill sections.
- `.namba/codex/README.md`, `README.md`, and `docs/workflow-guide.md` currently describe `$namba-help` as read-only guidance and `namba harness` as reusable skill, agent, workflow, or orchestration planning.

## Current Gap

No generated managed skill named `namba-coach` exists. Users who have vague intent or choose the wrong Namba command must rely on `$namba` routing or `$namba-help` usage guidance, which blurs the distinction between documentation help and goal-specific workflow coaching.

## Relevant Tests

Existing tests already cover adjacent contracts:

- `internal/namba/templates_test.go` covers generated skill and agent template contracts.
- `internal/namba/readme_contract_test.go` and `internal/namba/readme_sync_test.go` cover README and workflow guide generated text.
- `namba_test.go` and scaffold tests cover init-generated repo assets.
- `internal/namba/help_contract_test.go` covers public CLI command help; it should not gain a `namba coach` expectation in this slice.

## Constraints

- Generated managed skill outputs must survive `namba regen`.
- Obsolete generated skill cleanup must not remove `namba-coach` after it is registered.
- No public CLI command should be introduced.
- Existing skill responsibilities must remain stable.
