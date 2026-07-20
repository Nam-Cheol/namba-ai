# Product Review

- Status: clear
- Last Reviewed: 2026-07-14
- Reviewer: Codex product review
- Command Skill: `$namba-plan-pm-review`
- Recommended Role: `namba-product-manager`

## Findings

- Legacy projects must remain on `legacy-static-v1`; only an explicit policy migration activates adaptive routing.
- Dry-run, report, fallback, blocked model, and migration states are operator-facing product contracts, not incidental logs.
- A major release requires compatibility, migration/rollback guidance, and generated-artifact proof in addition to quality commands.

## Decisions

- `doctor/status` exposes migration state/action/conflicting keys and adaptive plus global `model:` fails deterministically.
- Dry-run/report expose ordered routing, recovery and observed-vs-effective state; required Sol is never silently downgraded.
- Acceptance includes legacy/new-init, conflict, documentation and clean-main release gates.

## Follow-ups

- Implement the contracts with fixtures before release preparation.

## Recommendation

- Clear for implementation.
