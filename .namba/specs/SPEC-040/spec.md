# SPEC-040

## Goal

Evolve the existing NambaAI skill surface by auditing `ComposioHQ/awesome-codex-skills` for reusable Codex-skill patterns and adapting only the useful ideas into current Namba-managed skills, templates, and workflow guidance.

This SPEC is a reviewable planning package only. It must not install third-party skills, add new standalone skills, introduce Composio CLI or OAuth dependencies, automate Slack/Notion/app workflows, or create any `$CODEX_HOME/skills` install path.

## Command Choice Rationale

- The user explicitly requested `$namba-harness`, so this package was scaffolded through the harness planning command.
- The implementation will touch Namba-managed skill templates and generated repo-local skill surfaces, so the typed harness metadata classifies it as `core_harness_change` even though the scaffold originated from `$namba-harness`.
- Execution must use core-owned change discipline: template changes first, `namba regen` second, generated artifacts reviewed afterward.
- The outcome is not a new skill artifact. It is a controlled evolution of existing Namba skills and their safety/evaluation criteria.

## Verified External Audit

The audit source is `ComposioHQ/awesome-codex-skills` on GitHub, reviewed on 2026-04-27. The reusable patterns are:

- `skill-creator`: precise trigger metadata, lean `SKILL.md` bodies, progressive disclosure through `scripts/`, `references/`, and `assets/`, and a useful rule for matching freedom level to task fragility.
- `gh-address-comments`: thread-aware PR review handling through GitHub state instead of flat comments only.
- `gh-fix-ci`: PR check inspection, GitHub Actions log retrieval, failure snippet extraction, and explicit external-check scoping.
- `webapp-testing`: helper scripts treated as black boxes, `--help` first, managed server lifecycle, rendered-DOM inspection after load, screenshots, and browser-console evidence.
- `mcp-builder`: workflow-first tool design, concise outputs, actionable errors, bounded context, and evaluation questions that are independent, read-only, realistic, verifiable, and stable.

Rejected patterns:

- Skill installation into `$CODEX_HOME/skills`.
- Composio connect/connect-apps, CLI/OAuth setup, or app automation.
- Slack, Notion, email, meeting, lead, invoice, or similar third-party workflow automation.
- Copying upstream skills wholesale or creating new standalone Namba skills.

## Local Baseline

- Namba command-entry skills currently live under `.agents/skills/*/SKILL.md` and are generated from `internal/namba/templates.go`.
- Current generated skills are generally concise. The main risk is not current file length, but future guidance accretion if CI, review-thread, frontend, and MCP evaluation details are pasted directly into command-entry skill bodies.
- `$namba-review-resolve` already requires thread-aware review state and outcome labels, but it does not yet require CI-log evidence when review feedback or PR health depends on failing checks.
- `namba pr` already runs sync and validation and ensures the Codex review marker, but it does not yet require PR check inspection or failure-log evidence before handoff.
- `namba harness` already creates a reviewable harness SPEC package, but its quality criteria do not yet encode the stronger MCP/evaluation guidance found in the audited patterns.
- The repo already has frontend gate concepts in `namba-run` and `namba-workflow-execution`; this SPEC should strengthen validation guidance without inventing a separate frontend workflow.

## Problem

The Namba skill surface is moving toward richer review, CI, frontend, and harness evaluation behavior. If that guidance is added as loose prose, Namba risks:

- duplicated trigger wording across command-entry skills
- bloated `SKILL.md` bodies that make Codex load too much context too early
- review-resolution loops that resolve comments without enough CI or thread evidence
- PR handoffs that request review before failing GitHub Actions logs are inspected
- frontend validation advice that is not deterministic enough for repeated use
- harness/MCP evaluation criteria that are too subjective to test
- large skill-template migrations that are hard to review because mechanical and behavioral changes are mixed together

## Desired Outcome

- Existing Namba skills gain sharper trigger metadata and safer execution guidance while staying lean.
- Detailed recipes move behind progressive-disclosure references or deterministic helper-script candidates when they would otherwise overload a command-entry skill.
- `$namba-review-resolve` and `namba pr` require thread-aware review evidence plus CI-log inspection where GitHub checks are failing or review feedback depends on check status.
- Large managed-skill changes follow a batch-migration discipline: inventory, classify mechanical versus behavioral edits, update templates, regenerate, diff generated outputs, and validate.
- Harness and MCP quality criteria become concrete enough to drive review and tests.
- Frontend validation guidance uses reproducible server, browser, screenshot, DOM, and console evidence rather than visual intuition alone.

## Scope

- Update Namba-owned skill templates and generated skill surfaces only where useful.
- Refine `name` and `description` metadata when trigger boundaries are ambiguous or missing command synonyms.
- Add progressive-disclosure guidance for existing Namba skills if a skill would otherwise absorb long procedural recipes.
- Identify deterministic helper-script candidates for PR review threads, CI checks/logs, and frontend validation; only implement helpers when the implementation plan can keep them small, tested, and dependency-light.
- Strengthen `$namba-review-resolve`, `namba pr`, `namba-harness`, and related Namba guidance with the adapted patterns.
- Add batch-migration guidance for large managed-skill refactors.
- Improve harness/MCP quality criteria around workflow-first design, context budget, actionable errors, and stable evaluation scenarios.

## Non-Goals

- Do not add new standalone skills under `.agents/skills/`.
- Do not copy or vendor upstream `ComposioHQ/awesome-codex-skills` content.
- Do not add third-party skill installation, skill-sharing, or `$CODEX_HOME/skills` flows.
- Do not introduce Composio CLI, Composio OAuth, Slack, Notion, email, meeting, lead, invoice, or other app-automation dependencies.
- Do not make `namba pr` or `$namba-review-resolve` resolve external CI providers beyond reporting their URLs and status.
- Do not weaken existing `namba plan`, `namba harness`, `$namba-create`, `namba run`, `namba sync`, `namba pr`, or `namba land` boundaries.

## Design Constraints

- Keep `.namba/` as source of truth.
- Template-level changes belong in `internal/namba/templates.go` with regression coverage in the matching tests.
- Generated `.agents/skills/*/SKILL.md` changes should come from `namba regen`, not one-off manual edits.
- Keep command-entry skills compact. If a behavior needs more than a short checklist, move it into a reference or helper candidate rather than expanding the main skill body.
- Helper scripts, if implemented, must expose `--help`, support deterministic fixture-based tests, avoid destructive behavior, and make network access explicit through the calling workflow.
- External audit findings are pattern inputs only; Namba policy and local code remain authoritative.
