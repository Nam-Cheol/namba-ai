# SPEC-046 Harness Map

## Surfaces

- Direct PR handoff: `internal/namba/pr_land_command.go`.
- Queue PR handoff: `internal/namba/queue_command.go`.
- GitHub Actions automation: `.github/workflows/codex-review-request.yml`.
- Review marker helpers: `codexReviewRequestMarker`, `buildReviewRequestCommentBody`, and `isReviewRequestComment`.
- Generated docs and skills: `internal/namba/readme.go`, `internal/namba/templates.go`, `.agents/skills`, docs, README bundles, and `.namba/codex`.
- Guard hook: `.codex/hooks/namba_codex_guard.py`.

## State And Config

- `AutoCodexReview` remains parseable legacy state but is no longer an execution trigger for PR or queue review requests.
- `CodexReviewComment` still defines the review command text that is emitted when `--review` is explicitly present.
- Queue state under `.namba/logs/queue/` should preserve existing conveyor behavior; only PR review request triggering changes.

## Trust Boundaries

- CLI flags represent explicit operator intent.
- GitHub Actions PR events are not trusted as review request entrypoints.
- Existing PR comments are trusted only for exact normalized command or non-empty marker duplicate detection.
- Generated docs and skills must not override CLI intent by instructing automatic review requests.

## Implementation Order

1. Lock failing tests around PR and queue opt-in behavior.
2. Update PR and queue parsing plus execution gates.
3. Fix marker detection and duplicate suppression.
4. Remove workflow automation and invert workflow tests.
5. Add guard hook tests.
6. Regenerate or update docs, skills, and project artifacts.
7. Run validation and repository-wide search.
