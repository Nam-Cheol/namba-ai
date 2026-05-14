# Design Review

- Status: approved
- Last Reviewed: 2026-05-14
- Reviewer: namba-designer
- Command Skill: `$namba-plan-design-review`
- Recommended Role: `namba-designer`
- Evidence Status: not-applicable
- Gate Decision: not-applicable

## Focus

- Clarify art direction, palette discipline, anti-generic composition, purposeful motion, and visual risks before implementation starts.

## Approved Direction

- Treat this SPEC as an asset-gate contract review, not a page-style review.
- Default first frontend and first major screen work to `Asset mode: generated-images`, then require explicit proof for `not-applicable` or `covered-by-existing-assets` decisions.
- Count icons, small object images, thumbnails, sprites, textures, cutouts, badges, and similar small generated assets as first-class evidence when they are saved, wired, and rendered in the intended UI element.

## Banned Patterns

- Token-only styling presented as a substitute for concrete visual assets.
- CSS gradients, abstract shapes, empty placeholders, and manually drawn decorative forms used to bypass image generation.
- Generic SaaS card walls used as a fallback visual identity for asset-dependent surfaces.
- `existing-assets` used as a loophole without proof that user-provided assets fully cover the required visual roles.
- Evidence standards that only recognize hero or section imagery and ignore valid small generated assets.

## Negative-First Contract

- The contract defines invalid substitutes first, then defines what valid evidence must prove.
- Failure messages should distinguish missing `Asset mode`, weak `Not-applicable proof`, missing manifest entries, missing saved asset paths, and missing rendered usage evidence for the intended UI element.

## Reference-Driven Asset Manifest

- Approved. The manifest must stay asset-granular.
- Each icon, thumbnail, sprite set, texture, cutout, object image, or other generated asset should remain individually attributable by role, prompt/spec, output path, intended placement, and validation evidence.

## Generated Image Plan

- Approved. Generated imagery is treated broadly and is not overfit to hero art.
- Valid paths include product-specific icon packs, product/object thumbnails, sprites, cutouts, textures, micro-illustrative surfaces, small badges, object markers, and other concrete UI imagery whose value is local to one component rather than page-scale.

## Visual Grammar

- Concrete asset evidence over decorative approximation.
- Per-asset traceability over aggregate claims.
- Rendered usage proof over file-exists proof.
- Explicit exception paths over omission.
- Semantic allowance for small assets, not just hero media.

## Open Questions

- None blocking.

## Recommendation

- Approved for implementation. The asset gate is strong enough to block generic visual fallback while allowing icons, thumbnails, sprites, textures, and other small generated assets as first-class evidence.
