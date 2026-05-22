# Acceptance

- [x] Queue-owned active-run metadata and evidence writes do not trip queue branch-clean checks.
- [x] Unrelated user edits, modified source/config/SPEC files, and mixed user edits plus queue-owned files still block when queue requires a clean branch.
- [x] Non-queue clean checks for `plan`, `pr`, `land`, and `release` are not weakened by the queue-owned metadata exception.
- [x] `codex exec` and `codex exec resume` no longer receive the full execution prompt as an argv element.
- [x] Long execution prompts pass under a Windows command-line length simulation by using stdin or a temp request file.
- [x] Existing Codex capability resolution for approval policy, sandbox, model, profile, web search, add-dir, JSON output, and resume mode is preserved.
- [x] When remote PR handoff is unavailable after implementation, sync, and validation success, queue records an explicit local fallback, locally merges the active branch into the configured base branch, marks the SPEC landed, and continues to the next selected SPEC.
- [x] Queue does not use local fallback for validation failures, failed checks, negative mergeability, review blockers, pause/stop requests, or local merge conflicts.
- [x] Queue output/reporting makes fallback-used, blocked, resumed, landed, and queue-complete states clear to the operator.
- [x] Queue regression coverage includes fixture-driven dry-run/e2e-style tests without requiring real Codex or GitHub services.
- [x] Validation commands pass: `go test ./...`, `gofmt -l "cmd" "internal" "namba_test.go"`, and `go vet ./...`.
- [x] Existing queue behavior outside the affected continuity paths is preserved.
