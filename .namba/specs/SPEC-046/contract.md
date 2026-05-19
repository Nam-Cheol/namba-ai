# SPEC-046 Contract

## Review Request Contract

- Normal direct handoff: `namba pr "title"` must never post `@codex review`.
- Explicit direct handoff: `namba pr --review "title"` may post `@codex review`, exactly once.
- Normal queue handoff: `namba queue start SPEC-001` must never post `@codex review`.
- Explicit queue handoff: `namba queue start SPEC-001 --review` may post `@codex review`, exactly once during PR handoff.
- Legacy config such as `AutoCodexReview` is readable compatibility state only and must not trigger review requests.

## Deprecated Flag Contract

- `--skip-codex-review` remains accepted only on `namba queue start`.
- Because review is off by default, `--skip-codex-review` is a deprecated no-op when `--review` is absent.
- `--review` and `--skip-codex-review` together must fail with a clear conflict error.

## Marker Contract

- The Namba-owned marker is `<!-- namba:codex-review-request -->`.
- `buildReviewRequestCommentBody` emits the marker followed by the normalized review command.
- `isReviewRequestComment` matches only an exact normalized review command or a comment containing the non-empty marker.
- Unrelated comments must not suppress a requested review.

## Automation Contract

- `.github/workflows/codex-review-request.yml` is removed.
- No GitHub Actions workflow may post `@codex review` on PR opened, reopened, or ready-for-review events.
- No `workflow_dispatch` replacement is introduced.
