# SPEC-046 Baseline

## Current Behavior Evidence

- `internal/namba/pr_land_command.go` calls `ensureReviewComment` from `runPR` when `profile.AutoCodexReview` is true.
- `internal/namba/queue_command.go` calls `ensureReviewComment` during queue PR handoff when `profile.AutoCodexReview` is true and `SkipCodexReview` is false.
- `parsePRArgs` supports `--remote`, `--no-sync`, and `--no-validate`, but does not support `--review`.
- `parseQueueInvocation` supports `--skip-codex-review`, but not a positive `--review`.
- `.github/workflows/codex-review-request.yml` posts a marker plus `@codex review` on PR opened, reopened, and ready-for-review events.
- `codex_review_request_workflow_test.go` currently expects that automatic workflow to exist.
- Generated docs and skills describe normal `namba pr` handoff as guaranteeing or checking a Codex review marker.

## Known Drift To Remove

- Automatic review request behavior exists in both CLI and GitHub Actions.
- Queue currently uses a negative skip flag around an automatic default.
- Documentation frames `@codex review` as part of normal handoff rather than an explicit opt-in action.
- Project docs generated before implementation will continue to reflect current behavior until `namba sync` runs after the code and templates change.
