# SPEC-046 Plan

1. Establish the baseline.
   - Search for `@codex review`, `AutoCodexReview`, `skip-codex-review`, `codexReviewRequestMarker`, and Codex review marker language.
   - Run the targeted PR and queue tests before implementation when practical to confirm the current automatic behavior fails the new contract.

2. Implement PR command opt-in.
   - Add `RequestReview bool` to `prOptions`.
   - Parse `--review` in `parsePRArgs`, including composition with `--remote`, `--no-sync`, and `--no-validate`.
   - Change `runPR` so `ensureReviewComment` is called only when `opts.RequestReview` is true.
   - Stop using `profile.AutoCodexReview` for PR review requests.
   - Update success output for normal handoff versus explicit Codex review handoff.
   - Add tests for parsing, normal no-review behavior, explicit review behavior, idempotency, and legacy config no-op behavior.

3. Implement queue opt-in.
   - Add `RequestReview bool` to queue options.
   - Parse `namba queue start ... --review`.
   - Keep `--skip-codex-review` accepted as deprecated compatibility syntax.
   - Make `--skip-codex-review` no-op when `--review` is absent.
   - Return an explicit conflict error when `--review` and `--skip-codex-review` are both present.
   - Change queue PR handoff so `ensureReviewComment` is called only when queue `RequestReview` is true.
   - Add queue parser and handoff tests for default no-review, explicit review, deprecated skip flag, and conflict behavior.

4. Fix review marker detection.
   - Replace the marker with `<!-- namba:codex-review-request -->`.
   - Ensure `buildReviewRequestCommentBody` emits marker plus normalized command.
   - Ensure `isReviewRequestComment` matches only the normalized exact command or the non-empty marker.
   - Add regression tests proving unrelated comments do not suppress a review request.

5. Remove GitHub Actions auto-review.
   - Delete `.github/workflows/codex-review-request.yml`.
   - Remove or rewrite `codex_review_request_workflow_test.go`.
   - Update `workflow_test.go` so default PR event workflows are not allowed to post `@codex review`.
   - Do not add `workflow_dispatch` or any other workflow-based review request path.

6. Add guard hook regression coverage.
   - Add or update tests for `.codex/hooks/namba_codex_guard.py`.
   - Cover ambiguous prompt guidance, clear prompt no unnecessary guidance, dangerous bash command denial, and safe bash command allowance.
   - If the repository does not already run Python hook tests in CI, document the required validation command or add it to CI in a minimal way.

7. Update docs and generated artifacts.
   - Update CLI help text for `namba pr` and `namba queue`.
   - Update README, localized workflow guides, `.namba/codex/README.md`, PR checklist output, command skills, and generated templates.
   - Remove normal-handoff wording that says `namba pr` guarantees or always creates a Codex review marker.
   - Document `AutoCodexReview` as legacy readable config that no longer triggers PR or queue review requests.

8. Validate and sync.
   - Run `gofmt` and confirm no diff remains.
   - Run targeted PR, queue, workflow, and hook tests.
   - Run `go test ./...`.
   - Run `go vet ./...` if still configured.
   - Run the Python hook tests or document the command as required validation.
   - Run the repository-wide search terms from step 1 and confirm all remaining occurrences match the new opt-in contract.
   - Run `namba sync` after implementation so generated docs and readiness artifacts reflect the final behavior.
