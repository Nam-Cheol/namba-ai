# Engineering Review

- Status: clear
- Last Reviewed: 2026-05-22
- Reviewer: `namba-planner`
- Command Skill: `$namba-plan-eng-review`
- Recommended Role: `namba-planner`

## Focus

- Lock architecture, sequencing, failure modes, trust boundaries, and validation strategy before execution starts.

## Findings

- The safest implementation path is to preserve profile detection and flag overrides, then change only the interactive presentation/state-machine layer.
- Language-first ordering touches `runInitWizard`, `nextWizardStep`, `previousWizardStep`, visible step numbering, and tests; this is now called out in the plan as a navigation test matrix.
- Plain/ASCII fallback needs a conservative capability model beyond the current terminal/non-terminal split; this is now captured as an explicit helper and output-contract requirement.
- Docs and tests are tightly coupled to wizard strings, help text, and generated getting-started docs, so implementation must update code, tests, and managed docs together.
- Codex access can remain behaviorally stable because `internal/namba/codex_access.go` already isolates preset mapping and preview rendering.
- Readiness metadata should be refreshed after review files are updated so the stale frontend-major/imagegen gate is replaced by the CLI UX passthrough classification.

## Decisions

- Preserve `resolveInitProfileWithScan`, non-interactive skip behavior, and existing init flags.
- Split implementation into state-machine reorder, terminal visual/fallback helpers, localized copy, and docs/tests.
- Use TDD for navigation and fallback behavior before broad string assertions are updated.
- Keep Codex access preset IDs and `approval_policy` / `sandbox_mode` pairs unchanged.

## Follow-ups

- Add helper-level tests for terminal visual markers and plain output before rewriting the full wizard transcript tests.
- Add navigation tests for language-first ordering, `b`/`back`, manual Git skipping, GitHub skipping GitLab URL, and GitLab URL inclusion.
- Run `go test ./...`, `go vet ./...`, and `scripts/quality.sh` after implementation.

## Recommendation

- Cleared for implementation.
