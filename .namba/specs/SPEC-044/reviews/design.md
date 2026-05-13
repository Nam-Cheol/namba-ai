# Design Review

- Status: clear
- Last Reviewed: 2026-05-13
- Reviewer: Codex
- Command Skill: `$namba-plan-design-review`
- Recommended Role: `namba-designer`

## Focus

- Clarify art direction, palette discipline, anti-generic composition, purposeful motion, and visual risks before implementation starts.

- Evidence Status: complete for workflow-contract scope; this SPEC is `frontend-minor`, but the negative-first gate design is specified in enough detail to implement and review.
- Gate Decision: advisory clear
- Approved Direction: Treat this as a contract-design review, not a UI moodboard. The approved direction is a deterministic negative-first review surface that names default bans, forces context-specific replacements, records the visual grammar and generic-section proof, and hands concrete constraints to architecture and implementation.
- Banned Patterns: generic "avoid bland UI" prose without detection hints; blanket bans with no allowed replacement; design review approval based only on five-gate completeness; card/grid/gradient bans phrased as taste law instead of context; fake page-scale redesign requirements for non-page work; architecture handoff that leaves component families or scope boundaries implicit.
- Open Questions: Whether the generated `reviews/design.md` scaffold will introduce fixed review labels for each contract axis or keep them in narrative form only.
- Unresolved Questions: None blocking the SPEC direction.
- Design Review Axes: evidence, assets, alternatives, hierarchy, craft, functionality, differentiation
- Keep / Fix / Quick Wins: keep the contract stable and parser-friendly; fix any scaffold language that sounds like universal style bans; add explicit reviewer-visible cues for violation-check evidence.

## Review Checklist

- Art direction is clear and fits the task context.
- Palette temperature and undertone logic are coherent, saturation stays restrained, and the result does not collapse into washed-out gray minimalism.
- Semantic components and layout primitives match the content instead of defaulting to generic cards, border-heavy framing, or bento/grid fallback.
- Motion, if proposed, has a concrete hierarchy, attention, or state-change purpose.
- The most generic section is redesigned when the task is page-, screen-, or section-scale; component-scale tasks call out the risk without gratuitous scope creep.
- Anti-overcorrection guardrails hold: no novelty for novelty's sake, no decorative asymmetry without payoff, and no loss of accessibility, design-system fit, or implementation realism.

## Findings

- The SPEC is strong on negative-first contract completeness. `spec.md`, `contract.md`, and `acceptance.md` consistently require contract status, default anti-pattern library, context-specific bans, allowed replacements, brand/category/trust reasoning, visual grammar, generic-section proof, architecture handoff, and post-implementation checks.
- The default anti-pattern library is appropriately concrete. It bans recognizable generic fallbacks such as card walls, bento-as-identity, decorative glass/orbs, stock SaaS heroes, KPI-card misuse, weak empty states, one-note palettes, and low-fidelity stock imagery without treating the underlying primitives as globally forbidden.
- The replacement model is correct. Requiring rationale, detection hints, allowed replacement, and exception path prevents the common failure mode where review language becomes subjective taste policing with no implementation path.
- The library-fit rule is well framed for Namba. The SPEC makes cards, borders, grids, gradients, and metrics valid when semantically justified, which preserves design-system realism and avoids overcorrection into gray anti-style minimalism.
- The visual grammar requirement is strong enough for architecture handoff. The grammar explicitly covers layout primitives, hierarchy, type, color roles, density, depth, imagery/icons, and motion, which gives `namba-frontend-architect` something actionable beyond approval prose.
- The most-generic-section proof is one of the highest-value additions. It forces the review to identify where an agent would most likely collapse into template output and to record the better replacement before implementation starts.
- The architecture handoff and violation-check plan are correctly treated as design outputs, not implementation extras. That is necessary for blocking approved-looking but generic UI work after coding.
- Scope discipline is intact. Because this SPEC is harness work, there is no need to invent palette or motion direction for a real product surface; the design value here is the quality and enforceability of the review contract itself.

## Decisions

- Approve the negative-first direction as the default design-review contract for future `frontend-major` SPECs.
- Preserve the default anti-pattern library as a starting library, not a universal house style ban list.
- Require context-specific replacements and exception paths in review output; a ban without an allowed path should continue to count as insufficient evidence.
- Keep generic-section redesign proof mandatory for page-, screen-, and section-scale work, with a clear not-applicable path for narrower scopes.
- Treat frontend architecture handoff and post-implementation violation checks as first-class design-review deliverables.

## Follow-ups

- Ensure the generated `reviews/design.md` scaffold explicitly prompts reviewers for default-library fit, context-specific bans/replacements, visual grammar, generic-section proof, architecture handoff, and violation-check plan.
- During implementation, keep wording concrete enough that readiness and run-time messages can point to missing or contradictory contract fields without interpretation drift.
- When execution/output artifacts are updated, make the violation-check evidence visible enough that a reviewer can tell which banned pattern was inspected and why it passed or failed.

## Recommendation

- Proceed. The design direction is clear and implementation-ready for workflow scope, with the main requirement being that the scaffold and validation text stay concrete rather than collapsing back into generic review language.
