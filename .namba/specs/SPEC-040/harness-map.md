# SPEC-040 Harness Map

## Core Relationship

This SPEC adapts external Codex-skill design patterns into the existing Namba-managed skill surface.

- External source: `ComposioHQ/awesome-codex-skills`
- Namba target: existing repo-local Namba skills under `.agents/skills/`, generated from `internal/namba/templates.go`
- Delivery model: reviewable SPEC first, then template-level implementation and `namba regen`
- Explicit boundary: no new standalone skills and no third-party install flow

## Adaptation Map

| Upstream pattern | Namba target | Adaptation |
| --- | --- | --- |
| `skill-creator` trigger metadata | all affected Namba command-entry skills | Tighten `description` text and trigger boundaries without broad catch-all language. |
| `skill-creator` progressive disclosure | `$namba-review-resolve`, `namba-pr`, `namba-harness`, frontend validation guidance | Keep `SKILL.md` concise; move long recipes into references or helper candidates only when needed. |
| `gh-address-comments` thread discovery | `$namba-review-resolve` | Reinforce unresolved review-thread state, thread ids, paths, per-thread outcomes, original-thread replies, and resolution discipline. |
| `gh-fix-ci` check/log inspection | `namba-pr`, `$namba-review-resolve` | Add PR check inspection, bounded GitHub Actions failure snippets, pending-log handling, and external-check URL-only reporting. |
| `webapp-testing` helper/server pattern | `namba-run`, `namba-workflow-execution`, frontend validation references | Use server lifecycle, rendered DOM, screenshot, and console evidence for frontend validation. |
| `mcp-builder` workflow/eval quality | `namba-harness`, harness SPEC guidance, MCP quality criteria | Add workflow-first design, context-budgeted output, actionable errors, and stable read-only evaluation criteria. |

## Rejection Map

| Upstream area | Namba decision |
| --- | --- |
| `skill-installer` and manual installation | Reject. Namba uses repo-local `.agents/skills/` and generated managed surfaces. |
| `$CODEX_HOME/skills` | Reject. Do not create install flow or duplicate skill discovery. |
| `connect` and `connect-apps` | Reject. Composio CLI/OAuth dependencies are out of scope. |
| Slack/Notion/app automation skills | Reject. Third-party app workflows are outside the requested Namba skill-surface evolution. |
| Whole-skill copying | Reject. Only pattern-level adaptation is allowed. |

## Planned Namba Surfaces

- `internal/namba/templates.go`
- `internal/namba/templates_test.go`
- `internal/namba/update_command_test.go`
- `.agents/skills/namba-review-resolve/SKILL.md`
- `.agents/skills/namba-pr/SKILL.md`
- `.agents/skills/namba-harness/SKILL.md`
- `.agents/skills/namba-run/SKILL.md`
- `.agents/skills/namba-workflow-execution/SKILL.md`
- `.agents/skills/namba-create/SKILL.md`
- `.agents/skills/namba/SKILL.md`
- `.namba/codex/README.md` or generated README/workflow guide sections only if the implementation changes operator-facing docs

## Helper Candidate Map

| Candidate | Purpose | Minimum proof |
| --- | --- | --- |
| PR review thread fetcher | Emit unresolved review thread state as stable JSON for `$namba-review-resolve`. | Fixture tests for resolved/unresolved/outdated threads and pagination shape. |
| PR CI inspector | Emit failing PR check summaries and bounded GitHub Actions failure snippets. | Fixture tests for successful, failing, pending, external, and missing-log cases. |
| Frontend validation runner | Manage local server lifecycle and collect Playwright evidence. | Local dummy-server test with screenshot path, DOM assertion, and console-error capture. |

Helper candidates are optional implementation outputs. They should not be added unless they stay deterministic, small, tested, and clearly useful beyond prose guidance.
