# Design Review

- Status: approved
- Last Reviewed: 2026-04-23
- Reviewer: namba-designer
- Command Skill: `$namba-plan-design-review`
- Recommended Role: `namba-designer`

## Focus

- Clarify art direction, evidence quality, anti-generic composition, and whether the frontend gate stays rigorous without collapsing into taste policing.

## Evidence Status

- sufficient
- Why:
  - `spec.md` frames the regression target as generic fallback behavior and weak evidence quality, not personal preference.
  - The five-gate contract is reviewable because it names concrete evidence categories, distinguishes `missing` from `insufficient`, and now defines explicit invalid-contract states.
  - `frontend-brief.md` is canonical enough to keep implementation grounded while still allowing multiple valid visual directions.
  - `acceptance.md` checks inspectable artifacts and explicit run-block conditions rather than vague promises of “better design.”

## Findings

- The strongest part of the SPEC is that it targets generic fallback behavior instead of prescribing a single house style.
- The anti-generic rules are credible because they reject misuse patterns such as default card walls, unjustified depth, and section-by-section visual drift rather than banning common primitives outright.
- The review standard is concrete enough to constrain implementation: adopt/avoid/why, hierarchy intent, spacing-density logic, depth budget, prototype evidence, and banned patterns are all explicit.
- The added classification rationale and blocked-run remediation requirements improve operator clarity without weakening design rigor.
- The remaining design risk is bureaucratic drift if reviewers treat headings as a checklist instead of asking whether the evidence actually reduces execution ambiguity.

## Decisions

- Approve the direction to make `frontend-brief.md` the machine-readable frontend evidence artifact for new frontend-touching work.
- Approve the shift from “user-provided references exist” to “reference synthesis is complete.”
- Approve selective blocking only for explicitly classified `frontend-major` work, with invalid-contract surfacing when the canonical artifact contradicts itself.
- Treat the design review as an evidence-quality gate, not a taste-approval ceremony.

## Approved Direction

- Force real synthesis before frontend implementation starts.
- Make generic fallback patterns easy to call out with concrete language.
- Preserve implementation realism by allowing multiple valid visual directions.
- Block only when execution ambiguity remains materially high.

## Banned Patterns

- Generic “cards everywhere” layouts without concept-level justification
- Card-inside-card containment as default structure
- Ornamental gradients, glass, or layered treatments without semantic purpose
- Typography systems with too many token steps or weak heading separation
- Section-level aesthetic switching that breaks product coherence
- Evidence sections filled with links or adjectives but no adopt/avoid reasoning
- Design-gate language that can only be enforced through reviewer taste

## Open Questions

- How should planners describe “most generic section” redesign guidance when the scope is a full page rather than one module?
- Should readiness render insufficient frontend gates in a fixed order for faster scanning?

## Unresolved Questions

- None that block implementation of the planning contract.

## Follow-ups

- Keep reviewer guidance explicit that card, border, and grid use are valid when semantically justified.
- Ensure generated templates preserve the difference between `missing`, `insufficient`, and invalid-contract states.
- Verify that blocked `frontend-major` runs route back to synthesis work with specific remediation prompts.

## Recommendation

- Proceed. The design gate is concrete, evidence-based, and strict where ambiguity causes generic UI output.
