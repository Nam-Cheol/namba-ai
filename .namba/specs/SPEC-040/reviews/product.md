# Product Review

- Status: clear
- Last Reviewed: 2026-04-27
- Reviewer: Codex acting as `namba-product-manager`
- Command Skill: `$namba-plan-pm-review`
- Recommended Role: `namba-product-manager`

## Focus

- Challenge the problem framing, scope, user value, and acceptance bar before implementation starts.

## Findings

- The user value is clear: operators need richer Namba skill guidance for review, CI, frontend validation, and harness/MCP evaluation without losing the repo-local, managed-skill discipline that prevents duplicate skill discovery and unmanaged install flows.
- The scope is correctly constrained to existing Namba-managed skills, templates, docs, and optional helper candidates. The SPEC explicitly rejects third-party skill installation, Composio CLI/OAuth, Slack/Notion/app automation, and `$CODEX_HOME/skills`.
- The upstream audit is framed as pattern input rather than import authority. That keeps product ownership with Namba and avoids turning this into a third-party skill-vendoring slice.
- Acceptance is strong enough for product review because it names the target operator outcomes: sharper triggers, progressive disclosure, CI/thread evidence, frontend validation evidence, batch-migration discipline, and harness/MCP quality criteria.
- The helper-script posture is appropriately conservative. Treating scripts as candidates protects the product surface from premature tooling while still allowing deterministic helpers when repeated manual steps become costly.

## Decisions

- Keep this as one core-owned Namba skill-surface evolution SPEC. Splitting PR/review guidance, frontend validation, and MCP criteria would make the implementation less coherent because the common product theme is safe skill-surface evolution.
- Preserve the "no new standalone skill" rule as a hard product boundary for this slice.
- Keep helper scripts optional until implementation proves they are deterministic, small, tested, and useful beyond prose guidance.
- Treat `core_harness_change` metadata as correct despite the `$namba-harness` entrypoint, because the implementation will touch Namba-managed templates.

## Follow-ups

- During implementation, verify the final diff still reads as skill-surface evolution, not as new automation-product scope.
- If helper scripts are added, make their user-facing names and outputs operator-readable enough to support PR handoff summaries.
- Keep the Korean PR defaults and existing collaboration policy unchanged.

## Recommendation

- Clear for implementation. Product risk is bounded by the explicit exclusion list, managed-template ownership, and acceptance criteria.
