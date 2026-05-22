# Acceptance

- [ ] Generator/source changes in `internal/namba/codex.go`, `internal/namba/templates.go`, `internal/namba/create_engine.go`, `internal/namba/agent_runtime.go`, and related `internal/namba/*_test.go` files explain every generated output diff.
- [ ] A fresh temporary repo created with `namba init` automatically generates improved `.agents/skills/*`, `.codex/agents/*.md`, and `.codex/agents/*.toml` surfaces.
- [ ] An existing repo regenerated with `namba regen` reproduces the same quality contract idempotently.
- [ ] Current repo managed generated files are updated only via `namba regen` after generator/source changes and are treated as secondary evidence.
- [ ] Snapshot, fixture, or golden tests prove generator-first behavior rather than current-repo patching.
- [ ] Deterministic tests cover role purpose, read-only versus mutating boundaries, required outputs, pass/fail criteria, evidence expectations, security responsibilities, fallback implementer boundaries, secrets prohibitions, destructive command/escalation policy, and non-project-specific guidance.
- [ ] Test coverage includes Go unit/golden tests for template drift.
- [ ] Test coverage includes CLI-style temporary-repo integration tests for `namba init` and `namba regen`.
- [ ] `.namba/manifest.json` managed ownership remains correct for generated skill and custom-agent outputs.
- [ ] The implementation does not introduce downstream repo backfills, network-dependent checks, LLM-judge evaluations, or unrelated workflow restructuring.
- [ ] Generated guidance remains non-project-specific and suitable for future repositories.
- [ ] `gofmt` is run on Go changes.
- [ ] `go test ./...` passes, or any failure is reported with exact cause and impact.
- [ ] `go vet ./...` passes, or any failure is reported with exact cause and impact.
