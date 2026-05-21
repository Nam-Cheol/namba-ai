# Acceptance

- [x] `namba eval` is available as a public read-only top-level command.
- [x] `namba eval --suite harness` runs without Codex, GitHub API, network access, live browser automation, telemetry, or LLM judge behavior.
- [x] `namba eval --format json` emits CI-parseable JSON with summary, metrics, scenario results, failures, and baseline regression fields.
- [x] `namba eval --format markdown` emits a human-readable report that explains failed scenarios with expected versus actual values.
- [x] The v1 scenario schema is documented and validated.
- [x] The v1 result JSON schema is documented and tested.
- [x] A checked-in harness baseline exists and regression comparison can fail the command.
- [x] At least 20 scenarios are present; target minimum is 24.
- [x] Scenario coverage includes direct artifact generation, domain feature change, core runtime or harness change, security-sensitive change, release-related change, docs-only change, ambiguous clarification, unsafe or blocked commands, missing evidence, and review-required cases.
- [x] Metrics cover route selection, delivery mode, artifact targets, required evidence, required reviews, clarification, SPEC field completeness, and execution readiness.
- [x] Existing public behavior for `namba plan`, `namba harness`, `$namba-create`, `namba run`, `namba pr`, and `namba release` is unchanged.
- [x] CI includes a `namba eval` regression check.
- [x] Tests cover command help, invalid fixtures, baseline regressions, JSON rendering, Markdown rendering, and at least one failing-scenario diagnostic.
- [x] Validation passes with `go test ./...`.
