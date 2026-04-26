# Design Review

- Status: clear
- Last Reviewed: 2026-04-25
- Reviewer: namba-designer
- Command Skill: `$namba-plan-design-review`
- Recommended Role: `namba-designer`

## Focus

- Clarify art direction, palette discipline, anti-generic composition, purposeful motion, and visual risks before implementation starts.

## Review Checklist

- Art direction is clear and fits the task context.
- Palette temperature and undertone logic are coherent, saturation stays restrained, and the result does not collapse into washed-out gray minimalism.
- Semantic components and layout primitives match the content instead of defaulting to generic cards, border-heavy framing, or bento/grid fallback.
- Motion, if proposed, has a concrete hierarchy, attention, or state-change purpose.
- The most generic section is redesigned when the task is page-, screen-, or section-scale; component-scale tasks call out the risk without gratuitous scope creep.
- Anti-overcorrection guardrails hold: no novelty for novelty's sake, no decorative asymmetry without payoff, and no loss of accessibility, design-system fit, or implementation realism.

## Findings

- This SPEC has no end-user UI. The relevant design surface is the operator
  contract in `.namba/hooks.toml`, hook evidence records, artifact paths, and
  failure/status language exposed in manifests and docs.
- Event names are mostly readable for operators. `before_preflight`,
  `after_preflight`, `before_execution`, `after_execution`,
  `before_validation`, `after_validation`, and `on_failure` map cleanly to the
  run lifecycle. The tool-boundary names `after_patch`, `after_bash`, and
  `after_mcp_tool` also read as observations rather than control points, which
  is the right information design for this SPEC.
- The revised spec now fixes the main readability gap with an explicit
  `Operator Outcome Glossary`. It separates hook-process `status`,
  pipeline `stage_status`, terminal `failure_status`, stopping semantics via
  `blocking`, and post-failure behavior via `failure_action`. That removes the
  prior ambiguity between process result, phase result, and run outcome.
- The evidence contract is now readable in a stable operator sequence. The spec
  defines the manifest reading order as
  `event -> hook_name -> status -> exit_code -> blocking/failure_action -> error_summary -> stdout_path/stderr_path`,
  which is the right information hierarchy for fast diagnosis.
- `blocking`, `failure_action`, and `error_summary` are no longer underspecified
  advisory ideas. The spec and acceptance criteria now treat them as expected
  v1 operator fields, with concrete value sets and explicit config-error
  expectations such as `status: "error"`, `exit_code: -1`,
  `blocking: true`, `failure_action: "stopped"`, and a concrete
  `error_summary`.
- The manifest example is materially better. Showing one successful advisory
  hook, one failed-but-continued hook, and one blocking failure makes the
  contract readable without cross-referencing multiple sections. It also proves
  that the glossary maps cleanly onto real evidence rows.
- Artifact path design is readable. Grouping outputs under
  `.namba/logs/runs/<log-id>-hooks/<event>/` is operator-friendly because it
  makes the event model visible in the filesystem. The example path naming is
  easier to scan than a flat hook artifact directory.
- Config naming is good. Using `[hooks.<hook_name>]` and recording `hook_name`
  in evidence creates a strong operator mental model between manifest row,
  config key, and artifact filenames.
- The prior wording ambiguity around operator fields has been resolved. The
  spec now says new v1 hook-result records must include the operator fields,
  while older evidence manifests remain compatible without them.

## Decisions

- Do not treat this as a visual design review. Treat it as an operator-facing
  information architecture review for naming, evidence schema readability, and
  docs clarity.
- Keep the current event names. They are specific enough for implementation and
  do not need a terminology rewrite.
- The revised operator glossary, reading order, and manifest example are
  sufficient to mark the evidence contract readable for implementation.

## Follow-ups

- Keep the manifest example and glossary aligned during implementation. If the
  evidence writer drops any of `blocking`, `failure_action`, or
  `error_summary`, readability will regress immediately.
- No design-blocking follow-ups remain.

## Recommendation

- Recommendation: clear
- The revised SPEC is ready from a design and operator-information-architecture
  standpoint. The outcome glossary, evidence reading order, blocking/failure
  fields, and manifest example now make the contract readable without guesswork.
