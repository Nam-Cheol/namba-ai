# Product Review

- Status: approved
- Last Reviewed: 2026-05-19
- Reviewer: Codex as `namba-product-manager`
- Command Skill: `$namba-plan-pm-review`
- Recommended Role: `namba-product-manager`

## Focus

- Challenge the problem framing, scope, user value, and acceptance bar before implementation starts.

## Findings

- The product goal is coherent: Codex review requests are an externally visible handoff action, so requiring explicit `--review` removes surprise and makes PR behavior easier to explain.
- Queue behavior is now aligned with direct PR behavior: default no-review, positive `--review` opt-in, and deprecated `--skip-codex-review` retained only for compatibility.
- Deleting the GitHub Actions workflow is the right product constraint because keeping a second review-request entrypoint would weaken the CLI-only contract.
- The acceptance list is strong enough to catch the main user-visible regressions: normal handoff no longer posts review, explicit handoff posts one review, queue mirrors the same behavior, and legacy config cannot silently opt users in.

## Decisions

- Treat `--review` as the only source of truth for Codex review requests in `namba pr` and queue PR handoff.
- Keep `--skip-codex-review` for queue compatibility only; document it as deprecated and no-op under the new default.
- Do not provide a manual GitHub Actions workflow replacement.

## Follow-ups

- After implementation, consider a release note or migration note for users who previously relied on automatic Codex review.
- Track golden eval coverage in the follow-up backlog rather than expanding this SPEC.

## Recommendation

- Approved for implementation. The scope is focused and the acceptance criteria capture the product contract.
