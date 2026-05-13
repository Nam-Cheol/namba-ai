# SPEC-044

## Goal

Embed a negative-first design reasoning system into NambaAI `frontend-major` workflows so AI-generated UI work cannot pass the frontend synthesis gate by merely filling positive evidence labels while still relying on banned generic UI fallbacks.

## Context

- Project: namba-ai
- Project type: existing
- Language: go
- Mode: tdd
- Work type: harness planning
- Planning surface: `namba harness "<description>"`
- Baseline: `SPEC-036` introduced `frontend-brief.md`, the five-gate frontend synthesis contract, design-review summaries, readiness rendering, and `namba run` blocking for incomplete or contradictory `frontend-major` evidence.
- Current code surfaces to inspect during implementation:
  - `internal/namba/frontend_brief.go`: classification, frontend-brief parsing, readiness lines, execution blocking, remediation text.
  - `internal/namba/spec_review.go`: design-review scaffold and readiness aggregation.
  - `internal/namba/templates.go`: generated command skills, custom agents, AGENTS guidance, README/workflow content.
  - `internal/namba/execution_test.go`, `internal/namba/frontend_brief_test.go`, `internal/namba/spec_command_test.go`, `internal/namba/spec_review_test.go`, `internal/namba/templates_test.go`: regression coverage for frontend gate behavior.

## Problem

The existing frontend synthesis gate is evidence-first but not negative-first. It requires a problem frame, references, critique, a decision, and prototype evidence, and it blocks obvious missing or contradictory states. That still leaves a path for weak AI output:

1. The brief can mark the five gates `complete`.
2. The design review can mark `Evidence Status: complete` and `Gate Decision: approved`.
3. The implementation can still default to generic UI patterns such as card walls, bento grids, glassy gradient heroes, fake SaaS metrics rows, stock testimonial strips, or decorative containers that are not justified by the brand, category, trust model, or state model.

This is a workflow failure, not a taste preference. The gate needs a required "do not design this way" contract that forces the system to name the obvious generic fallback, ban it when inappropriate, provide allowed replacements, and verify after implementation that the output did not quietly regress to the banned fallback.

## Desired Outcome

- `frontend-major` work includes a required Do-Not Design Contract in `frontend-brief.md`.
- The contract starts from a default anti-pattern library and then narrows it with context-specific banned patterns for the SPEC.
- Each banned pattern has an allowed replacement pattern or an explicit exception path, so the system blocks generic output without forcing novelty for novelty's sake.
- The contract records brand, category, and trust reasoning before visual decisions are approved.
- The contract records a visual grammar for layout primitives, type scale, color/palette role, density, depth, imagery/iconography, and motion.
- Page-, screen-, and section-scale work includes a most-generic-section redesign proof that names the weakest/generic section and explains the replacement.
- `namba-frontend-architect` receives a concrete handoff from the design gate before planning component boundaries, state ownership, or file structure.
- `namba-frontend-implementer` and `namba run` perform post-implementation violation checks against the contract.
- `frontend-major` execution is blocked when the brief, design review, architecture handoff, implementation proof, or post-implementation check shows reliance on banned generic UI fallbacks, even if the previous reference and gate labels are marked complete.

## Negative-First Contract

Extend `frontend-brief.md` for `frontend-major` work with a required Do-Not Design Contract. The implementation may choose exact parser structure, but the generated artifact must keep the fields stable enough for deterministic readiness and run-time checks.

Required content:

- Contract status: `complete`, `missing`, or `insufficient`.
- Default anti-pattern library: the baseline banned-pattern set that applies unless a SPEC explicitly explains why an item is irrelevant.
- Context-specific banned patterns: the patterns this SPEC must not use, with rationale and detection hints.
- Allowed replacement patterns: acceptable layout, content, hierarchy, state, and interaction patterns that implementation may use instead.
- Brand/category/trust reasoning: how the product category, user trust burden, brand assets, and audience expectations constrain visual choices.
- Visual grammar contract: the approved grammar for layout primitives, typography scale, palette roles, density, depth/elevation, imagery/iconography, and motion.
- Most-generic-section redesign proof: the section most likely to collapse into generic output, the rejected fallback, the replacement direction, and the evidence that makes the replacement better.
- Frontend architecture handoff: what the frontend architect is allowed to plan, what must stay out of scope, which primitives/components are encouraged, and which are banned.
- Post-implementation violation checks: the checks the implementer must run after coding, including which banned patterns to inspect, what file or screenshot evidence to cite, and what result blocks the run.

## Default Anti-Pattern Library

The v1 default library should be concrete enough to guide agents and stable enough for tests. It should include, at minimum:

