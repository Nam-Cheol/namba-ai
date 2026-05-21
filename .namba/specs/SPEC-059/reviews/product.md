# Product Review

- Status: clear
- Last Reviewed: 2026-05-21
- Reviewer: Codex local product review
- Command Skill: `$namba-plan-pm-review`
- Recommended Role: `namba-product-manager`

## Focus

- Challenge the problem framing, scope, user value, and acceptance bar before implementation starts.

## Findings

- The product value is clear: operators need a single local health view instead of manually inspecting scattered `.namba` files.
- The scope is appropriately local-first and avoids telemetry, dashboards, databases, daemons, and external vendors.
- The user flow is concrete enough: run `namba report` for human diagnosis, run `namba report --json` or `namba status --json` for automation.
- The report must be careful not to turn a fresh workspace with no run evidence into a false failure. Missing evidence should be visible and actionable, not inherently fatal.
- The most important product outcome is cause discovery: blocked reasons, stale candidates, missing evidence, validation failures, and review readiness must be named near the top of the report.

## Decisions

- Keep `namba report` as the full observability surface.
- Keep `namba status --json` as a compact compatibility extension, not a replacement for the existing text status.
- Treat report health as advisory by default; CI failure is opt-in through `--fail-on`.
- Make JSON stable enough for CI and eval consumers by anchoring it on `schema_version: namba-report/v1`.

## Follow-ups

- During implementation, keep the default text report concise enough for a terminal operator to act on without scrolling through every detail.
- Document any future field additions as optional v1 extensions.

## Recommendation

- Proceed to implementation. Product scope is clear and aligned with NambaAI's goal of making workflow state visible before handoff or execution.
