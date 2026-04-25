# SPEC-036

## Goal

Replace the current frontend implementation prerequisite of "user-provided references exist" with "reference synthesis is complete" so `frontend-major` work cannot move into architecture or implementation until Namba has collected, critiqued, and approved a concrete direction.

## Context

- Project: namba-ai
- Project type: existing
- Language: go
- Mode: tdd
- Work type: plan
- Verified local context as of 2026-04-23:
  - `buildSpecReviewReadinessDoc` always renders review readiness as advisory-only and summarizes only product, engineering, and design tracks; there is no frontend-specific gate summary or selective hard-block path today. (`internal/namba/spec_review.go`)
  - `buildSpecReviewDoc` seeds `reviews/design.md` with a generic review shape (`Status`, `Findings`, `Decisions`, `Follow-ups`, `Recommendation`) and no fields for evidence status, gate decision, approved direction, banned patterns, or unresolved questions. (`internal/namba/spec_review.go`)
  - The generated `namba run` command skill explicitly says review readiness is advisory by default and routes UI work as design -> frontend planning -> approved UI implementation, but it does not require reference synthesis or design clearance before UI coding. (`internal/namba/templates.go`)
  - Synced docs repeat the same advisory contract and describe `namba-frontend-implementer` as handling approved UI work without a formal reference-synthesis gate. (`README.md`, `docs/workflow-guide.md`, `docs/workflow-guide.ko.md`)
  - The generated role cards for `namba-designer`, `namba-frontend-architect`, and `namba-frontend-implementer` describe design direction, frontend planning, and approved UI implementation, but none of them define a mandatory pre-implementation frontend evidence gate. (`internal/namba/templates.go`, `.codex/agents/namba-designer.md`, `.codex/agents/namba-frontend-architect.md`, `.codex/agents/namba-frontend-implementer.md`)
  - Feature planning scaffolds currently create `spec.md`, `plan.md`, `acceptance.md`, and `reviews/*.md`; there is no frontend-specific planning artifact such as `frontend-brief.md`. (`internal/namba/namba.go`, `internal/namba/spec_command_test.go`)

## Problem

Namba already carries strong anti-generic UI guidance, but the workflow still allows teams to jump from a frontend request straight into hierarchy planning or code before anybody proves that the problem, references, critique, and directional choices are grounded. That leaves three structural gaps:

1. "Approved UI work" is ambiguous because approval can happen without a documented evidence trail.
2. Design review artifacts do not say whether the reference set is complete, what direction is approved, or which patterns are explicitly banned for the SPEC.
3. `namba run` treats missing frontend research as invisible advisory context instead of a selective execution blocker for high-risk `frontend-major` work.

The result is predictable drift toward generic UI output: bootstrap/card overuse, excessive nested containers, weak typography hierarchy, gratuitous glass or gradients, and section-level style changes that are not justified by the product or state model.

## Desired Outcome

- Namba classifies frontend requests as `frontend-major` or `frontend-minor`.
- `frontend-major` work requires five hard gates before implementation: Problem Gate, Reference Gate, Critique Gate, Decision Gate, and Prototype Gate.
- User-provided references become optional input, not the gate itself. When references are missing or insufficient, Namba must collect current external references first.
- `frontend-brief.md` becomes the canonical frontend planning artifact for new frontend-touching work: `frontend-major` uses the full five-gate contract, while `frontend-minor` uses the same artifact to persist explicit classification, a short rationale, and lightweight `not-applicable` gate state.
- Planning artifacts persist the evidence needed to explain why a direction was chosen, what was rejected, and what implementation is allowed to do next.
- Design judgment is anchored to evidence, assets, alternatives, and review axes instead of reviewer taste or unstructured visual preference.
- General plan-review readiness stays advisory for the rest of the system, but `frontend-major` execution is allowed to block when the frontend gate is missing required evidence.
- Blocked frontend runs tell the operator exactly what to do next: gather or replace references, improve weak synthesis, add prototype evidence, or split mixed work into separate SPECs/phases when independent delivery matters.
- Role routing becomes explicit: `namba-designer` owns research and synthesis, `namba-frontend-architect` plans structure only after the gate is satisfied, and `namba-frontend-implementer` codes only after synthesis plus design clearance.

