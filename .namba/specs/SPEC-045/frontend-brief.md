# Frontend Brief

Task Classification: frontend-minor
Classification Rationale: This SPEC changes NambaAI frontend execution policy, validators, prompts, skills, and role cards, but it does not implement a user-facing frontend screen. Use the lightweight advisory path for this SPEC's own brief while enforcing the new `frontend-major` asset policy for future runs.
Frontend Gate Status: not-applicable
Problem Gate: not-applicable
Reference Gate: not-applicable
Critique Gate: not-applicable
Decision Gate: not-applicable
Prototype Gate: not-applicable
Prototype Evidence: n/a

## Problem Frame

- Problem statement: First frontend implementations can still pass with token-only styling or generic substitutes when the screen actually needs imagegen-backed visual assets.
- User goal: Make imagegen review and evidence a default, enforceable step for first frontend work and first major screen work.
- Target user: NambaAI operators and Codex agents executing frontend SPECs.
- Success metric: Future `frontend-major` runs fail when imagegen is required but the brief or run output lacks the required manifest and evidence.
- Why now: Existing frontend gates already know about asset manifests and generated evidence, so the remaining gap is policy strictness and small-asset coverage.
- Scope boundary: This SPEC updates workflow and validation behavior, not a product UI.

## Asset Evidence

- Brand assets: Not applicable for this SPEC because no product screen is being implemented.
- Product or domain imagery: Not applicable for this SPEC because no end-user page is being designed.
- Generated image assets required: Not applicable for this SPEC's own implementation. The SPEC defines when future frontend work must use imagegen.
- Asset manifest path: Not applicable for this SPEC.
- Existing UI screenshots: Not applicable.
- Asset constraints and gaps: Future frontend briefs must treat generated image assets broadly and must not limit evidence to hero or section imagery.

## Policy Contract For Future Frontend Runs

### Decision Fields

- Frontend implementation phase: first-frontend, first-major-screen, incremental, or not-applicable.
- Asset mode: generated-images, existing-assets, or not-applicable.
- Imagegen requirement: required, covered-by-existing-assets, or not-applicable.
- Asset decision proof: Required when Imagegen requirement is not `required`; must explain why existing assets fully cover the visual roles or why generated imagery is not appropriate.
- Default: first-frontend and first-major-screen begin as `Asset mode: generated-images` with `Imagegen requirement: required`.

### Generated-Images Definition

- `generated-images` means imagegen-created visual assets saved into the project and rendered in the UI.
- Valid generated assets include hero images, section imagery, icons, small object images, product thumbnails, cutouts, textures, sprites, badges, illustrations, and other concrete visual assets.
- A generated asset is valid only when the run result proves the saved project path and rendered usage in the UI element the asset was created for.
- Generated asset scope is broad, but the requirement is specific: imagegen is mandatory for small assets only when they are product-specific visual content, not when ordinary UI chrome is already covered by a design-system or icon-library primitive.

### Small Asset Boundary

- Existing UI chrome example: search, close, menu, save, pagination, and status icons from a standard design system may remain existing assets when they do not carry product-specific visual identity.
- Imagegen-required small asset examples: product-specific icon sets, object thumbnails, commerce inspection images, game sprites, custom badges, cutouts, branded micro-illustrations, textures, or illustrated markers that carry visual identity, reference fidelity, or concrete screen quality.
- Decorative textures or patterns are optional unless the reference or product category makes them part of the concrete visual quality bar. If generated, they still require manifest and rendered usage evidence.

### Default Mode

- First frontend implementations and first major screen implementations start in `Asset mode: generated-images`.
- The implementation may move to `Asset mode: not-applicable` only with explicit proof that imagery would harm the UX.
- Existing assets may be documented, but they must not become a loophole that avoids imagegen when the brief's visual quality depends on generated concrete assets.
- `Asset mode: existing-assets` is valid for first frontend work only when the brief proves user-provided or design-system assets fully cover every concrete visual role that would otherwise require imagegen.
- Mixed existing/generated usage keeps `Asset mode: generated-images`; the manifest must distinguish existing inputs from generated outputs.

### Required Imagegen Triggers

