# Acceptance

- [ ] `internal/namba/execution_evidence.go` exposes structured Codex diagnostics
  evidence for version, doctor status, doctor log path, timeout metadata,
  redaction status, Namba-detected root, Git root, configured workspace roots,
  Codex effective workspace roots when available, `sandbox_mode`,
  `approval_policy`, and permission profile.
- [ ] The same diagnostics payload is reusable across project, run, queue, and
  update/version advisory surfaces without forcing a global success/failure
  state for partial diagnostics.
- [ ] The Codex runner is optional and safe: missing Codex does not block
  project, run, or queue workflows unless the selected workflow explicitly
  requires Codex.
- [ ] `codex --version` handling covers parseable, unparsable, older, equal, and
  newer versions relative to the Codex 0.131 advisory baseline.
- [ ] `codex doctor` handling covers success, failure, timeout, and local
  fallback status.
- [ ] Doctor stdout and stderr are persisted under the existing `.namba` evidence
  or log structure, and persisted output redacts auth-sensitive or
  local-secret-looking values.
- [ ] Evidence JSON shape tests cover available diagnostics, missing diagnostics,
  redaction metadata, workspace root mismatch reporting, and local fallback.
- [ ] `namba project` writes
  `.namba/logs/project/codex-diagnostics-evidence.json` with schema
  `project-codex-diagnostics-evidence/v1` and Codex diagnostics metadata when
  available.
- [ ] `namba run` evidence includes Codex diagnostics metadata when available.
- [ ] `namba queue` includes Codex diagnostics metadata or a queue-safe summary
  with log pointers in the canonical
  `.namba/logs/runs/<spec-id-lower>-queue-evidence.json` handoff artifact, and
  preserves the full payload in run evidence when the normal run path is used.
- [ ] `namba update` and self-update remain NambaAI-only while giving clear
  advisory Codex baseline guidance.
- [ ] Workspace-root mismatches are persisted as advisory evidence and, when
  printed, use concise non-blocking CLI language that names compared roots when
  safe.
- [ ] Codex baseline guidance appears only in update/version-oriented surfaces,
  not in every workflow execution path.
- [ ] `README.md`, `README.ko.md`, and generated `.namba/codex/README.md`
  explain how to inspect Codex doctor logs and workspace evidence with exact
  paths or command examples; `.namba/codex/output-contract.md` is updated only
  if final-response evidence reporting changes.
- [ ] CLI, evidence JSON, English docs, Korean docs, and generated runtime
  guidance use a shared neutral status vocabulary for optional diagnostics.
- [ ] Tests include deterministic coverage for Codex missing, parseable version,
  unparsable version, older/equal/newer version, doctor success, doctor failure,
  timeout, redaction, evidence JSON shape, workspace root mismatch reporting,
  and local fallback.
- [ ] `go test ./...` passes.