## Frontend Task Classification

- `frontend-major`
  - New screen or major new section
  - Landing or marketing page work
  - Dashboard restructure
  - Significant visual tone change
  - Redesign of a core component or interaction model
- `frontend-minor`
  - Bug fix within an existing pattern
  - Small spacing or alignment correction
  - Additional state inside an established component pattern
  - Copy replacement without hierarchy or layout change
- If classification is ambiguous, treat the `frontend-major` tie-break rule as authoritative when the request changes hierarchy, visual language, or the shape of a primary workflow.
- Fix-only requests with lightweight maintenance signals such as spacing, alignment, button, copy, existing-component work, or a broad surface-only dashboard bugfix stay `frontend-minor`, unless they also include a structural redesign signal.
- Broad nouns such as dashboard, settings, form, component, or page, and overloaded structure words such as hierarchy, do not classify work as frontend-touching by themselves when paired only with backend/API/storage signals; generic text/copy terms do not override that backend-only filter. They need an explicit UI/screen/layout signal or a non-generic major/minor frontend signal so backend/API work is not blocked by frontend synthesis.
- Overloaded document-structure wording such as `section` is not an explicit frontend-touch signal by itself; documentation-only requests such as README section updates must not scaffold a frontend gate without additional UI evidence.
- Explicit `frontend` or `front-end` wording counts as a frontend-touch signal so obvious UI-layer work cannot bypass `frontend-brief.md` scaffolding.
- For new SPECs that touch frontend scope, `frontend-brief.md` must persist `Task Classification` plus a short `Classification Rationale`; `namba run` must not infer major/minor status from freeform prose at execution time.

## Canonical Gate Contract

- `frontend-brief.md` is the canonical source of truth for new frontend-touching classification and for `frontend-major` evidence and gate state.
- New frontend-touching SPECs always get `frontend-brief.md`. `frontend-minor` work keeps the artifact lightweight by recording classification, classification rationale, current pattern, intended change, and `not-applicable` gate fields instead of the full research burden.
- `reviews/design.md` and `reviews/readiness.md` summarize and evaluate that state, but they do not redefine it independently.
- `reviews/readiness.md` must include frontend gate state in the initial planning scaffold and fold frontend gate advisory state into its summary blocker line so missing, insufficient, invalid, or mismatched frontend gates cannot appear alongside an all-clear summary.
- The top block of `frontend-brief.md` must remain machine-readable through fixed labels so template/render/parser code can evolve without guessing from prose.
- v1 required fixed-label fields:
  - `Task Classification: frontend-major | frontend-minor`
  - `Classification Rationale: <short explanation>`
  - `Frontend Gate Status: approved | blocked | needs-research | not-applicable`
  - `Problem Gate: complete | missing | insufficient | not-applicable`
  - `Reference Gate: complete | missing | insufficient | not-applicable`
  - `Critique Gate: complete | missing | insufficient | not-applicable`
  - `Decision Gate: complete | missing | insufficient | not-applicable`
  - `Prototype Gate: complete | missing | insufficient | not-applicable`
  - `Prototype Evidence: wireframe | annotated-layout | prototype | equivalent | n/a`
- Historical SPECs remain runnable by default unless they are explicitly marked `frontend-major`; missing `frontend-brief.md` on legacy SPECs must not trigger speculative gating.

## Parser Validity And Precedence

- Unknown required labels, unsupported enum values, missing required fixed-label lines, or required labels left blank in `frontend-brief.md` put the artifact in an explicit invalid-contract state; the runner must not guess the intended meaning from nearby prose.
- Invalid-contract artifacts hard-block execution only when the fixed-label `Task Classification` is not exactly `frontend-minor`; exact `frontend-minor` invalid-contracts remain visible in readiness as advisory repair work instead of forcing the full major gate.
- Impossible combinations such as `Frontend Gate Status: approved` with any `missing` or `insufficient` major gate, or `frontend-minor` combined with non-`not-applicable` major-gate states, are invalid until corrected.
- `frontend-brief.md` remains the canonical machine-readable source of truth. `reviews/design.md` and `reviews/readiness.md` summarize and evaluate that state.
- If summarized review docs disagree with the canonical `frontend-brief.md` header state, readiness must surface the mismatch explicitly instead of silently picking one interpretation.

