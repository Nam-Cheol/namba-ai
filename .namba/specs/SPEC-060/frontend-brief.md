# Frontend Brief

Task Classification: frontend-minor
Classification Rationale: Documentation-only renderer work mentions a hero image, but it does not create a web app, screen, component UI, or new visual asset. The renderer reuses existing repository assets and GitHub Markdown primitives.
Frontend Gate Status: not-applicable
Problem Gate: not-applicable
Reference Gate: not-applicable
Critique Gate: not-applicable
Decision Gate: not-applicable
Prototype Gate: not-applicable
Prototype Evidence: n/a

## Current Pattern

- Generated README and guide documents already include some orientation pieces, but they are produced by large renderer functions with duplicated locale-specific structure.
- The initial classifier treated `hero` as a frontend-major signal, but this SPEC is about GitHub-rendered Markdown documentation, not application UI implementation.

## Intended Change

- Reuse the existing README hero image behavior when configured.
- Add reusable Markdown documentation primitives for hero, badges, CTAs, command selection, workflow tables, quick-start steps, collapsible advanced sections, and cross-document navigation.
- Keep all output GitHub-safe and deterministic.

## Notes

- No new image generation is in scope.
- Browser or screenshot validation is not required unless implementation later introduces a real frontend surface, which this SPEC explicitly avoids.
