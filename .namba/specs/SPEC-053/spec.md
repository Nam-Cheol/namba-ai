# SPEC-053

## Title

Codex 0.131 platform readiness for NambaAI skills, plugins, remote control,
remote environments, and SDK naming

## Problem

Codex 0.131 expands platform surfaces that NambaAI references but does not own:
unified `@` mentions across files, directories, plugins, and skills;
plugin/list-backed metadata; marketplace and share flows; default-enabled
plugin hooks; daemon-managed `codex remote-control`; runtime remote-control
enable, disable, and status APIs; registry-backed and configured remote
environments; multi-environment `apply_patch` selection; and the Python SDK
rename to the `openai-codex` distribution with the `openai_codex` import
package.

NambaAI already has adjacent Codex 0.131 work:

- `SPEC-050` owns core hook, config, permission-display, and Git-helper
  compatibility.
- `SPEC-052` owns optional local Codex diagnostics and evidence capture.

This SPEC must cover the remaining platform-readiness layer without making
plugin installation, remote execution, marketplace publication, or Python SDK
adoption mandatory for local NambaAI use.

## Goal

Prepare NambaAI for Codex 0.131 platform features by updating skills, agent
metadata, harness tests and evals, runtime/evidence status surfaces, generated
documentation, and stale SDK references while preserving the default local-first
workflow.

## Primary Outcome

Maintainers and users should be able to inspect a regenerated NambaAI repo and
understand that plugins, shared workspaces, remote-control, configured remote
environments, and the renamed Python SDK are readiness paths. None of those
paths should be presented as required for normal local `namba` planning,
execution, queueing, review, sync, PR, or release workflows.

## Context

- Project: namba-ai
- Project type: existing Go CLI and repo-orchestration toolkit
- Language: go
- Mode: tdd
- Work type: plan
- Upstream evidence: OpenAI Codex `0.131.0` release notes published on
  2026-05-18 at `https://github.com/openai/codex/releases/tag/rust-v0.131.0`.
- Local evidence: `.namba/project/product.md`, `.namba/project/tech.md`,
  `.namba/project/systems/workspace.md`, `SPEC-050`, and `SPEC-052`.
- Validation baseline: `go test ./...`, `go vet ./...`, and
  `gofmt -l "cmd" "internal" "namba_test.go"`.
- Source priority: executable code and authoritative config are stronger than
  generated docs; preserve discovered code-vs-doc conflicts instead of smoothing
  them over.

## Target Surfaces

- `.agents/skills/*/SKILL.md`
- `.codex/agents` metadata and readable mirrors where generated
- Harness contract tests
- Harness evals, especially mention/search ambiguity scenarios
- `internal/namba/queue_command.go`
- `internal/namba/parallel_run.go`
- `internal/namba/runtime_harness.go`
- `internal/namba/execution.go`
- `internal/namba/execution_evidence.go`
- `.codex/hooks/namba_codex_guard.py`, only for platform-readiness wording,
  no-op safety verification, or stale-reference cleanup that does not introduce
  new hook behavior
- Generated Codex and Namba docs
- `README.md`
- `README.ko.md`

## Scope

### Skill And Command Routing Readiness

- Update every Namba-owned skill so invocation criteria are concise,
  non-overlapping, and explicit about command mapping.
- Make read-only versus mutating behavior visible in skill guidance, especially
  for `$namba-help`, `$namba-coach`, `$namba-create`, `$namba-plan`,
  `$namba-harness`, `$namba-fix`, `$namba-run`, `$namba-queue`,
  `$namba-sync`, `$namba-pr`, `$namba-land`, `$namba-release`,
  `$namba-review-resolve`, and review-related skills.
- Avoid trigger language that makes `$namba-plan`, `$namba-harness`,
  `$namba-fix`, `$namba-run`, `$namba-queue`, and review skills compete for the
  same user request.
- Preserve the existing clarification gate for vague planning and bugfix
  planning requests.

### Unified Mention And Plugin Metadata Readiness

- Document and test how Namba guidance should interpret Codex 0.131 unified `@`
  mentions across files, directories, plugins, and skills.
- Add harness eval coverage for mention/search ambiguity, including ambiguous
  skill names, plugin names, file paths, directory paths, and mixed `@`
  references.
- Ensure Namba guidance prefers explicit skill command surfaces for Namba
  workflows while allowing Codex unified mention search to remain a platform
  feature, not a Namba parser.