## Five-Gate Contract

### Problem Gate

- Record `user goal`, `target user`, `success metric`, `why now`, and `in/out of scope`.
- Capture one core UX metaphor plus section-role statements that explain what each section must do.

### Reference Gate

- Require at least three references for `frontend-major` work.
- Treat user references as optional. If the user did not provide them, Namba must gather them before moving forward.
- When current external research is constrained or low-signal, authoritative user-provided references or repo-local references may satisfy the minimum set, but weak synthesis still counts as `insufficient`.
- Record asset evidence separately from references: brand assets, product or domain imagery, existing UI screenshots, asset constraints, and known gaps.
- For each reference, record `adopt`, `avoid`, and `why`.
- Use authority guides such as Apple HIG, Primer, Carbon, Atlassian, GOV.UK, Material, and NN/g as normative grounding, and practical product examples such as Linear, Stripe, Vercel, Attio, and Raycast as adopt/avoid heuristics rather than style-copy targets.

### Critique Gate

- Produce a reference synthesis, not just a list of links.
- Record anti-generic bans, typography intent, spacing and density intent, depth and container budget, primary hierarchy explanation, and at least three direction alternatives with tradeoffs before selecting one.
- State why proposed layout primitives match the product/domain/state model better than the most obvious generic card/grid fallback.
- Evidence quality matters, not just artifact presence. A filled section that does not narrow execution ambiguity should count as `insufficient`, not `complete`.

### Decision Gate

- `reviews/design.md` must record:
  - `Evidence Status`
  - `Gate Decision: approved | blocked | needs-research`
  - `Approved Direction`
  - `Banned Patterns`
  - `Open Questions`
  - `Unresolved Questions`
  - `Design Review Axes`
  - `Keep / Fix / Quick Wins`
- `approved` means architecture and implementation may proceed.
- `blocked` or `needs-research` means work routes back into synthesis instead of UI implementation.
- A `frontend-major` run blocked solely by `Frontend Gate Status: blocked` or `needs-research` must still emit concrete remediation that points back to design review, frontend-brief synthesis, and the approval transition.
- If `reviews/design.md` disagrees with the canonical gate state in `frontend-brief.md`, treat the Decision Gate as unresolved until the artifacts are reconciled.
- For `frontend-major`, `Gate Decision: approved` and `Evidence Status: complete` are not enough by themselves; `Approved Direction`, `Banned Patterns`, `Open Questions`, and `Unresolved Questions` must be resolved from their scaffolded `pending` placeholders before execution can proceed.
- Those decision detail fields may use inline values or indented Markdown continuation lines; populated continuation bullets must not be treated as `pending`, while placeholder variants such as `- pending` or `Pending.` must still block execution.
- For `frontend-major`, frontend architecture planning and frontend implementation stay blocked until the Decision Gate is `approved` and the other four gates are at least `complete`.

### Prototype Gate

- `frontend-major` work must record wireframe or prototype expectation before coding starts.
- v1 accepts low-fidelity but reviewable evidence such as a wireframe, annotated layout, interaction flow sketch, or equivalent structure that makes hierarchy and interaction direction concrete.
- Missing prototype expectation counts as `blocked` or `needs-research`; it is not a silent omission.

## Anti-Generic Rules

These rules exist to block systemically weak output, not to force one house style.

### Cards

- Do not default to cards as the primary layout grammar.
- Allow cards only when they represent a real standalone concept unit.
- Prefer lists, tables, sections, or single-column flows for homogeneous lists, state groups, and long explanations.

### Typography

