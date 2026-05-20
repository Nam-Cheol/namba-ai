# SPEC-053 Plan

1. Refresh implementation context.
   - Re-read `SPEC-050` and `SPEC-052` so this work stays on the platform
     readiness layer instead of duplicating hook/config or diagnostics
     evidence scope.
   - Inspect `.agents/skills/*/SKILL.md`, `.codex/agents`, harness tests,
     harness evals, `internal/namba/queue_command.go`,
     `internal/namba/parallel_run.go`, `internal/namba/runtime_harness.go`,
     `internal/namba/execution.go`, `internal/namba/execution_evidence.go`,
     `.codex/hooks/namba_codex_guard.py`, `README.md`, `README.ko.md`, and
     generated docs.
   - Treat code and authoritative config as stronger evidence than generated
     docs.

2. Run plan review.
   - Run product, engineering, and design review artifacts under
     `.namba/specs/SPEC-053/reviews/`.
   - Refresh `reviews/readiness.md` and resolve contradictions or coverage
     gaps before implementation starts.

3. Update skill and agent routing guidance.
   - Inventory Namba-owned skill trigger language.
   - Make invocation criteria concise, non-overlapping, and explicit about
     command mapping.
   - Add read-only versus mutating behavior labels where missing.
   - Align `.codex/agents` metadata and readable mirrors with the updated
     routing contract.

4. Add mention and plugin readiness coverage.
   - Add harness eval scenarios for unified `@` mention ambiguity across files,
     directories, plugins, and skills.
   - Update harness contract tests so plugin/list-backed metadata and Namba
     command skills remain distinguishable.
   - Document plugin marketplace commands, version-aware sharing, share
     checkout, shared-workspace buckets, and default-enabled plugin hooks as
     optional readiness paths.
   - Treat `.codex/hooks/namba_codex_guard.py` as read-mostly. Touch it only for
     readiness wording, stale-reference cleanup, or no-op safety verification;
     route any new hook behavior back to `SPEC-050`.

5. Add remote-control and remote-environment status evidence.
   - Identify stable local help, config, or API signals for
     `codex remote-control`, runtime enable/disable/status APIs, configured
     environments, `CODEX_HOME` environments, and multi-environment
     `apply_patch` selection.
   - Define a field-level schema map before implementation: JSON key, owner
     type, allowed enum values, producer, consumer surfaces, and which evidence
     artifacts may contain the field.
   - Define source precedence for each status: stable local CLI/help output or
     API response, explicit config or `CODEX_HOME` snapshot, already-collected
     project diagnostics, then neutral fallback.
   - Add optional status and evidence fields only where stability is confirmed.
   - Cover unavailable, disabled, enabled/status read, configured environment
     snapshot, evidence output, and local fallback in tests.
   - Keep project diagnostics as the richest probe path and keep run, queue,
     hook, and parallel-run evidence non-blocking.
   - Preserve local-first queue, parallel-run, runtime-harness, and execution
     behavior.

6. Update SDK naming references.
   - Search the repository for stale Codex Python SDK imports, distribution
     names, and examples.
   - Replace misleading old references with `openai-codex` and `openai_codex`
     only where the SDK is relevant.
   - Avoid adding a Python runtime dependency unless an existing integration
     already requires it.

7. Update generated and public docs.
   - Update README, Korean README, and generated docs to frame plugins and
     remote-control as readiness paths rather than required workflows.
   - Explain unified mentions, plugin metadata, remote-control status evidence,
     configured environments, and SDK naming in compact inspection-first
     language.
   - Use the documentation order from `spec.md`: local command flow, explicit
     skill routing, unified mentions, optional plugin/share paths, optional
     remote status evidence, and SDK rename only where relevant.
   - Add a compact glossary or command-routing note and avoid dashboard,
     scorecard, or gate-style readiness framing.
   - Run the repo generation/sync path required for managed docs.

8. Validate.
   - Run targeted tests for changed skill contracts, harness evals,
     remote-control evidence, and SDK reference checks.
   - Run `go test ./...`.
   - Run `go vet ./...`.
   - Run `gofmt -l "cmd" "internal" "namba_test.go"` and format touched Go
     files if needed.
   - Run `namba sync` after implementation.

9. Final audit.
   - Confirm no plugin, marketplace, remote-control, remote-environment, or
     Python SDK path became required for local NambaAI use.
   - Confirm `SPEC-050` and `SPEC-052` ownership boundaries remain intact.
