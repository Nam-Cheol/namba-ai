# SPEC-052 Plan

1. Refresh and verify context
   - Review the existing execution evidence schema and log-writing paths in
     `internal/namba/execution_evidence.go`, `internal/namba/execution.go`, and
     `internal/namba/runtime_harness.go`.
   - Confirm how `namba project`, `namba run`, and `namba queue` currently write
     evidence and where `.namba` logs are persisted.
   - Use `internal/namba/namba.go` as the source for `runProject` and top-level
     command dispatch, and `internal/namba/self_update_command.go` as the owner
     of `namba update`.
   - Treat `internal/namba/update_command.go` as managed-output/regen support,
     not the update command implementation.
   - Preserve the current project validation contract from
     `.namba/project/tech.md`.

2. Design the Codex diagnostics evidence contract
   - Add structured evidence for Codex version, version parse status, doctor
     status, doctor log path, timeout metadata, redaction status, repository
     roots, configured workspace roots, Codex effective workspace roots,
     sandbox mode, approval policy, and permission profile.
   - Keep all fields optional or explicitly marked unavailable so missing Codex
     produces deterministic local fallback evidence instead of workflow failure.
   - Add workspace-root mismatch reporting as advisory evidence.
   - Define the project evidence artifact as
     `.namba/logs/project/codex-diagnostics-evidence.json`.
   - Define queue canonical evidence as
     `.namba/logs/runs/<spec-id-lower>-queue-evidence.json`, while preserving
     full diagnostics in run evidence when queue uses the normal run path.
   - Define partial diagnostics as field-level statuses rather than one global
     success/failure flag.

3. Implement the safe Codex runner
   - Add a runner abstraction for `codex --version` and `codex doctor` with
     bounded timeout behavior.
   - Capture doctor stdout and stderr, redact sensitive values, and persist the
     redacted logs under the existing `.namba` evidence or log structure.
   - Treat missing Codex, unparsable version output, doctor failures, and
     timeouts as structured statuses.

4. Integrate evidence into workflows
   - Attach Codex diagnostics metadata to project evidence from
     `internal/namba/namba.go` when `runProject` refreshes docs; use
     `internal/namba/project_analysis.go` for root comparison inputs and
     generated project-doc references.
   - Attach Codex diagnostics metadata to run evidence through
     `internal/namba/execution.go`, `internal/namba/runtime_harness.go`, and the
     execution evidence writer.
   - Attach Codex diagnostics metadata to queue evidence in
     `internal/namba/queue_command.go`; the queue-runner evidence artifact must
     carry a diagnostics summary or pointers even when the full run evidence is
     stored separately.

5. Add update and baseline guidance
   - Add Codex 0.131 baseline awareness in version/update surfaces using
     `internal/namba/version.go` and `internal/namba/self_update_command.go`.
   - Touch `internal/namba/update_command.go` only if regen or generated
     Codex-guidance ownership needs to include the updated runtime docs.
   - Ensure `namba update` remains scoped to NambaAI and only advises users
     about Codex when their detected version is missing, older, equal, newer, or
     unparsable.

6. Document and regenerate contracts
   - Update `README.md` and `README.ko.md` with instructions for inspecting
     Codex doctor evidence, workspace evidence, and baseline advice.
   - Refresh `.namba/codex/README.md` as the generated runtime guidance surface
     for evidence inspection.
   - Update `.namba/codex/output-contract.md` only if the final response
     contract needs new evidence-path reporting language.
   - Keep examples compact and inspection-first: summary status, why it has that
     status, where the evidence lives, and what action, if any, is recommended.

7. Test and validate
   - Expand `internal/namba/execution_evidence_test.go` and nearby command tests
     for all required Codex runner, evidence, redaction, mismatch, and fallback
     scenarios.
   - Add command-level tests for `runProject` writing project diagnostics
     evidence and queue writing/reading the queue-runner diagnostics field.
   - Add output/update tests proving Codex baseline advice stays advisory and
     does not change NambaAI update behavior.
   - Run formatting and validation:
     - `gofmt` on touched Go files
     - `go test ./...`
     - `go vet ./...`
   - Run `namba sync` after implementation to refresh generated artifacts.