- Do not use a base body size below `16px`.
- Prefer `17px` for long-form or mobile-primary body copy when the context supports it.
- Limit a screen to at most six text-style tokens.
- Keep heading steps explicit, e.g. `16 / 20 / 24 / 32`.
- Avoid light weights on small text.

### Layers And Containers

- Default to flat structure plus spacing and alignment.
- Limit depth semantics to roughly `base / raised / overlay`.
- Ban card-inside-card, meaningless double backgrounds, and combined shadow + border + tint + radius overuse.
- Use shadows or elevation only when overlay, drag, or clear focal semantics justify them.

### Visual Language

- Do not let each hero or section invent a different aesthetic.
- Ban unjustified gradient or glass treatments.
- Repeat a small, coherent visual grammar across sections.
- Let product UI, domain language, and state flow drive hierarchy before decoration does.

## Scope

- Add frontend task classification plus persisted planning metadata for `frontend-major` vs `frontend-minor`.
- Extend planning scaffolds (`namba plan`, `namba harness`, `namba fix --command plan`) so new frontend-touching work seeds `frontend-brief.md`, with the full evidence contract for `frontend-major` and the lightweight classification/rationale path for `frontend-minor`.
- Expand design review rendering so it can express evidence status, gate decision, approved direction, banned patterns, and unresolved questions.
- Expand readiness aggregation so it shows a dedicated frontend gate summary alongside the existing advisory review-track summary, including invalid-contract or mismatch cases when frontend artifacts disagree.
- Add pre-execution gate checks to `namba run SPEC-XXX` so `frontend-major` work returns a clear "blocked for frontend synthesis" result when required artifacts or evidence fields are missing, insufficient, or internally invalid.
- Update role cards, command skills, README/workflow guidance, and synced docs to explain the research-before-implementation rule and revised role ownership.
- Add regression coverage for scaffold generation, review rendering, readiness aggregation, classification, invalid-contract handling, selective run blocking, operator guidance, and legacy SPEC compatibility.

## Public Interfaces And Contracts

- New artifact:
  - `.namba/specs/<SPEC>/frontend-brief.md`
- Expanded planning contract:
  - `spec.md` carries narrative goal, scope, risks, and rollout constraints.
  - `frontend-brief.md` is the canonical frontend sidecar for task classification, classification rationale, gate state, asset evidence, agent-collected references, adopt/avoid/why, direction alternatives, reference synthesis, anti-generic bans, typography scale, spacing/density intent, depth/container budget, design review axes, prototype evidence, open questions, and decisions.
  - `frontend-brief.md` must distinguish `missing` evidence from `insufficient` evidence so operators know whether to create material or improve weak material.
  - `frontend-minor` uses the same artifact to make classification and rationale visible without forcing the full five-gate burden.
- Expanded `reviews/design.md`:
  - `Evidence Status`
  - `Gate Decision`
  - `Approved Direction`
  - `Banned Patterns`
  - `Open Questions`
  - `Unresolved Questions`
  - `Design Review Axes`
  - `Keep / Fix / Quick Wins`
- Expanded `reviews/readiness.md`:
  - keep the existing advisory product/engineering/design summary
  - add a separate Frontend Gate summary when the SPEC uses `frontend-brief.md`
  - reflect `missing` versus `insufficient` gate states distinctly
  - surface invalid fixed-label contracts and cross-artifact mismatches explicitly
- Execution contract:
  - `namba run SPEC-XXX` blocks `frontend-major` implementation when frontend evidence is incomplete or internally contradictory.
  - When blocked, the output must list which gates or fixed-label fields are `missing`, `insufficient`, or invalid and route the operator toward research, synthesis, critique, prototype work, or gate reconciliation instead of silently failing.
  - When a mixed frontend-major/non-frontend SPEC is blocked, the output must also recommend split-SPEC or phased delivery instead of implying runtime partial unblocking exists.
  - `namba run` may only apply the full five-gate block when the persisted task classification is explicitly `frontend-major`.
  - If `frontend-brief.md` exists but its required fixed-label contract is malformed or contradictory, `namba run` returns an explicit contract error rather than inferring a state from prose.
  - `frontend-brief.md` wins over summarized review prose for machine enforcement, but visible disagreements between the canonical artifact and review summaries count as unresolved contract mismatches for `frontend-major` execution.
  - v1 does not require slice-aware partial-run unblocking for mixed SPECs. When independent backend/non-frontend delivery matters, planners should split mixed work into separate SPECs or explicit phases instead of relying on runtime inference.
