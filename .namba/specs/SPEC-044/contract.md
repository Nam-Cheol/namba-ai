# SPEC-044 Contract

## Invariants

- Extend the existing `frontend-major` synthesis gate; do not introduce a second frontend artifact outside `.namba/specs/<SPEC>/frontend-brief.md`.
- Keep `frontend-brief.md` the canonical parser-visible source for frontend gate state.
- Keep `reviews/design.md` and `reviews/readiness.md` as summaries and review surfaces, not independent sources that can override the canonical brief.
- Preserve advisory behavior for non-frontend, `frontend-minor`, and legacy no-brief work.
- Block `frontend-major` architecture and implementation when the negative-first contract is missing, insufficient, contradicted, or violated.
- Treat the default anti-pattern library as a starting point, not a universal ban list; context-specific reasoning decides what is banned, allowed, or excepted.
- Require an allowed replacement or exception path for every context-specific banned pattern.
- Keep the implementation Codex-native and repository-local; do not add Claude-only primitives or third-party design-tool dependencies.

## Required Do-Not Design Contract

Every `frontend-major` brief must capture:

| Field | Required behavior |
| --- | --- |
| Contract status | Distinguishes complete, missing, and insufficient negative-first evidence. |
| Default anti-pattern library | Names the baseline generic patterns considered during synthesis. |
| Context-specific banned patterns | Lists SPEC-specific bans with rationale and detection hints. |
| Allowed replacement patterns | Names the implementation patterns that should replace each ban. |
| Brand/category/trust reasoning | Explains how category norms, brand assets, trust burden, and audience workflow constrain the direction. |
| Visual grammar contract | Defines layout primitives, hierarchy, type, color, density, depth, imagery/icons, and motion. |
| Most-generic-section redesign proof | Identifies the weakest/generic section and proves the replacement direction. |
| Frontend architecture handoff | Gives `namba-frontend-architect` concrete component, state, responsive, and accessibility constraints. |
| Post-implementation violation checks | Defines the evidence required after coding and what blocks success. |

## Default Anti-Pattern Library

The v1 library must include at least:

- generic card grids as the primary page grammar
- bento grids used as visual identity instead of information architecture
- card-inside-card nesting and double-background framing
- decorative glass, blurred blobs, gradient orbs, or unearned gradients
- stock SaaS hero formulas
- interchangeable feature-card walls
- generic testimonial, FAQ, pricing, and stats sections
- dashboard KPI-card rows used against the task model
- low-information empty states
- weak typography defaults such as tiny body text or too many styles
- one-note palettes and decorative color without state or brand meaning
- stock-like imagery where real product/domain/state inspection is needed

## Context-Specific Ban Schema

Each banned pattern should be reviewable through this shape:

- Banned pattern
- Why banned here
- Detection hints
- Allowed replacement
- Exception path

## Blocking Rules

- A complete five-gate header is not enough for `frontend-major` approval.
- `Gate Decision: approved` and `Evidence Status: complete` are not enough when the Do-Not Design Contract is missing or weak.
- Missing replacement patterns block, because a ban without an allowed path creates vague taste policing.
- Missing brand/category/trust reasoning blocks, because visual grammar must come from context.
- Missing architecture handoff blocks, because implementation planning cannot safely infer component boundaries from style prose.
- A failed post-implementation violation check blocks the run and must surface a concrete banned pattern plus remediation.

## Deterministic Helper Candidate Criteria

If implementation adds a standalone negative-first checker, it must:

- support `--help`
- run read-only by default
- use bounded text or JSON output
- avoid network and authentication assumptions
- work with fixture files or a local test app
- report actionable path/line or artifact evidence
- avoid destructive repo or third-party mutations

