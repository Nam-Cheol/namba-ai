# Design Review

- Status: advisory
- Last Reviewed: 2026-04-23
- Reviewer: namba-designer
- Command Skill: `$namba-plan-design-review`
- Recommended Role: `namba-designer`

## Focus

- Review the CLI/config interaction model for access selection clarity before implementation starts.
- Check wording hierarchy, operator comprehension, and whether access choices are presented as meaningful presets instead of raw low-level toggles.
- Avoid inventing visual polish scope for a terminal-first workflow.

## Review Checklist

- Art direction is clear and fits the task context.
- Palette temperature and undertone logic are coherent, saturation stays restrained, and the result does not collapse into washed-out gray minimalism.
- Semantic components and layout primitives match the content instead of defaulting to generic cards, border-heavy framing, or bento/grid fallback.
- Motion, if proposed, has a concrete hierarchy, attention, or state-change purpose.
- The most generic section is redesigned when the task is page-, screen-, or section-scale; component-scale tasks call out the risk without gratuitous scope creep.
- Anti-overcorrection guardrails hold: no novelty for novelty's sake, no decorative asymmetry without payoff, and no loss of accessibility, design-system fit, or implementation realism.

## Findings

- The main design risk is ambiguous choice framing. The spec correctly identifies that `approval_policy` and `sandbox_mode` are currently exposed as two raw controls, but it should go further and require the new flow to lead with operator-intent presets first and implementation values second. If the wizard still visually or verbally centers the raw pair, users will continue to assemble meaning themselves.
- Access options need a stronger hierarchy than "pick a mode, then see two config values." For a CLI onboarding step, the operator should encounter: 1) what level of autonomy they want, 2) what that means in practice, and only then 3) the exact resulting `approval_policy` and `sandbox_mode`. Reversing that order will feel technically accurate but cognitively backwards.
- The wording must distinguish safety, convenience, and repository impact without collapsing into generic labels like "standard," "custom," or "recommended" on their own. Those labels are too soft for a permission-setting surface. Each choice needs a plain-language consequence statement such as whether commands may run without prompts, whether filesystem access is constrained, and when manual approval will interrupt flow.
- `namba codex access` needs an inspectable current-state presentation, not only a mutation path. Operators changing access after init need to answer "what is configured now?" before "what do I want to change?" The design quality here is comprehension and confidence, not visual treatment. If the command only accepts flags or immediately opens a selection flow without first summarizing the active setting, it will feel opaque.
- The spec mentions a concise preview and session-refresh notice, which is directionally correct, but the copy boundary matters. Preview text should describe the effective behavior before apply; refresh text should describe the operational next step after apply. Mixing those messages into one generic confirmation block will blur decision-making and reduce trust.

## Decisions

- Treat this as a CLI interaction design review, not a visual design exercise.
- Keep the design track advisory and allow implementation to proceed, provided the command and wizard center intent-based access presets and explicit behavior summaries.
- Favor a small number of clearly named access presets plus an explicit advanced/manual path only if needed; do not present every valid policy combination as if all combinations are equally interpretable for first-run users.

## Follow-ups

- During implementation review, verify that the wizard presents access as a single guided decision with subordinate technical detail, not as two adjacent low-level selectors with improved labels.
- Require help text and command output examples to show the mapping between preset language and effective config values so operators can build a stable mental model across init, post-init edits, and docs.
- Check that `namba codex access` can communicate current state, proposed state, and required refresh action as three distinct moments in the flow.
- If an advanced/manual combination path is introduced, ensure it is clearly marked as an expert override and does not visually or verbally compete with the default guided path.

## Recommendation

- Recommendation: proceed with advisory revisions in mind. The spec is sound, but implementation should be judged on whether access-setting choices read as clear operator intentions with explicit consequences, rather than as a generic presentation of underlying config knobs.
