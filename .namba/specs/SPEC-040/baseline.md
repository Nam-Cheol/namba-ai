# SPEC-040 Baseline

## External Audit Sources

- `ComposioHQ/awesome-codex-skills` README: skill anatomy, trigger metadata, progressive disclosure through `scripts/`, `references/`, and `assets/`, and the upstream install flows that this SPEC rejects.
- `skill-creator/SKILL.md`: concise skill bodies, matching specificity to task fragility, and bundled resources for deterministic or long-context work.
- `gh-address-comments/SKILL.md` and `scripts/fetch_comments.py`: thread-aware review discovery through GitHub GraphQL, including resolved state.
- `gh-fix-ci/SKILL.md` and `scripts/inspect_pr_checks.py`: PR check inspection, GitHub Actions log retrieval, failure snippet extraction, pending-log handling, and external-check scoping.
- `webapp-testing/SKILL.md`: black-box helper scripts, `--help` before reading script source, server lifecycle management, rendered-state inspection, screenshots, and console logs.
- `mcp-builder/SKILL.md`: workflow-first MCP tool design, context-budgeted output, actionable errors, and evaluation-driven quality criteria.

## Local Skill Surface

- `.agents/skills/namba-review-resolve/SKILL.md`
  - Already requires unresolved thread state, meaningful/actionable classification, per-thread outcomes, original-thread replies, validation before resolve, and no duplicate review marker.
  - Gap: does not explicitly require PR check or CI-log evidence when review comments relate to failing checks or when handoff depends on check state.

- `.agents/skills/namba-pr/SKILL.md`
  - Already requires configured base/language/marker, `namba sync`, validation, review readiness in PR summary, commit, push, PR creation/reuse, and marker de-duplication.
  - Gap: does not require check-status inspection or bounded GitHub Actions failure snippets before review handoff.

- `.agents/skills/namba-harness/SKILL.md`
  - Already creates a reviewable harness SPEC under `.namba/specs/<SPEC>` with review artifacts.
  - Gap: does not yet encode richer harness/MCP quality criteria or helper-candidate evaluation criteria.

- `.agents/skills/namba-run/SKILL.md` and `.agents/skills/namba-workflow-execution/SKILL.md`
  - Already distinguish frontend-major and frontend-minor work through frontend briefs and design gate evidence.
  - Gap: can better define deterministic frontend validation evidence when implementation reaches a browser-rendered surface.

- `.agents/skills/namba-create/SKILL.md`
  - Already rejects `.codex/skills`, requires preview-first creation, normalizes paths, and keeps user-authored outputs distinct from Namba-managed built-ins.
  - Gap: can borrow progressive-disclosure wording without reintroducing install flows or new built-in skills.

- `internal/namba/templates.go`
  - Owns the managed skill templates. Durable implementation belongs here.

- `internal/namba/templates_test.go` and `internal/namba/update_command_test.go`
  - Already assert important generated skill strings and managed-skill registry behavior. These are the right regression-test entrypoints for template changes.

## Current Risk

- The existing generated skills are short, so this is not a cleanup of dense files today.
- The risk is that adding review-thread, CI-log, frontend, and MCP evaluation details directly to command-entry skills would make the surface dense and repetitive.
- The implementation should add contract-level guidance and helper-candidate criteria now, then only add reference/helper artifacts where the actual template diff would otherwise become too large.

## Source Links

- Repository: https://github.com/ComposioHQ/awesome-codex-skills
- `gh-address-comments`: https://github.com/ComposioHQ/awesome-codex-skills/tree/master/gh-address-comments
- `gh-fix-ci`: https://github.com/ComposioHQ/awesome-codex-skills/tree/master/gh-fix-ci
- `webapp-testing`: https://github.com/ComposioHQ/awesome-codex-skills/tree/master/webapp-testing
- `mcp-builder`: https://github.com/ComposioHQ/awesome-codex-skills/tree/master/mcp-builder
- `skill-creator`: https://github.com/ComposioHQ/awesome-codex-skills/tree/master/skill-creator
