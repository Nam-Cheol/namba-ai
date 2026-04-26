# Design Review

- Status: clear
- Last Reviewed: 2026-04-26
- Reviewer: Codex
- Command Skill: `$namba-plan-design-review`
- Recommended Role: `namba-designer`

## Focus

- Re-review the conversational UX contract after source artifact revisions.
- Verify canonical response order, essential clarification definition, correction-first wrong-command behavior, the todo-list example, and `$namba-help` vs `$namba-coach` boundary clarity.
- Confirm the coaching surface stays read-only, routing-oriented, and distinct from planning or implementation.

## Verification

- Canonical response order is now explicit and consistent across `spec.md`, `contract.md`, and `acceptance.md`: brief restatement, up to three essential clarification questions when required, one primary executable handoff, optional single alternative, and a short reason.
- "Essential clarification" is now defined consistently as information needed to choose the correct Namba workflow or make the handoff command usable, not information needed to fully specify implementation.
- Wrong-command handling is now correction-first. The spec and contract both state that `$namba-coach` should correct a clearly wrong command choice before recommending execution, and the wrong-command examples make that behavior concrete.
- The todo-list ambiguity example is present and correctly scoped. It asks only routing-relevant questions first, then hands off to `namba plan "<description>"` without turning coach into a mini-spec generator.
- The `$namba-help` vs `$namba-coach` boundary is clear in the revised source set: help explains NambaAI usage, command semantics, and doc locations; coach selects the next workflow handoff for the user's current goal.
- `plan.md` and `eval-plan.md` both preserve the same conversational UX expectations, which reduces drift risk during implementation and test authoring.

## Findings

No design or conversational UX findings in the targeted re-review scope.

## Decisions

- The revised artifacts now define the coaching interaction tightly enough to implement without inventing behavior at coding time.
- The boundary between explanatory help and situational coaching is sufficiently legible for generated docs and skill text.
- The current wording keeps `$namba-coach` terse, routing-oriented, and safely read-only.

## Follow-ups

- Preserve the exact response-order wording during implementation and generated doc updates; this is now part of the UX contract, not optional copy style.
- Keep `$namba-help` and `$namba-coach` descriptions intentionally asymmetric in generated surfaces so they do not collapse back into generic "guidance" language.

## Recommendation

- Approved for implementation from a design/conversational UX perspective.
- No further wording changes are required in this review scope before proceeding.
