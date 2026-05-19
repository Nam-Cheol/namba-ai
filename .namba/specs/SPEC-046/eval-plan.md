# SPEC-046 Eval Plan

## Targeted Regression Tests

- PR args: parse without `--review`.
- PR args: parse with `--review`.
- PR args: compose `--review` with `--remote`, `--no-sync`, and `--no-validate`.
- PR behavior: no review comment without `--review`, even when `AutoCodexReview` is true.
- PR behavior: one review comment with `--review`.
- PR behavior: no duplicate review when an exact command or the marker already exists.
- PR behavior: unrelated comments do not suppress review requests.
- Queue args: parse default no-review behavior.
- Queue args: parse `--review`.
- Queue args: accept deprecated `--skip-codex-review` as no-op.
- Queue args: fail on `--review` plus `--skip-codex-review`.
- Queue behavior: no review comment by default, even when `AutoCodexReview` is true.
- Queue behavior: review comment only when `--review` is present.
- Workflow behavior: no default PR-event GitHub Actions workflow posts `@codex review`.
- Hook guard behavior: ambiguous prompt guidance, clear prompt no unnecessary guidance, dangerous bash denial, and safe bash allowance.

## Validation Commands

- `gofmt` check for Go files touched by the implementation.
- Targeted Go tests for PR, queue, workflow, docs, and template expectations.
- `go test ./...`.
- `go vet ./...` if it remains configured validation.
- Python hook tests, either wired into CI or documented as a required validation command.
- Repository search for `@codex review`, `AutoCodexReview`, `skip-codex-review`, `codexReviewRequestMarker`, and `Codex review marker`.

## Out Of Scope

- Golden eval framework for route classification, prompt refinement, evidence generation, and guardrail behavior.
- Race testing, static analysis, vulnerability scanning, shellcheck, installer signature verification, and SBOM generation.
