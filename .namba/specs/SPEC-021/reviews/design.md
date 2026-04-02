# Design Review

- Status: approved
- Last Reviewed: 2026-04-02
- Reviewer: Codex
- Command Skill: `$namba-plan-design-review`
- Recommended Role: `namba-designer`

## Focus

- Clarify interaction quality, responsive states, accessibility, and visual direction before implementation starts.

## Findings

- The current docs are structurally accurate, but the scan path for a first-time user is still too list-heavy and does not help them choose the next command fast enough.
- Prioritizing information architecture and readability is the correct design target for this SPEC because the main UX problem is orientation, not visual polish.
- The root README should feel like a guided entry surface: what NambaAI is, which command to choose first, what the shortest end-to-end flow looks like, and where to go for depth.
- Command-skill explanations need to be intent-led. A first-time user should understand "when to use this" before reading command syntax or internal terminology.
- The workflow guide can carry denser operational detail, but its section order still needs to mirror the README's mental model so users do not feel like they switched to a different product.

## Decisions

- Treat README layout as a guided onboarding sequence, not a flat capability inventory.
- Keep the root README concise and highly scannable, with a prominent command-choice section near the top and short copyable examples.
- Reserve longer lifecycle detail, execution mode nuance, and review-readiness explanation for `docs/workflow-guide*.md`.
- Group commands and skills by user intent such as bootstrap, plan, execute, sync, and handoff rather than by implementation category.
- Preserve the same section rhythm and relative prominence across localized README and workflow guide variants.

## Follow-ups

- Add an explicit "which command should I use?" section early in the root README that distinguishes `namba project`, `namba plan`, `namba harness`, and `namba fix`.
- Expand the command-skill section so each skill explains the situation it is for, not just the matching CLI command.
- Keep quick start examples short enough to be copied line-by-line by a first-time user without cross-referencing deeper docs.
- Make the workflow guide visibly deeper than the README rather than duplicating the same bullets with slightly different wording.
- Use heading structure and short paragraphs to reduce visual fatigue from long uninterrupted bullet lists.

## Recommendation

- Advisory recommendation: approved. The SPEC has a sound design direction for onboarding, scanability, and command/skill discoverability. Proceed to implementation with the README as the primary first-session surface and the workflow guide as the detailed operational reference.