- Update plugin/list-backed metadata guidance where Namba-generated skill or
  agent metadata describes discoverability.

### Plugin Packaging And Sharing Readiness

- Explain plugin packaging, marketplace CLI commands, version-aware plugin
  sharing, share checkout, clearer shared-workspace plugin buckets, and
  default-enabled plugin hooks as readiness paths.
- Do not publish any plugin to a marketplace in this SPEC.
- Do not require plugin installation for local NambaAI use.
- Keep default-enabled plugin hook guidance aligned with `SPEC-050` so this
  SPEC does not re-own hook dedupe or hook schema compatibility.
- Add tests or evals that prove plugin references do not imply marketplace
  publication, installation, or remote execution requirements.
- If `.codex/hooks/namba_codex_guard.py` is touched, limit the change to
  readiness wording, stale-reference cleanup, or no-op safety verification.
  New hook event behavior, hook schema compatibility, and duplicate guard
  prevention remain owned by `SPEC-050`.

### Remote-Control And Remote-Environment Readiness

- Add discovery, status, and evidence surfaces for daemon-managed
  `codex remote-control` only where stable CLI or API behavior is available.
- Cover unavailable, disabled, enabled/status read, configured environment
  snapshot, evidence output, and local fallback in tests.
- Treat runtime enable, disable, and status APIs as capability metadata unless a
  stable invocation is confirmed; do not enable remote execution by default.
- Represent registry-backed or configured remote environments, configured
  environments from `CODEX_HOME`, and multi-environment `apply_patch` selection
  as observable readiness status or documentation, not as required runtime
  behavior.
- Keep queue, parallel-run, runtime-harness, execution, and execution-evidence
  changes compatible with local-first operation and existing evidence
  consumers.
- Add new remote-control and remote-environment status fields through the
  existing shared `codex_diagnostics` evidence payload unless implementation
  proves that a separate additive payload is safer. Do not create parallel
  project, run, or queue schemas with divergent status meanings.
- Keep project diagnostics as the richest command-running path. Run, queue,
  hook, and parallel-run evidence must not introduce blocking Codex probes; they
  should use config snapshots, already-collected diagnostics, stable local-only
  metadata, or neutral fallback statuses.

### Python SDK Naming Readiness

- Search the repository for stale Codex Python SDK imports, package names, and
  misleading references.
- Update docs, generated guidance, tests, or examples to use
  `openai-codex` for the distribution and `openai_codex` for the import package
  where the Python SDK is actually referenced.
- Do not add a Python runtime dependency unless an existing integration truly
  requires it.

## Visibility Rules

- Normal local workflows such as `namba plan`, `namba run`, `namba queue`,
  `namba sync`, `namba pr`, `namba land`, and `namba release` must not push
  plugin, marketplace, remote-control, or remote-environment adoption guidance
  unless the user explicitly asks for readiness inspection or the command is
  reporting already-collected status evidence.
- Public docs and generated docs must separate "what is required for local
  NambaAI use" from "what Codex 0.131 can expose as optional platform
  readiness".
- Plugin and remote readiness messages should be inspection-first: status, why
  that status was chosen, where evidence lives, and whether any user action is
  optional.

## Evidence Schema And Status Source Strategy

- New platform readiness fields should be additive to the shared
  `codex_diagnostics` payload used by project, run, queue, and update/version
  advisory surfaces.
- Implementation must include a field-level schema map before changing evidence
  structs: JSON key, owner type, allowed enum values, producer, consumer
  surfaces, and whether the field may appear in project, run, queue, hook, or
  parallel-run evidence.
- Source precedence for each remote-control or environment status is:
  1. stable local CLI/help output or stable local API response
  2. explicit config or `CODEX_HOME` environment snapshot
  3. already-collected project diagnostics evidence
  4. neutral fallback status such as `unavailable` or `local_fallback`
- Run and queue evidence must not execute networked or mutating Codex probes for
  this SPEC. If status cannot be established cheaply and locally, record the
  absence as neutral evidence.
- Status vocabulary must stay normalized across JSON, CLI output, README,
  Korean README, and generated docs. Existing diagnostics terms such as
  `detected`, `not_detected`, `unavailable`, `timed_out`,
  `advisory_mismatch`, and `redacted` should be reused where they fit; new
  terms such as `disabled`, `enabled`, `configured`, and `local_fallback` must
  be defined once and tested.

