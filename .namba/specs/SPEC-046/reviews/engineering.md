# Engineering Review

- Status: approved
- Last Reviewed: 2026-05-19
- Reviewer: Codex as `namba-planner`
- Command Skill: `$namba-plan-eng-review`
- Recommended Role: `namba-planner`

## Focus

- Lock architecture, sequencing, failure modes, trust boundaries, and validation strategy before execution starts.

## Findings

- The main code paths are bounded: `internal/namba/pr_land_command.go` owns direct PR handoff, and `internal/namba/queue_command.go` owns queue PR handoff.
- The most important failure mode is hidden automation. The SPEC addresses all known sources: PR command, queue command, legacy config, GitHub Actions workflow, generated docs, and skills.
- Marker handling should be treated as behavior, not formatting. Tests must cover exact normalized command, non-empty marker, old duplicate cases if compatibility is desired, and unrelated comments.
- Workflow deletion affects root-level tests. The old workflow test should be removed or inverted so the suite protects against reintroducing automatic PR-event comments.
- Hook guard tests are appropriately small for this slice. They improve harness reliability without turning the SPEC into a full eval framework.

## Decisions

- Implement `RequestReview` flags rather than reinterpreting `AutoCodexReview`.
- Leave legacy config parsing intact unless dead-code removal is explicitly safe; document it as no-op for PR and queue review triggering.
- Make queue `--review` plus `--skip-codex-review` a hard argument conflict.
- Prefer targeted tests around existing fake command runners before broader integration work.

## Follow-ups

- Verify whether any generated PR checklist still requires checking an `@codex review` marker during normal handoff.
- Verify whether old marker comments should still prevent duplicate comments for existing open PRs. If compatibility is desired, add a test for the old marker during migration; otherwise document the marker change as intentionally strict.

## Recommendation

- Approved for implementation with the test-first sequencing in `plan.md`.
