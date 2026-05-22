# Product Review

- Status: clear
- Last Reviewed: 2026-05-22
- Reviewer: `namba-product-manager`
- Command Skill: `$namba-plan-pm-review`
- Recommended Role: `namba-product-manager`

## Focus

- Challenge the problem framing, scope, user value, and acceptance bar before implementation starts.

## Findings

- The user value is clear: operators and Codex-facing workflows get a low-friction hint when the installed NambaAI CLI appears behind, without turning advisory discovery into an update action.
- The scope correctly targets CLI maintenance surfaces: `doctor`, `status`, `regen`, `project`, and `sync`.
- The request explicitly excludes auto-update behavior, telemetry, non-GitHub external calls, and noise in JSON or automation-oriented flows.
- The most important product contract is consent: advisory text may recommend `namba update`, but Codex-facing guidance must tell Codex to ask the user first.
- The acceptance criteria now distinguish behind, current, dev, offline, stale, cache-miss, lookup failure, and parse failure states, which is enough to prevent vague "check for updates" behavior from expanding during implementation.

## Decisions

- Keep this as advisory-only CLI messaging, not a self-update automation feature.
- Treat `namba doctor --check-update` or equivalent as the explicit user-requested refresh path.
- Keep default maintenance commands quiet unless trustworthy cached or explicitly refreshed data proves a behind state.
- Keep `namba update` distinct from upstream `codex update` in every surfaced next step.

## Follow-ups

- During implementation, verify the exact behind advisory says to ask the user before running `namba update`.
- During implementation, verify docs generated from `internal/namba/readme.go` and Codex-facing skill/template text carry the same consent language.
- During implementation, verify current and dev builds are not nagged.

## Recommendation

- Cleared for implementation.