## Documentation Information Architecture

Docs that discuss these Codex 0.131 surfaces should use this order where
practical:

1. Normal local Namba command flow.
2. Explicit Namba command and skill routing.
3. How Codex unified `@` mentions relate to, but do not replace, Namba routing.
4. Plugin packaging and sharing as optional readiness paths.
5. Remote-control and remote-environment status as optional evidence.
6. Python SDK rename only where SDK references are relevant.

Implementation should avoid dashboards, maturity scorecards, or ceremony that
turn readiness into a new gate in front of normal NambaAI usage.

## Out Of Scope

- Publishing NambaAI or any generated plugin to a plugin marketplace.
- Requiring users to install plugins for normal local NambaAI workflows.
- Enabling remote execution by default.
- Certifying upstream remote-control behavior beyond discovery, status, and
  evidence unless a stable local CLI/API contract is confirmed.
- Re-implementing `SPEC-050` hook/config compatibility.
- Adding new hook guard behavior; this SPEC may only align readiness wording or
  prove existing guard behavior remains unaffected.
- Re-implementing `SPEC-052` Codex diagnostics evidence.
- Adding a Python runtime dependency solely to mention the renamed SDK.
- Replacing explicit Namba command routing with Codex unified mention search.

## Compatibility Requirements

1. Local `namba` workflows remain usable without plugins, remote-control, remote
   environments, or the Python SDK.
2. Skill guidance keeps command routing deterministic and non-overlapping.
3. Remote-control and remote-environment data is optional, status-oriented, and
   represented as `unavailable`, `disabled`, `enabled`, `configured`, or
   `local_fallback` style evidence instead of as a hard failure.
4. Plugin packaging docs describe readiness and sharing paths without implying
   marketplace publication or installation.
5. Hook guard changes remain stdlib-only and do not weaken dangerous command
   blocking, prompt-refinement guidance, approval-risk notes, or final response
   checks.
6. Generated docs distinguish Codex platform features from Namba-owned workflow
   policy.

## Test Strategy

- Add or update skill contract tests for concise invocation criteria,
  read-only/mutating labels, explicit command mapping, and non-overlapping
  trigger language.
- Add harness evals for mention/search ambiguity across files, directories,
  plugins, and skills.
- Add remote-control tests for unavailable, disabled, enabled/status read,
  configured environment snapshot, evidence output, and local fallback.
- Add evidence JSON shape tests for optional remote-control and environment
  status fields where execution evidence changes.
- Add fixture or golden coverage for `execution-evidence/v1`,
  `queue-runner-evidence/v1`, and `project-codex-diagnostics-evidence/v1`
  presence and absence cases when evidence schema changes.
- Add repository search assertions or deterministic tests proving stale SDK
  imports and misleading old package references are absent.
- Run `go test ./...`, `go vet ./...`, and
  `gofmt -l "cmd" "internal" "namba_test.go"`.

## Documentation Requirements

- `README.md`, `README.ko.md`, and generated docs explain plugin packaging,
  sharing, remote-control, and remote environments as readiness paths, not
  required workflows.
- Documentation explains how unified `@` mentions interact with Namba command
  skills without claiming that Namba owns Codex mention search.
- Documentation explains what remote-control evidence can mean when the feature
  is unavailable, disabled, enabled, configured, or falling back to local
  execution.
- Documentation explains the Python SDK rename only where SDK references are
  relevant.
- Documentation includes a compact glossary or command-routing note covering
  `@` mentions, Namba command skills, plugins, remote-control, remote
  environments, and evidence statuses.
- Documentation must not imply that plugin installation is required,
  remote-control is expected for normal use, Namba owns Codex mention
  resolution, or remote readiness status equals workflow failure.

## Risks

- Remote-control CLI/API behavior may remain unstable; mitigate by limiting
  implementation to discovery, status, and evidence unless stability is
  confirmed in local help or code.
- Plugin marketplace language can sound like a requirement; mitigate with
  explicit local-first and optional-adoption wording.
- Unified mention language can blur Namba skill routing; mitigate with harness
  evals and concise, non-overlapping trigger criteria.
- SDK rename work can accidentally add runtime dependencies; mitigate with
  repository search and docs-only updates unless code requires otherwise.
