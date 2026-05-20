# SPEC-052

## Problem

NambaAI already records workflow evidence, but it does not consistently capture
Codex environment, diagnostics, workspace, and permission context in a structured
shape for project, run, and queue workflows. That makes it harder to explain why
a Codex-backed workflow behaved differently across machines, whether Codex was
absent, whether `codex doctor` found local issues, or whether workspace roots and
permission settings matched the repository's expectations.

Codex 0.131 introduces baseline context that NambaAI should understand as
advisory evidence. NambaAI should surface that awareness without taking
ownership of Codex installation or upgrades.

## Goal

Add local-first Codex diagnostics evidence integration so NambaAI can safely
capture Codex version, doctor output metadata, workspace roots, sandbox and
approval context, and permission profile details when those signals are
available. The evidence must be optional, redacted, deterministic in tests, and
non-blocking for workflows that do not explicitly require Codex.

The MVP product slice is: structured Codex diagnostics evidence is captured
when available and saved in existing `.namba` evidence or log locations without
blocking local workflows. Workspace-root mismatch messaging and Codex baseline
guidance are secondary advisory surfaces layered on that same evidence model.

## Context

- Project: namba-ai
- Project type: existing
- Language: go
- Mode: tdd
- Work type: plan
- Validation: `go test ./...`, `gofmt -l "cmd" "internal" "namba_test.go"`,
  and `go vet ./...` are the project-level quality commands from
  `.namba/project/tech.md`.
- Source priority: code and config are stronger evidence than generated docs;
  preserve code-vs-doc conflicts if implementation discovers them.

## Scope

Primary implementation files:

- `internal/namba/execution_evidence.go`
- `internal/namba/execution_evidence_test.go`
- `internal/namba/execution.go`
- `internal/namba/runtime_harness.go`
- `internal/namba/queue_command.go`
- `internal/namba/namba.go`
- `internal/namba/project_analysis.go`
- `internal/namba/self_update_command.go`
- `internal/namba/update_command.go`
- `internal/namba/version.go`
- `README.md`
- `README.ko.md`
- generated `.namba/codex` runtime guidance documentation

Functional scope:

- Add an optional safe runner for `codex --version` and `codex doctor`.
- Persist `codex doctor` stdout and stderr under the existing `.namba` log or
  evidence structure.
- Include these fields in evidence JSON when available:
  - Codex version and parse status
  - Codex doctor status
  - Codex doctor log path
  - Codex doctor timeout metadata
  - redaction status
  - Namba-detected repository root
  - Git root
  - configured workspace roots
  - Codex effective workspace roots when available
  - `sandbox_mode`
  - `approval_policy`
  - permission profile
- Report workspace-root mismatches without blocking local workflows.
- Add advisory Codex baseline awareness for features introduced in Codex 0.131.
- Keep `namba update` and self-update behavior NambaAI-only; Codex version
  guidance must remain advice, not package management.

Artifact ownership:

- Shared Codex diagnostics payload: add one reusable payload type and builder
  used by project, run, queue, and update/version advisory surfaces.
- Run evidence: embed the full payload in
  `.namba/logs/runs/<log-id>-evidence.json`.
- Project evidence: `namba project` writes a project-scoped diagnostics artifact
  at `.namba/logs/project/codex-diagnostics-evidence.json` with schema
  `project-codex-diagnostics-evidence/v1`, plus the reusable payload and any
  project-analysis root comparison metadata.
- Queue evidence: the queue-owned canonical handoff artifact is
  `.namba/logs/runs/<spec-id-lower>-queue-evidence.json` with schema
  `queue-runner-evidence/v1`; it must include either the reusable payload or a
  queue-safe summary with pointers to doctor logs. When queue runs through the
  normal run path, the matching run evidence also carries the full payload.
- Doctor logs: redacted `codex doctor` stdout and stderr live beside the
  evidence artifact that triggered them, using deterministic names such as
  `.namba/logs/runs/<log-id>-codex-doctor-stdout.txt`,
  `.namba/logs/runs/<log-id>-codex-doctor-stderr.txt`, or
  `.namba/logs/project/codex-doctor-stdout.txt`.

Surface ownership:

- `internal/namba/namba.go` owns top-level command dispatch and `runProject`.
- `internal/namba/project_analysis.go` owns project-analysis docs and root
  comparison inputs, but project diagnostics persistence is called from
  `runProject`.
- `internal/namba/self_update_command.go` owns `namba update`; it may print
  Codex baseline advice but must not install or update Codex.
- `internal/namba/update_command.go` owns regen/managed-output plumbing and is
  only in scope when generated Codex guidance or managed-output ownership needs
  adjustment.
- `.namba/codex/README.md` is the generated runtime guidance surface for
  evidence inspection. `.namba/codex/output-contract.md` remains response-shape
  guidance and should be updated only if the final response contract needs to
  mention the new evidence paths.

Advisory visibility rules:

- Workspace-root mismatches are always persisted in evidence. CLI output may
  emit one concise advisory summary that names the compared roots when safe and
  explicitly states that the workflow was not blocked.
- Codex baseline guidance appears only in update/version-oriented surfaces,
  not in every project, run, or queue execution path.
- Partial diagnostics availability is normal. Missing version output, doctor
  failure, timeout, unavailable effective workspace roots, or unavailable
  permission profile are represented field-by-field instead of collapsed into
  one global failure state.
- Status vocabulary must stay neutral and shared across CLI output, evidence
  JSON, `README.md`, `README.ko.md`, and `.namba/codex/README.md`; prefer terms
  like `detected`, `not_detected`, `unavailable`, `timed_out`,
  `advisory_mismatch`, and `redacted`.

## Constraints

- Do not assume network access.
- Do not block project, run, or queue workflows when Codex is absent unless the
  selected workflow explicitly requires Codex.
- Redact authentication-sensitive or local-secret-looking output before
  persisting doctor logs or evidence snippets.
- Do not infer permission escalation from free-form assistant text.
- Prefer explicit config, CLI output, structured status data, or deterministic
  local state over natural-language transcript inference.
- Preserve local-first behavior and deterministic tests.
- Keep evidence schema additions compatible with existing evidence consumers.
- Do not treat `.namba/codex/output-contract.md` as the runtime evidence schema
  reference.

## Non-Goals

- Installing, updating, or managing Codex through `namba update`.
- Requiring Codex for NambaAI workflows that can run locally without it.
- Using network calls to validate Codex state.
- Treating assistant prose as an authoritative permissions source.
