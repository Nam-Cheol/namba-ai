# SPEC-066 Harness Map

## Artifact Targets

- `skill`: `.agents/skills/*` receives the generated instruction contract through shared skill rendering.
- `agent`: `.codex/agents/*.md` and `.codex/agents/*.toml` receive the generated instruction contract through shared role-card and custom-agent rendering.
- `workflow`: `namba init`, `namba regen`, and `$namba-create` paths exercise the same source-owned contract.
- `docs`: `.namba/manifest.json`, project summaries, and readiness artifacts record the regenerated surfaces and validation evidence.

## Source Mapping

- `internal/namba/templates.go`: common contract helper plus managed skill, role-card, and custom-agent renderers.
- `internal/namba/create_engine.go`: user-created custom-agent instruction generation.
- `internal/namba/templates_test.go`: renderer and template drift coverage.
- `internal/namba/update_command_test.go`: regen integration coverage.
- `namba_test.go`: fresh init scaffold coverage.

## Evidence Mapping

- Current repository generated diffs prove local-source `regen` applied the source-owned template changes.
- Temporary repository `init` output proves future repositories receive the improved contract.
- Temporary repository `regen` with no instruction-surface diff proves existing repositories reproduce the same contract.