- User-provided design system or brand reference depends on photos, product visuals, illustration, iconography, object imagery, or other concrete assets.
- The requested screen is product, place, person, object, marketing, game, commerce, consumer, or otherwise visual-asset-dependent.
- The first screen or core workflow would otherwise fall back to gradients, abstract shapes, empty placeholders, generic SaaS card walls, or token-only styling.
- A positive trigger requires both visual-asset dependence and an uncovered concrete visual role. A non-example is a dense internal table that uses standard design-system icons only for sorting or filtering.

### Not-Applicable Proof

- Valid examples: pure data tables, internal operations tools, CLI-like admin screens, text-heavy settings screens, or other cases where images would reduce clarity.
- Required proof: name the screen type, explain why visual assets would harm the UX, identify the replacement visual grammar, and include validator-readable `Not-applicable proof`.
- Invalid proof: "no assets provided", "time constraints", "will add later", or silent omission.

### Reference-Driven Asset Manifest

- Required when imagegen is used or required.
- Each entry must use repeatable per-asset blocks with exact labels:
  - Asset ID
  - Asset type
  - Asset source: generated, existing-input, or existing-covered
  - Required status: required, supporting, or covered
  - Role in UI
  - Reference signal
  - Generation prompt/spec
  - Source input path
  - Saved asset path
  - Intended UI usage
  - Rendered usage evidence
- The manifest must cover small assets as first-class items, including icons, thumbnails, sprites, textures, cutouts, and object images.
- The manifest must stay asset-granular. Do not collapse multiple generated assets into a vague "supporting visuals" entry unless they are a single sprite sheet or set with one tightly scoped UI role.
- For mixed-mode screens, existing assets may appear as source inputs, but generated outputs still need output path, intended UI usage, and validation evidence.

### Generated Asset Evidence

- Required in `namba run` results when imagegen is used or required.
- Evidence must include manifest path, generated files, prompt summary, saved project asset paths, and rendered usage evidence.
- Rendered usage evidence may cite screenshots, browser verification, component render output, or another deterministic local proof that the intended UI element uses the generated asset.
- For small assets, sufficient evidence may include screenshot callouts, component render output, DOM/file-path linkage, Storybook or preview evidence, or test-render evidence tied to the intended icon, thumbnail, sprite, texture, cutout, badge, or object image.
- Evidence must be repeatable per asset:

```markdown
## Generated Asset Evidence
- Manifest path: .namba/specs/SPEC-XXX/frontend-brief.md
- Asset ID: product-state-icon
  Generated file: src/assets/product-state-icon.png
  Saved asset path: src/assets/product-state-icon.png
  Prompt summary: Product-specific state icon aligned to the approved reference.
  Intended UI usage: Status selector icon in the first-run setup panel.
  Rendered usage evidence: Component render output includes src/assets/product-state-icon.png in the status selector.
```

### Generation Plan

- The frontend brief should use `Generation plan status: ready` when the imagegen plan is complete enough for `namba run`.
- Completed generation belongs in `## Generated Asset Evidence` after implementation. The brief must not require `Generation status: complete` before the runner has generated, saved, and rendered the assets.

## Do-Not Design Contract

Contract Status: not-applicable for this SPEC's own implementation; required policy changes are listed above.

### Default Anti-Pattern Library

- CSS gradients, abstract shapes, empty placeholders, generic SaaS card walls, manually drawn decorative shapes, and token-only styling must not be accepted as generated image substitutes.
- Designs whose visual core is photo, product, illustration, icon, object, sprite, or concrete image content must not pass by applying only color tokens or CSS decoration.

### Frontend Architecture Handoff

- Allowed planning: Update Go validators, run prompts, generated templates, skill guidance, role cards, and deterministic tests.
- Out of scope: Building an actual frontend page or generating image assets for this SPEC.
- Required evidence: Tests proving first frontend defaulting, not-applicable proof, required Generated Asset Evidence, small-asset evidence, and rejected substitutes.
- Responsive and accessibility constraints: Future generated assets must still preserve accessibility, contrast, loading behavior, and responsive layout, but this SPEC validates the workflow contract rather than a specific UI.

## Open Decisions

- No blocking product or design decisions remain in the planning artifact.
- Implementation detail to settle in code: the parser shape for repeatable `Asset ID` evidence blocks should preserve deterministic failure messages for missing saved paths and missing rendered usage evidence.
