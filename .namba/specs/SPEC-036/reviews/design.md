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
- The added design-judgment structure closes a remaining gap: `Asset Evidence`, `Direction Alternatives`, and `Design Review Axes` force reviewers to connect visual choices to available brand/product material, rejected options, and repeatable review criteria.
- The remaining design risk is bureaucratic drift if reviewers treat headings as a checklist instead of asking whether the evidence actually reduces execution ambiguity.

## Decisions

- Approve the direction to make `frontend-brief.md` the machine-readable frontend evidence artifact for new frontend-touching work.
- Approve the shift from “user-provided references exist” to “reference synthesis is complete.”
- Approve selective blocking only for explicitly classified `frontend-major` work, with invalid-contract surfacing when the canonical artifact contradicts itself.
- Treat the design review as an evidence-quality gate, not a taste-approval ceremony.
- Approve the Huashu-inspired structure in concept: design judgment must be fixed through evidence, alternatives, assets, and review axes, without importing that repository's export tooling or fixed style taxonomy.

## Approved Direction

- Force real synthesis before frontend implementation starts.
- Make generic fallback patterns easy to call out with concrete language.
- Preserve implementation realism by allowing multiple valid visual directions.
- Block only when execution ambiguity remains materially high.
- Persist enough asset and alternative evidence that frontend architecture can plan from a selected direction rather than from a generic layout default.

## Design Judgment Structure

- Evidence: references must include adopt/avoid/why and synthesis quality, not only links.
- Assets: brand assets, product/domain imagery, existing UI screenshots, constraints, and gaps are recorded separately from external references.
- Alternatives: at least three directions with tradeoffs are compared before the selected direction is approved.
- Review axes: evidence fit, asset fidelity, alternative coverage, visual hierarchy, craft, functionality/accessibility, and differentiation are checked explicitly.
- Output discipline: this structure narrows ambiguity; it must not become a style-copy mandate or a Figma/export-platform requirement.

## Banned Patterns

- Generic “cards everywhere” layouts without concept-level justification
- Card-inside-card containment as default structure
- Ornamental gradients, glass, or layered treatments without semantic purpose
- Typography systems with too many token steps or weak heading separation
- Section-level aesthetic switching that breaks product coherence
- Evidence sections filled with links or adjectives but no adopt/avoid reasoning
- Design-gate language that can only be enforced through reviewer taste
- Approved directions with no recorded assets, no rejected alternatives, or no repeatable review axes

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
