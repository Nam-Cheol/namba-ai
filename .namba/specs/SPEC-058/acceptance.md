# Acceptance

- [ ] At least three E2E scenarios exist: happy path, blocked unsafe path, and missing evidence or validation failure path.
- [ ] Each scenario creates and uses an independent temp repo.
- [ ] Tests exercise production CLI command handlers through `App.Run` rather than bypassing command dispatch.
- [ ] External Codex, GitHub, network, token, PR, and release interactions are replaced by deterministic fakes.
- [ ] The fake runner fails with a clear diagnostic when an unexpected external command is attempted.
- [ ] The happy path verifies generated project docs, SPEC files, run or dry-run evidence, queue state, status output, PR simulation, and land or release readiness simulation.
- [ ] The blocked unsafe path verifies that unsafe state is stopped with actionable blocker or recovery detail.
- [ ] The missing evidence or validation failure path verifies structured failure evidence and does not silently pass.
- [ ] Assertions use JSON or focused field checks for durable artifacts and avoid brittle full-output golden comparisons.
- [ ] Tests are cross-platform and avoid shell-specific assumptions except through existing shell helper seams.
- [ ] Existing unit tests are not weakened or skipped.
- [ ] `go test ./...` passes after implementation.