- Role contract:
  - `namba-designer`: collect references, synthesize adopt/avoid/why, define anti-generic direction, propose banned patterns.
  - `namba-frontend-architect`: verify hierarchy and file/state planning only after the gate prerequisites exist.
  - `namba-frontend-implementer`: implement only reference-synthesized, design-cleared UI work.

## Non-Goals

- Do not convert all missing review artifacts into a global hard gate.
- Do not require `frontend-brief.md` for historical SPECs that predate this contract.
- Do not force `frontend-minor` work through the full five-gate process.
- Do not turn practical product references into style-copy mandates.
- Do not expand this v1 into a Figma-native approval system or a full design-ops platform.

## Design Constraints

- Preserve advisory review-readiness behavior for non-frontend or `frontend-minor` work.
- Keep legacy SPECs runnable unless they are explicitly classified as `frontend-major` under the new rules.
- Make classification transparent enough that operators can tell why a SPEC was treated as major or minor.
- Keep the gate readable in generated docs and artifacts; blocked work should redirect into evidence creation, not opaque errors.
- Keep blocked-run output action-oriented so operators can tell whether to gather references, deepen critique, add prototype evidence, fix an invalid contract, or split mixed scope.
- Update repo-managed templates instead of relying on manual file-by-file edits to synced docs or role cards.
- Follow TDD: tests should fail first for scaffolds, rendering, routing, and gate enforcement.

## Defaults And Assumptions

- v1 minimum reference set: three items.
- v1 hard gates apply only to `frontend-major`.
- v1 default artifact for new frontend-touching planning evidence is `frontend-brief.md`.
- `frontend-minor` persists explicit classification and rationale through the same artifact, with `not-applicable` gate fields instead of silent omission.
- v1 canonical parser target is the fixed-label header block in `frontend-brief.md`, not design-review prose.
- Authority guides are normative sources; real product examples are heuristic inputs.
- Global readiness remains advisory, but `frontend-major` execution may hard-block when the frontend gate is incomplete.

## Validation Strategy

- Planning
  - New landing-page request with no user references seeds agent-collected references plus `frontend-brief.md` instead of jumping straight to implementation planning.
  - New dashboard restructure with complete evidence records `Approved Direction` and `Banned Patterns` in `reviews/design.md`.
  - Minor UI fixes still seed `frontend-brief.md`, but only with explicit classification, classification rationale, current pattern/change summary, and `not-applicable` major-gate fields.
- Gating
  - `frontend-major` plus no reference synthesis blocks `namba run`.
  - `frontend-major` plus references but no adopt/avoid/why blocks `namba run`.
  - `frontend-major` plus no critique note blocks `namba run`.
  - `frontend-major` plus no prototype expectation yields `blocked` or `needs-research`.
  - Malformed or contradictory fixed-label headers return an explicit contract error instead of falling back to prose inference.
  - `frontend-brief.md` and `reviews/design.md` disagreement is surfaced in readiness and blocks `frontend-major` execution until reconciled.
- Rule enforcement
  - Card-only repetition plans are rejected by design review.
  - `14px` body-first plans are rejected by design review.
  - Deep nested container plus shadow-heavy plans are rejected by design review.
  - Section-by-section visual-language drift is rejected by design review.
- Regression
  - New `frontend-minor` SPECs persist visible classification rationale without entering the full five-gate burden.
  - Non-frontend domains keep advisory review behavior.
  - Historical SPECs do not fail simply because they lack `frontend-brief.md`.
  - Blocked-run output lists concrete remediation next steps, including split-SPEC guidance for mixed work.
  - Generated docs, role cards, and command skills stay aligned with the new contract.
