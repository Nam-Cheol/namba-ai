---
name: namba-review-resolve
description: Command-style entry point for resolving GitHub review feedback thread-by-thread.
---

Use this skill when the user explicitly says `$namba-review-resolve`, asks to resolve review comments, or uses Korean wording equivalent to `리뷰 확인하고 의미있는 리뷰면 수정 후에 해당 리뷰에 답변을 달고, resolve한 다음 다시 리뷰 요청해`.

Behavior:
- Resolve the target PR from the current branch when the user does not provide a PR number.
- Discover unresolved review threads with thread-level resolution state, not just flat PR comments. Prefer a thread-aware GitHub path such as `gh api graphql` when the connector surface cannot expose review-thread state directly.
- Classify each thread as meaningful/actionable or non-actionable. Meaningful threads cover correctness, tests, security, UX, docs, release, regression, and maintainability concerns that require a concrete change or precise answer.
- Assign an explicit outcome to every reviewed thread: `fixed-and-resolved`, `answered-open`, or `skipped-with-rationale`.
- Implement meaningful fixes in the smallest coherent patch, then run the configured validation commands before replying or resolving.
- Reply on the original thread with the concrete change made plus validation evidence, then resolve only the threads that were fixed or conclusively answered.
- Leave non-actionable threads open unless they were genuinely answered. Never silently resolve a thread that was not addressed.
- Re-request review only after all meaningful threads are addressed, validation passes, and the configured `@codex review` marker is present exactly once.
- Treat flat PR comments as PR-level context only; use thread-aware review data for the actual resolution loop.