- Generic card grids as the primary page grammar when the content is homogeneous or workflow-driven.
- Bento grids used as an identity substitute rather than as a true information architecture.
- Card-inside-card nesting, double backgrounds, and stacked border/shadow/tint/radius treatments without interaction semantics.
- Glassmorphism, blurred blobs, gradient-orb backgrounds, or decorative gradients that do not support the brand/category/trust model.
- Stock SaaS hero formulas: nav + oversized headline + vague CTA + fake metrics/logos + abstract dashboard preview.
- Feature-card walls with interchangeable icon/title/body cards.
- Testimonial, FAQ, pricing, or stats sections copied from generic landing-page templates without product-specific trust evidence.
- Dashboard KPI card rows when the primary task is comparison, triage, workflow progress, or dense operational scanning.
- Low-information empty states that use generic illustration/copy instead of next-step guidance.
- Typography fallback patterns such as tiny body text, too many text styles, weak heading steps, or light weights on small text.
- One-note palettes, especially unearned purple/blue gradients, washed-out gray minimalism, or decorative color that does not map to state, priority, or brand.
- Dark, blurred, cropped, stock-like, or purely atmospheric imagery where users need to inspect the real product, person, place, object, gameplay, or state.

The library should not ban primitives globally. A card, gradient, or dashboard metric can still be allowed when the contract explains the semantic reason, scope boundary, and replacement/exception rule.

## Context-Specific Bans And Replacements

For each `frontend-major` SPEC, the design synthesis must record:

- `Banned pattern`: a concrete pattern, not a vague adjective.
- `Why banned here`: the product/user reason it would be generic, misleading, low-trust, inaccessible, or hard to scan.
- `Detection hints`: words, component names, CSS classes, layout shapes, screenshot cues, or implementation signs that should trigger review.
- `Allowed replacement`: the pattern family implementation should use instead.
- `Exception path`: when the banned pattern is allowed, if ever, and what evidence must justify it.

Example:

- Banned pattern: "Three-column feature-card wall for core workflow explanation."
- Why banned here: "The user needs to compare state transitions, not browse interchangeable claims."
- Detection hints: "feature-card grid, repeated icon/title/body cards, `grid-cols-3` marketing modules."
- Allowed replacement: "A single workflow lane with state chips, decision points, and inline proof."
- Exception path: "Allowed only for truly independent feature categories with distinct actions."

## Brand, Category, And Trust Reasoning

The gate must require reasoning that connects visual choices to context:

- Brand: existing brand assets, typography, tone, product screenshots, and visual constraints.
- Category: whether the surface is SaaS, CRM, developer tool, marketplace, healthcare, finance, creative tool, game, portfolio, venue, or another category with specific visual expectations.
- Trust: what the user must believe before acting, which evidence builds that trust, and which generic patterns weaken it.
- Audience and workflow: scanning density, decision speed, accessibility burden, emotional tone, and repeated-use ergonomics.

This reasoning is required before the design review can approve the direction.

## Visual Grammar Contract

The approved direction must define a reusable visual grammar:

- Layout primitives: list, table, timeline, flow, command surface, inspector, canvas, split pane, cards, or other primitives, with justification.
- Hierarchy: how primary, secondary, and tertiary information are ordered.
- Typography: named scale and weight rules with a body-size floor.
- Color and palette roles: brand, surface, state, priority, danger, success, selected, disabled, and focus.
- Density and spacing: compactness, rhythm, and responsive behavior.
- Depth and containers: when elevation, borders, backgrounds, and radii are allowed.
- Imagery and icons: source, fidelity bar, and when an icon is functional versus decorative.
- Motion: allowed motion and the state/attention reason it exists.

Implementation should treat this grammar as the frontend architecture input, not as optional prose.

## Most-Generic-Section Redesign Proof

For page-, screen-, and section-scale `frontend-major` work, the synthesis must identify the section most likely to be generic and prove it was redesigned. The proof must include:

- The section name or role.
- The obvious generic fallback.
- Why that fallback would be weak in this context.
- The replacement structure.
- Evidence source: reference synthesis, product state model, brand/category/trust reasoning, or prototype evidence.
- Implementation implication: what the frontend architect and implementer must build or avoid.

Component-scale work may mark this proof as not applicable only when the brief explains why the task cannot reasonably contain a page/screen/section redesign.

## Frontend Architecture Handoff

`namba-frontend-architect` must not plan a `frontend-major` implementation from the five-gate status alone. The architecture handoff should include:

- Approved layout primitives and component families.
- State ownership and interaction boundaries that reinforce the visual grammar.
- Banned components or pattern families that must not be introduced.
- Required assets, screenshots, prototypes, or inspection evidence.
- Responsive and accessibility constraints.
- File/module planning notes where the visual grammar affects component boundaries.

