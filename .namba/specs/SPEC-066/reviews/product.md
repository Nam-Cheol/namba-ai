# Product Review

- Status: cleared
- Last Reviewed: 2026-05-22
- Reviewer: Codex product review
- Command Skill: `$namba-plan-pm-review`
- Recommended Role: `namba-product-manager`

## Focus

- Challenge the problem framing, scope, user value, and acceptance bar before implementation starts.

## Findings

- Product value is clear: future `namba init` users should receive stronger generated skill and custom-agent instruction contracts without requiring downstream repo backfills.
- The target surfaces are explicit: `.agents/skills/*`, `.codex/agents/*.md`, `.codex/agents/*.toml`, and `.namba/manifest.json` ownership boundaries.
- Primary evidence is correctly tied to fresh temporary repo `namba init` and existing repo `namba regen` parity. Current-repo generated diffs are secondary evidence only.
- Non-goals are explicit: no downstream backfills, no manual generated-file patch as durable fix, no network-dependent checks, no LLM-judge evaluations, and no unrelated workflow restructuring.

## Decisions

- Treat this SPEC as init/regen generated instruction surface quality contract work.
- Prioritize command-entry skills and project-scoped custom agents first while preserving a common contract across all Namba-managed generated surfaces.
- Product clearing depends on deterministic contract evidence, not subjective wording polish.

## Follow-ups

- Implementation should keep generated guidance non-project-specific and suitable for future repositories.
- Final PR notes should separate generator-source changes from secondary `namba regen` output diffs.

## Recommendation

- Cleared for implementation from a product perspective.
