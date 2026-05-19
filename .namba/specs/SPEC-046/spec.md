# SPEC-046

Title: Improve NambaAI harness reliability and make Codex PR review explicitly opt-in

## Problem

NambaAI currently treats Codex PR review as part of the default handoff path in more than one place. The PR command calls the review-comment path when legacy `AutoCodexReview` config is enabled, the queue PR handoff has a negative `--skip-codex-review` escape hatch around the same automatic behavior, and `.github/workflows/codex-review-request.yml` can post `@codex review` on PR events. That makes the review request contract harder to reason about and lets hidden automation override user intent.

The marker contract also needs to be deterministic. A review request should be considered present only when a PR comment is exactly the configured review command after normalization, or when it contains a non-empty Namba-owned marker. Unrelated comments must never suppress a requested review.

## Goal

Make Codex review request behavior explicit, deterministic, tested, and documented:

- `namba pr "title"` creates or reuses a PR without posting `@codex review`.
- `namba pr --review "title"` is the only direct PR command path that requests Codex review.
- `namba queue start SPEC-001` creates or reuses queue PRs without posting `@codex review`.
- `namba queue start SPEC-001 --review` is the only queue path that requests Codex review during PR handoff.
- GitHub Actions must not post `@codex review` automatically for PR opened, reopened, or ready-for-review events.

## Scope

- Add positive `--review` support to `namba pr`.
- Add positive `--review` support to `namba queue start`.
- Keep `--skip-codex-review` as a deprecated compatibility flag for queue, but make it a no-op because no-review is now the default.
- Return an explicit conflict error if queue receives both `--review` and `--skip-codex-review`.
- Stop using `profile.AutoCodexReview` or equivalent legacy config to trigger PR or queue review requests.
- Replace the review request marker with a non-empty marker: `<!-- namba:codex-review-request -->`.
- Delete `.github/workflows/codex-review-request.yml`; no `workflow_dispatch` replacement is allowed.
- Update CLI help, README output, localized workflow guides, generated docs, Namba skills, Codex docs, and tests that currently describe automatic review handoff.
- Add a small Phase 1 regression slice for `.codex/hooks/namba_codex_guard.py`.

## Non-goals

- Do not redesign the whole NambaAI architecture.
- Do not build a full benchmark or eval platform in this SPEC.
- Do not change `namba land` semantics unless a direct compile-time dependency requires a tiny adjustment.
- Do not preserve automatic Codex review through config, generated docs, queue defaults, GitHub Actions, or skill instructions.
- Do not introduce another non-CLI review request entrypoint.

## Evidence

- `internal/namba/pr_land_command.go` currently calls `ensureReviewComment(...)` from `runPR` when `profile.AutoCodexReview` is enabled.
- `internal/namba/queue_command.go` currently calls `ensureReviewComment(...)` during queue PR handoff when `profile.AutoCodexReview` is enabled and `--skip-codex-review` is absent.
- `.github/workflows/codex-review-request.yml` currently posts `@codex review` for default PR events.
- `codex_review_request_workflow_test.go` and `workflow_test.go` currently encode the workflow-based automatic review request contract.
- `README.md`, `docs/workflow-guide*.md`, `.agents/skills/namba-pr/SKILL.md`, `.agents/skills/namba-queue/SKILL.md`, `.namba/codex/README.md`, `internal/namba/readme.go`, and `internal/namba/templates.go` contain default-review language that must move to explicit opt-in language.

## Requirements

### PR Command

- Extend `prOptions` with `RequestReview bool`.
- Parse `--review` in `parsePRArgs`.
- Preserve existing unknown-flag behavior.
- Preserve composition with `--remote`, `--no-sync`, and `--no-validate`.
- Call `ensureReviewComment` only when `opts.RequestReview` is true.
- Ignore `profile.AutoCodexReview` for PR review request decisions.
- Print success text that distinguishes normal PR handoff from explicit Codex review request handoff.

### Queue Command

- Extend queue options with `RequestReview bool`.
- Parse `namba queue start ... --review`.
- Keep `--skip-codex-review` accepted as deprecated compatibility syntax, but no-op when `--review` is absent.
- Fail when `--review` and `--skip-codex-review` are both present.
- Call `ensureReviewComment` only when queue `RequestReview` is true.
- Ignore `profile.AutoCodexReview` for queue review request decisions.
- Update queue usage text and generated queue docs.

### Review Marker

- Use `<!-- namba:codex-review-request -->` as the Namba-owned marker.
- `buildReviewRequestCommentBody` must emit the marker plus the normalized review command.
- `isReviewRequestComment` must return true only when the comment exactly matches the normalized review command or contains the non-empty marker.
- Arbitrary unrelated comments must not be treated as existing review requests.

### GitHub Workflow

- Remove `.github/workflows/codex-review-request.yml`.
- Remove or rewrite workflow tests so the repository proves no default PR opened, reopened, or ready-for-review workflow posts `@codex review`.
- Do not keep a manual `workflow_dispatch` replacement.

### Harness Regression Slice

- Add or update tests for `.codex/hooks/namba_codex_guard.py`.
- Cover ambiguous prompt guidance, clear prompt no unnecessary guidance, dangerous bash command denial, and safe bash command allowance.
- Keep golden eval coverage as follow-up work, not part of this SPEC.

## Implementation Targets

- `internal/namba/pr_land_command.go`
- `internal/namba/pr_land_command_test.go`
- `internal/namba/queue_command.go`
- `internal/namba/queue_command_test.go`
- `.github/workflows/codex-review-request.yml`
- `codex_review_request_workflow_test.go`
- `workflow_test.go`
- `.codex/hooks/namba_codex_guard.py`
- Hook guard tests under the existing Python test convention or a documented new test command
- `README.md`
- `docs/workflow-guide.md`
- `docs/workflow-guide.ko.md`
- `docs/workflow-guide.ja.md`
- `docs/workflow-guide.zh.md`
- `internal/namba/readme.go`
- `internal/namba/templates.go`
- `.agents/skills/namba-pr/SKILL.md`
- `.agents/skills/namba-queue/SKILL.md`
- `.agents/skills/namba-run/SKILL.md`
- `.agents/skills/namba-workflow-execution/SKILL.md`
- `.namba/codex/README.md`
- Generated tests that assert PR checklist, README, template, and skill wording

## Risks

- Hidden automation could still post `@codex review`; mitigate with repository-wide search and workflow regression tests.
- Users who relied on automatic review may miss the new opt-in step; mitigate with clear CLI help and docs for `--review`.
- Duplicate detection could suppress requested review incorrectly; mitigate with exact-command, marker, and unrelated-comment tests.
- Hook guard tests may expose runtime payload assumptions; fix the narrow guard behavior if tests reveal inverted branches.

## Follow-up Backlog

- Add golden eval coverage for route classification, prompt refinement, evidence generation, and guardrail behavior.
- Add `go test -race ./...`, `staticcheck`, `govulncheck`, and installer `shellcheck` to CI.
- Add installer checksum and signature verification.
- Add supply-chain provenance and SBOM for release artifacts.
- Remove or justify runtime artifacts such as dump files if they remain in the repository.
