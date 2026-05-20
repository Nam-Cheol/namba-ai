# Acceptance

- [ ] Every Namba-owned skill has concise invocation criteria.
- [ ] Skill trigger language avoids overlap between `$namba-plan`,
      `$namba-harness`, `$namba-fix`, `$namba-run`, `$namba-queue`, and
      review-related skills.
- [ ] Skill guidance maps explicit user invocations to the correct Namba command
      or guidance path.
- [ ] Skill guidance clearly labels read-only behavior versus mutating behavior.
- [ ] `.codex/agents` metadata and readable mirrors stay aligned with the
      updated skill routing contract.
- [ ] Harness evals cover unified `@` mention ambiguity across files,
      directories, plugins, and skills.
- [ ] Harness evals cover plugin/list-backed metadata ambiguity without making
      Codex mention search the Namba command router.
- [ ] Plugin packaging docs explain marketplace CLI commands as optional
      readiness paths.
- [ ] Plugin sharing docs explain version-aware sharing, share checkout, and
      shared-workspace plugin buckets as optional readiness paths.
- [ ] Default-enabled plugin hook guidance is aligned with `SPEC-050` and does
      not re-own hook schema or duplicate-guard compatibility.
- [ ] No acceptance, docs, or tests require plugin installation for local
      NambaAI workflows.
- [ ] Normal local workflows do not surface plugin, marketplace,
      remote-control, or remote-environment adoption guidance unless the user
      asks for readiness inspection or the command reports already-collected
      status evidence.
- [ ] README, Korean README, and generated docs clearly separate what is
      required for local NambaAI use from optional Codex platform readiness
      paths.
- [ ] No implementation publishes to a plugin marketplace.
- [ ] Remote-control tests cover unavailable status.
- [ ] Remote-control tests cover disabled status.
- [ ] Remote-control tests cover enabled/status read behavior.
- [ ] Remote-control tests cover configured environment snapshot behavior.
- [ ] Remote-control tests cover evidence output.
- [ ] Remote-control tests cover local fallback behavior.
- [ ] Runtime enable, disable, and status API handling remains discovery,
      status, and evidence only unless stable CLI/API behavior is confirmed.
- [ ] Stable remote-control or environment behavior is treated as confirmed only
      when backed by local CLI/help output, stable local API behavior,
      authoritative config, or existing repository code evidence.
- [ ] Remote execution is not enabled by default.
- [ ] Registry-backed or configured remote environments and environments loaded
      from `CODEX_HOME` are represented as optional readiness status or
      documentation, not as required workflow state.
- [ ] Multi-environment `apply_patch` selection is documented or evidenced
      without changing the default local workspace editing path.
- [ ] `queue_command.go`, `parallel_run.go`, `runtime_harness.go`,
      `execution.go`, and `execution_evidence.go` remain local-first and
      compatible with existing evidence consumers.
- [ ] New platform readiness evidence fields are additive to the shared
      `codex_diagnostics` payload unless implementation documents a safer
      additive alternative.
- [ ] A field-level schema map exists for every new remote-control or
      environment status, including JSON key, owner type, allowed values,
      producer, consumer surfaces, and supported evidence artifacts.
- [ ] Status source precedence is implemented or documented as stable local
      CLI/help or API signal, explicit config or `CODEX_HOME` snapshot,
      already-collected diagnostics, then neutral fallback.
- [ ] Run, queue, hook, and parallel-run evidence do not introduce blocking
      Codex probes for this SPEC.
- [ ] Status vocabulary is normalized across JSON, CLI output, README, Korean
      README, and generated docs.
- [ ] Fixture or golden tests cover presence and absence cases for any changed
      `execution-evidence/v1`, `queue-runner-evidence/v1`, and
      `project-codex-diagnostics-evidence/v1` shapes.
- [ ] `.codex/hooks/namba_codex_guard.py` remains stdlib-only.
- [ ] Any `.codex/hooks/namba_codex_guard.py` change is limited to
      readiness wording, stale-reference cleanup, or no-op safety verification.
- [ ] New hook event behavior, hook schema compatibility, and duplicate guard
      prevention remain owned by `SPEC-050`.
- [ ] Hook guard changes do not weaken dangerous command blocking.
- [ ] Hook guard changes do not weaken prompt-refinement guidance.
- [ ] Hook guard changes do not weaken approval-risk notes.
- [ ] Hook guard changes do not weaken final response checks.
- [ ] Repository search proves there are no stale Codex Python SDK imports.
- [ ] Repository search proves there are no misleading old Codex Python package
      references where the SDK is discussed.
- [ ] SDK references use `openai-codex` for the distribution and `openai_codex`
      for the import package where relevant.
- [ ] No Python runtime dependency is added unless an existing integration
      requires it.
- [ ] `README.md`, `README.ko.md`, and generated docs explain plugin packaging
      and remote-control as readiness paths rather than required workflows.
- [ ] Documentation distinguishes Codex platform features from Namba-owned
      workflow policy.
- [ ] Docs explain these concepts in a stable order: normal local command flow,
      explicit Namba skill routing, Codex unified `@` mentions, optional
      plugin/share paths, optional remote status evidence, and Python SDK rename
      only where relevant.
- [ ] Docs include a compact glossary or command-routing note for `@` mentions,
      Namba command skills, plugins, remote-control, remote environments, and
      evidence statuses.
- [ ] User-facing docs do not imply that plugin install is required,
      remote-control is expected for normal use, Namba owns Codex mention
      resolution, or remote readiness status equals workflow failure.
- [ ] Docs do not add dashboards, maturity scorecards, or gate-style readiness
      ceremonies in front of normal NambaAI usage.
- [ ] Generated docs are refreshed where Namba owns the content.
- [ ] `go test ./...` passes.
- [ ] `go vet ./...` passes.
- [ ] `gofmt -l "cmd" "internal" "namba_test.go"` reports no touched Go files
      needing formatting.
