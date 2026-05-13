# Frontend Brief

Task Classification: frontend-minor
Classification Rationale: Documentation readability work touches presentation and hierarchy in Markdown, but it does not change application UI behavior or require a frontend prototype gate.
Frontend Gate Status: not-applicable
Problem Gate: not-applicable
Reference Gate: not-applicable
Critique Gate: not-applicable
Decision Gate: not-applicable
Prototype Gate: not-applicable
Prototype Evidence: n/a

## Problem Frame

- Problem statement: NambaAI docs are correct but low-scan, especially across the multilingual getting-started and workflow manuals.
- User goal: Make the manual set visibly easier to navigate by borrowing the high-readability structure of the Ouroboros Korean README reference.
- Target user: NambaAI operators, maintainers, and Codex users who need to choose the right command without rereading long prose.
- Success metric: A reader can identify the right install path, command path, workflow phase, and deeper reference from any supported locale in under a minute.
- Why now: The docs already cover the right commands; the next leverage point is presentation and translation parity.
- Scope boundary: Documentation structure and content design only. No CLI behavior, release automation, or generated README ownership changes unless the owning source is updated.

## Asset Evidence

- Brand assets: Existing README uses `assets/images/namba-ai-hero.png`; manual pages should not require new imagery.
- Product or domain imagery: Not needed for GitHub docs.
- Existing UI screenshots: Not applicable.
- Asset constraints and gaps: Use Markdown-native hierarchy, tables, badges, and short HTML blocks only when GitHub-safe.

## Reference Set

- Reference 1: Ouroboros Korean README.
  Adopt: language navigation, top identity, badge or status link row, compact contents, quick-start emphasis, command tables, and strong section breaks.
  Avoid: copying product-specific philosophy or decorative motifs that do not fit NambaAI.
  Why: It is the user's explicit readability benchmark.
- Reference 2: Current NambaAI README.
  Adopt: current product promise, command taxonomy, generated ownership warning, and links to key docs.
  Avoid: letting root README edits drift from sync-managed ownership.
  Why: It is the current public landing surface.
- Reference 3: Current localized getting-started and workflow guides.
  Adopt: existing factual content and locale coverage.
  Avoid: preserving the current flat hierarchy when tables or quick-start blocks would make decisions clearer.
  Why: They are the target manual corpus.

## Direction Alternatives

- Direction A: Full README-style transformation of every manual page.
  Tradeoff: Strongest visual impact, but risks over-formatting short docs and increasing translation maintenance.
- Direction B: Reusable documentation blueprint per manual family.
  Tradeoff: Balanced impact and maintainability; keeps each page purposeful while using the reference's proven scanning patterns.
- Direction C: Minimal polish only.
  Tradeoff: Lowest churn, but unlikely to satisfy the user's request for active adoption of badges, contents, card sections, quick start, and feature tables.
- Selected direction rationale: Direction B gives the manuals a visible upgrade without turning every page into a decorative landing page.

## Synthesis

- UX metaphor: Command cockpit. The reader lands, picks the correct command lane, and then dives into details.
- Section roles: Top language switcher and status links orient; quick start accelerates; command tables help choose; workflow sections explain sequence; reference links preserve depth.
- Hierarchy: Keep top sections short, tables decision-oriented, and explanatory prose below the fold.
- Reference synthesis: Borrow structure from Ouroboros, keep NambaAI's pragmatic Codex workflow voice.
- Anti-generic bans: No ornamental badges, no redundant card grids, no one-language polish, no unsupported claims.
- Typography scale: Markdown-native headings only; avoid visual tricks that GitHub renders inconsistently.
- Spacing and density intent: More scannable than current docs, but still compact enough for maintainers.
- Depth and container budget: Use tables and short callouts where they clarify decisions; avoid nesting cards inside cards.

## Design Review Axes

- Evidence fit: Strong; target docs and explicit external reference are known.
- Asset fidelity: Strong; no new imagery required.
- Alternative coverage: Adequate; three documentation directions considered.
- Visual hierarchy: Must be validated in rendered GitHub Markdown or a close local preview.
- Craft and detail: Translation parity and link hygiene are the main craft risks.
- Functionality and accessibility: Tables and HTML must remain readable in plain Markdown and on narrow screens.
- Differentiation without novelty drift: Use the reference for visibility, but keep NambaAI workflow-first.

## Prototype Evidence

- Artifact path or link: Changed Markdown files during implementation.
- Notes: A separate visual prototype is not required; rendered Markdown review is sufficient.

## Open Decisions

- Whether supporting reference docs receive only navigation links or a fuller readability pass should be decided after the implementation inventory.
- Whether README output changes are needed depends on locating the sync renderer or config ownership path.