If this handoff is missing, pending, or contradicted by design review, architecture and implementation stay blocked.

## Post-Implementation Violation Checks

After implementation and before a run is treated as successful, Namba must require violation checks against the Do-Not Design Contract. The check should be deterministic where possible and reviewer-visible where judgment is required.

Minimum behavior:

- The execution prompt must require a "Do-Not Design Violation Check" section in the implementation result for `frontend-major` work.
- The check must cite changed files and, when a local app is available, screenshot or browser evidence.
- If a banned pattern appears, the run result must block or fail with a violation message that names the banned pattern and remediation path.
- If the implementation uses an exception path, the result must cite the contract evidence that allows the exception.
- `namba sync` and PR support artifacts must surface any failed or unresolved violation check so a PR cannot look clean only because the original gate labels are approved.

## Public Interfaces And Contracts

- Expanded artifact: `.namba/specs/<SPEC>/frontend-brief.md`
  - Add a required Do-Not Design Contract section for `frontend-major`.
  - Keep fixed-label parsing deterministic; if new fixed labels are added, update allowed/required-label parsing and compatibility tests intentionally.
- Expanded artifact: `.namba/specs/<SPEC>/reviews/design.md`
  - Add review focus for the Do-Not Design Contract, default library fit, context-specific banned patterns, allowed replacements, visual grammar, generic-section proof, architecture handoff, and violation-check plan.
- Expanded readiness summary: `.namba/specs/<SPEC>/reviews/readiness.md`
  - Surface missing, insufficient, failed, or unresolved negative-first contract state separately from the existing five-gate evidence state.
- Expanded execution contract: `namba run SPEC-XXX`
  - Block before implementation when the Do-Not Design Contract is missing or insufficient for `frontend-major`.
  - Block after implementation when the violation check reports banned generic UI fallback reliance.
- Expanded role and skill contracts:
  - `namba-designer`: owns negative-first synthesis, default library application, context-specific bans, replacements, visual grammar, and generic-section proof.
  - `namba-frontend-architect`: consumes the architecture handoff and refuses to plan from generic primitives alone.
  - `namba-frontend-implementer`: implements within the grammar and reports violation-check evidence.
  - `$namba-run`, `$namba-plan-design-review`, `$namba-plan-review`, README/workflow docs, and generated agent mirrors must describe the negative-first gate.

## Implementation Scope

- Extend frontend-brief generation for `frontend-major` scaffolds.
- Extend frontend-brief parsing/reporting to track Do-Not Design Contract state and contract issues.
- Extend design-review scaffolds and readiness aggregation.
- Extend run-time blocking and remediation output.
- Extend execution prompt materialization and run-result handling so post-implementation violation checks are visible and blocking.
- Update templates and generated surfaces through managed template changes, then regenerate.
- Add regression tests before implementation changes, following the repo's TDD mode.

## Non-Goals

- Do not create a second artifact model outside `.namba/specs/`.
- Do not turn design review into a single house style or ban valid primitives globally.
- Do not require live Figma, browser, network, or third-party app access for the minimum gate.
- Do not require image recognition for v1; browser/screenshot evidence can strengthen checks when available, but the baseline must remain runnable in local CI.
- Do not hard-block non-frontend or `frontend-minor` work with the full negative-first contract.
- Do not emit Claude-only runtime primitives such as `.claude/*`, `TeamCreate`, `SendMessage`, `TaskCreate`, or a mandatory `model: "opus"` requirement.

## Validation Strategy

- Planning scaffold tests:
  - New `frontend-major` SPECs include the Do-Not Design Contract section and default anti-pattern library prompts.
  - `frontend-minor` and non-frontend SPECs do not inherit the full contract burden.
- Parser and readiness tests:
  - Missing, pending, or insufficient Do-Not Design Contract fields block `frontend-major`.
  - A fully complete five-gate header plus approved design review still blocks when the negative-first contract is missing.
  - Context-specific banned patterns and allowed replacements are rendered in readiness.
- Execution tests:
  - `namba run` blocks before runner dispatch when the negative-first contract is incomplete.
  - `namba run` fails after runner output when the violation check reports a banned pattern.
  - Exception-path usage passes only when the implementation result cites allowed contract evidence.
- Template tests:
  - Generated skills, custom agents, README/workflow docs, and design-review scaffolds include the negative-first gate without Claude-only primitives.
- Regression tests:
  - Existing `frontend-minor`, legacy no-brief, malformed-header, design-review mismatch, and approved-frontend-major tests still behave intentionally after the new contract is added.
