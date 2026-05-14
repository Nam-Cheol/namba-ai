# SPEC-045 Baseline

## Existing Frontend Asset Gate

- `internal/namba/frontend_brief.go` owns frontend task classification, brief generation, fixed-label parsing, asset-mode parsing, negative-first contract validation, and readiness reporting.
- `internal/namba/execution.go` owns post-run checks for `## Do-Not Design Violation Check` and `## Generated Asset Evidence`.
- `internal/namba/namba.go` contributes the run prompt text that tells Codex what evidence must be returned.
- `internal/namba/templates.go` owns generated Namba skill text, custom agent mirrors, role-card mirrors, README text, and workflow-guide text.
- `.agents/skills/namba-run/SKILL.md` currently tells runners to include Generated Asset Evidence when `Asset mode: generated-images` is present.
- `.codex/agents/namba-designer.*` and `.codex/agents/namba-frontend-implementer.*` currently mention generated-image plans and evidence, but do not carry the stricter first-frontend asset gate.

## Current Strengths

- `frontend-major` already has a canonical `frontend-brief.md` artifact.
- Existing readiness checks distinguish missing, insufficient, invalid, and mismatched frontend evidence.
- Existing run validation can already fail when Generated Asset Evidence is missing for `Asset mode: generated-images`.
- Existing design and implementation roles already know about the Do-Not Design Contract, asset manifests, and generated-image evidence.

## Current Gaps

- First frontend and first major screen work do not default to generated-image review.
- `Generated Asset Evidence` is enforced by asset mode, not by an explicit pre-run imagegen requirement decision.
- The manifest schema is singular and loose-label based, so multiple small assets can collapse into one weak claim.
- The generated-image plan currently uses `Generation status: complete`, which conflates pre-run readiness with completed generation.
- Existing-assets and not-applicable paths need stricter proof so they do not become loopholes.
- Small generated assets such as product-specific icons, sprites, cutouts, textures, badges, thumbnails, and object images are not clearly treated as first-class evidence.
- CSS gradients, placeholders, abstract shapes, and token-only styling are not explicitly rejected as generated image substitutes.

## Implementation-Risk Notes

- Parser changes must preserve backwards-compatible behavior for historical frontend briefs where practical, while enforcing the stricter contract for new first-frontend and first-major-screen briefs.
- The validator should avoid judging visual taste from code alone; it should verify the contract and evidence fields are present, non-pending, and internally consistent.
- Template-owned surfaces should be updated in source first, then regenerated with `namba regen`.
- Tests should use fixture text and synthetic runner output rather than live image generation.
