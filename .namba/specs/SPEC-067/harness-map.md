# Harness Map

`harness-request.json` classifies this work as touching Namba core. This SPEC maps that core-harness concern to behavior-preserving refactor evidence.

## Core Surfaces

- CLI command routing and help.
- SPEC creation and plan review handoff.
- Runtime execution and evidence.
- Queue conveyor behavior.
- Report and eval artifacts.
- Guardrails and Codex hook boundaries.
- PR, land, release, update, and regen workflows.

## Required Evidence

- Phase 0 baseline checklist.
- Characterization tests for public command behavior.
- Golden/fixture assertions for command output, generated layout, JSON schema, and evidence schema.
- Fake-adapter tests for external integrations.
- Pre-refactor and post-refactor validation command logs.
- `wc -l internal/namba/namba.go` before and after.
- `go list ./...` after refactor.

## Review Mapping

- Product review: confirms this is behavior-preserving refactor scope, not feature work.
- Engineering review: confirms phase ordering, package-boundary caution, rollback, and validation depth.
- Design review: marks frontend/visual design as not applicable and checks CLI-facing semantics remain unchanged.
