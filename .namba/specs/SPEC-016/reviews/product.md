# Product Review

- Status: approved
- Last Reviewed: 2026-03-29
- Reviewer: Codex
- Command Skill: `$namba-plan-pm-review`
- Recommended Role: `namba-product-manager`

## Focus

- Challenge the problem framing, scope, user value, and acceptance bar before implementation starts.

## Findings

- The problem framing is strong. For Namba users, the failure today is not "Codex support is missing" but "Codex behavior is too implicit and too shallow," which shows up as confusing execution semantics, unclear retries, misleading delegation, and parallelism that is named more strongly than it behaves.
- The scope is still product-coherent as one SPEC because every slice improves the same user promise: `namba run` should behave like a trustworthy Codex-native runtime rather than an opaque shell wrapper.
- The highest-value user outcomes are truthful mode semantics, clearer failure recovery, and better observability. Internal enablers such as live smoke coverage matter, but they should not dominate delivery sequencing over those user-facing trust improvements.
- The session-refresh requirement is product-relevant, not just technical cleanup. If Namba regenerates instructions and users keep talking to a stale session, they will experience the tool as inconsistent even if the underlying files are correct.

## Decisions

- Keep the work in one SPEC under the product theme of runtime trust and operator clarity.
- Prioritize user-visible truthfulness of existing entrypoints over introducing new command surface area. The product win is that `namba run`, `--solo`, `--team`, and `--parallel` finally mean what the documentation says they mean.
- Treat session refresh messaging as part of the user contract. When Namba changes instruction surfaces that a running Codex session may not re-read, the user should be told explicitly and immediately.
- Treat opt-in live smoke coverage as a delivery confidence tool, not a product-facing acceptance gate. The shipped experience should not require users to opt into heavier validation to get predictable runtime behavior.

## Follow-ups

- During implementation, keep the default `namba run SPEC-XXX` flow comprehensible for users who do not care about Codex internals; advanced runtime metadata should improve debugging without turning the default UX into operator jargon.
- Document degraded or fallback session behavior in plain language. Users need to know when Namba kept one continuous runtime versus when it had to fall back to a less capable recovery path.
- Make sure backward compatibility stays explicit for existing entrypoints, especially `namba run SPEC-XXX`, `--dry-run`, `--solo`, `--team`, and `--parallel`, so the product value is higher trust rather than surprise.
- Make README updates part of the delivery slice, not an optional follow-up, so users can discover the revised mode semantics without reading release notes or source code.

## Recommendation

- Advisory recommendation: approved. Proceed with implementation, but sequence delivery around user-facing trust improvements first: truthful mode behavior, repair-loop clarity, session-refresh messaging, and observable runtime metadata.
