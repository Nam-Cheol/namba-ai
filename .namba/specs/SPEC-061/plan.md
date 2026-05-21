# SPEC-061 Plan

1. Refresh project context with `namba project` if project docs drift before implementation.
2. Lock regression tests before changing behavior.
   - Add queue cleanliness tests that prove queue-owned active-run artifacts do not block branch readiness.
   - Add negative tests proving unrelated user edits and mixed dirty states still block.
   - Add execution transport tests that prove `req.Prompt` is not present in argv for normal exec or resume, including Windows command-length simulation.
   - Add queue fallback tests for remote PR handoff unavailability and local merge conflict blocking.
3. Implement queue-scoped cleanliness.
   - Refactor away from a broad global queue-log ignore in shared clean checks.
   - Introduce a queue-only classifier or checker keyed to the active queue state/SPEC.
   - Preserve strict dirty checks for non-queue commands.
4. Implement non-argv Codex prompt transport.
   - Keep capability flag/config resolution intact.
   - Pass `-` as the prompt argument and send the prompt body through stdin where supported.
   - Use a temp request file fallback only if a runner path cannot supply stdin.
   - Preserve request JSON/markdown evidence artifacts.
5. Implement explicit local fallback for remote PR handoff unavailability.
   - Keep remote PR handoff as the preferred path.
   - Detect only remote-unavailable handoff failures after sync and validation success.
   - Commit the active branch if needed, merge it into the configured local base branch, record local fallback evidence, mark the SPEC landed, and continue.
   - Block with recovery guidance on validation failure, checks failure, negative mergeability, review blocker, pause/stop, or local merge conflict.
6. Update operator-facing queue messages and report output.
   - Distinguish queue-owned writes, user edits, mixed dirty state, fallback-used, blocked, resumed, landed, and queue-complete outcomes.
7. Run targeted tests, then full validation.
   - Targeted: queue command tests, execution/capability tests, and e2e workflow fixture tests.
   - Full: `go test ./...`, `gofmt -l "cmd" "internal" "namba_test.go"`, `go vet ./...`.
8. Sync artifacts with `namba sync`.
