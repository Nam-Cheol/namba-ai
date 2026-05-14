# Product Review

- Status: approved
- Last Reviewed: 2026-05-14
- Reviewer: namba-product-manager
- Command Skill: `$namba-plan-pm-review`
- Recommended Role: `namba-product-manager`

## Focus

- Challenge the problem framing, scope, user value, and acceptance bar before implementation starts.

## Findings

- No blocking product issues remain in the revised SPEC.
- The policy now adequately separates ordinary UI chrome from product-specific small assets. Standard design-system icons are treated as existing chrome, while product-specific icons, thumbnails, sprites, cutouts, textures, badges, and similar small assets route into the imagegen-required path when they carry visual identity, reference fidelity, or concrete screen quality.
- Mixed existing/generated usage is defined clearly enough for implementation and review. Mixed-mode screens stay in `Asset mode: generated-images` whenever any uncovered product-specific role still requires generation, while existing-covered roles can be documented in the same manifest.
- Rendered evidence expectations for small assets are sufficiently explicit for product review. The per-asset manifest and run evidence require saved path, intended UI usage, and rendered usage evidence, with proof options beyond hero or section screenshots.

## Decisions

- Accept the revised boundary between ordinary UI chrome and product-specific small generated assets.
- Accept the mixed existing/generated model as the correct default for first-frontend and first-major-screen work.
- Accept the rendered-evidence bar for small assets, provided implementation preserves deterministic per-asset validation and actionable failure messages.

## Follow-ups

- Keep parser and validator implementation scoped so missing `Saved asset path` and missing `Rendered usage evidence` produce asset-by-asset failure messages.
- Keep generated skill and role-card text aligned with the approved mixed-mode and small-asset rules so reviewers and runners do not reintroduce the old ambiguity.

## Recommendation

- Approved for implementation. No further product-track revisions are required before engineering execution.
