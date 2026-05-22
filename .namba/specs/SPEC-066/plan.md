# SPEC-066 Plan

1. Refresh project context with `namba project`.
2. Inspect generator and template ownership for `.agents/skills/*`, `.codex/agents/*.md`, `.codex/agents/*.toml`, and `.namba/manifest.json` in `internal/namba/codex.go`, `internal/namba/templates.go`, `internal/namba/create_engine.go`, `internal/namba/agent_runtime.go`, and related `internal/namba/*_test.go` files.
3. Identify the shared generated contract fields needed across role purpose, read-only versus mutating boundaries, required outputs, pass/fail criteria, evidence expectations, security responsibilities, fallback implementer boundaries, destructive command/escalation policy, and non-project-specific guidance.
4. Update generator/template/render source so command-entry skills and project-scoped agents receive the improved contract first, then ensure the common contract reaches the remaining managed generated instruction surfaces.
5. Add deterministic Go unit/golden tests for generated template drift and boundary coverage.
6. Add CLI-style temporary-repo integration tests proving fresh `namba init` and existing-repo `namba regen` produce the same quality contract idempotently.
7. Run `namba regen` in the current repo and verify generated diffs are explained by generator/source changes.
8. Run `gofmt`, `go test ./...`, and `go vet ./...`, reporting exact cause and impact for any failure.
9. Sync artifacts with `namba sync`.
