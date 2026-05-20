# Acceptance

- [ ] `namba eval` is available as a public read-only top-level command.
- [ ] `namba eval --suite harness` runs without Codex, GitHub API, network access, live browser automation, telemetry, or LLM judge behavior.
- [ ] `namba eval --format json` emits CI-parseable JSON with summary, metrics, scenario results, failures, and baseline regression fields.
- [ ] `namba eval --format markdown` emits a human-readable report that explains failed scenarios with expected versus actual values.
- [ ] The v1 scenario schema is documented and validated.
- [ ] The v1 result JSON schema is documented and tested.
- [ ] A checked-in harness baseline exists and regression comparison can fail the command.
- [ ] At least 20 scenarios are present; target minimum is 24.
- [ ] Scenario coverage includes direct artifact generation, domain feature change, core runtime or harness change, security-sensitive change, release-related change, docs-only change, ambiguous clarification, unsafe or blocked commands, missing evidence, and review-required cases.
- [ ] Metrics cover route selection, delivery mode, artifact targets, required evidence, required reviews, clarification, SPEC field completeness, and execution readiness.
- [ ] Existing public behavior for `namba plan`, `namba harness`, `$namba-create`, `namba run`, `namba pr`, and `namba release` is unchanged.
- [ ] CI includes a `namba eval` regression check.
- [ ] Tests cover command help, invalid fixtures, baseline regressions, JSON rendering, Markdown rendering, and at least one failing-scenario diagnostic.
- [ ] Validation passes with `go test ./...`.
