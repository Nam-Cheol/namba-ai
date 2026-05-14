# SPEC-045

## Problem

NambaAI already has a `frontend-major` gate with a Do-Not Design Contract, a reference-driven asset manifest, and run-result evidence expectations. The current contract still leaves too much room for a first frontend implementation to pass with token-only styling, CSS decoration, abstract placeholders, or a generic SaaS card wall when the screen quality actually depends on concrete visual assets.

The risk is highest on first frontend implementations and first major screen implementations. In those moments, Codex is establishing the visual grammar of the product, and the default path should assume generated visual assets are needed unless the task is clearly harmed by imagery. The policy must also treat generated assets broadly: `generated-images` means any imagegen-created UI asset, not only large hero or section imagery. Icons, small object images, product thumbnails, cutouts, textures, sprites, and other specific visual assets must count when they are generated, saved, wired into the app, and rendered in the UI element they were created for.

The gate should not force imagegen for ordinary UI chrome that is already covered by a design system, icon library, or user-provided brand assets. A small asset becomes imagegen-required when it is product-specific visual content: it carries brand, reference fidelity, product identity, gameplay, commerce inspection, marketing persuasion, or another concrete visual role that generic UI chrome cannot satisfy.

## Goal

Add an imagegen-based visual asset gate to the NambaAI frontend execution engine so first frontend implementations and first major screen implementations must explicitly decide and prove the asset mode before coding and before a run can be accepted.

## Context

- Project: namba-ai
- Project type: existing
- Language: go
- Mode: tdd
- Work type: plan
- Primary surfaces:
  - `internal/namba/frontend_brief.go`
  - `internal/namba/execution.go`
  - `internal/namba/namba.go`
  - `internal/namba/templates.go`
  - `.agents/skills/namba-run/SKILL.md`
  - `.codex/agents/namba-frontend-implementer.md`
  - `.codex/agents/namba-frontend-implementer.toml`
  - `.codex/agents/namba-designer.md`
  - `.codex/agents/namba-designer.toml`
- Primary tests:
  - `internal/namba/frontend_brief_test.go`
  - `internal/namba/execution_test.go`
  - `internal/namba/spec_review_test.go`
  - `internal/namba/templates_test.go`
  - `internal/namba/spec_command_test.go`
  - `internal/namba/readme_contract_test.go`

## Contract Model

The implementation should make the asset decision explicit before a runner starts implementation. The frontend brief must expose validator-readable decision fields rather than relying only on prose:

- `Frontend implementation phase`: `first-frontend`, `first-major-screen`, `incremental`, or `not-applicable`.
- `Asset mode`: `generated-images`, `existing-assets`, or `not-applicable`.
- `Imagegen requirement`: `required`, `covered-by-existing-assets`, or `not-applicable`.
- `Asset decision proof`: non-pending explanation for any mode that does not require generation.

For `first-frontend` and `first-major-screen`, the default decision is `Asset mode: generated-images` and `Imagegen requirement: required`. The validator may accept `covered-by-existing-assets` or `not-applicable` only when the proof fields satisfy the policy below. `namba run` must enforce Generated Asset Evidence when the pre-run decision says imagegen is required, not merely when a loose label happens to say `generated-images`.

## Repeatable Asset Schema

The Reference-Driven Asset Manifest must use repeatable per-asset blocks so small assets are not hidden behind one aggregate claim. Each asset block should use exact labels:

- `Asset ID`
- `Asset type`
- `Asset source`: `generated`, `existing-input`, or `existing-covered`
- `Required status`: `required`, `supporting`, or `covered`
- `Role in UI`
- `Reference signal`
- `Generation prompt/spec`
- `Source input path`
- `Saved asset path`
- `Intended UI usage`
- `Rendered usage evidence`

Generated assets require non-pending `Generation prompt/spec`, `Saved asset path`, `Intended UI usage`, and `Rendered usage evidence`. Existing-covered assets require non-pending `Source input path`, `Role in UI`, `Intended UI usage`, and proof that the existing asset covers a concrete role that would otherwise trigger imagegen.

The run result section should mirror the manifest with validator-readable per-asset evidence:

```markdown
## Generated Asset Evidence
- Manifest path: ...
- Asset ID: ...
  Generated file: ...
  Saved asset path: ...
  Prompt summary: ...
  Intended UI usage: ...
  Rendered usage evidence: ...
```

The validator should accept multiple `Asset ID` blocks and should fail when any generated asset lacks saved-path or rendered-usage proof.

## Scope

- Update the frontend-major brief scaffold and validator so first frontend implementation and first major screen work starts from `Asset mode: generated-images` by default.
- Define `generated-images` as any imagegen-created visual asset used in the UI, including hero images, section imagery, icons, small object images, product thumbnails, cutouts, textures, sprites, badges, illustrations, and other specific image assets.
- Require imagegen use when a user-provided design system, brand reference, product, place, person, object, marketing page, game, commerce surface, consumer screen, or any other visual-asset-dependent surface determines screen quality and the needed concrete visual roles are not fully covered by existing assets.
- Define mixed-mode usage: if existing assets cover some visual roles and imagegen fills the remaining roles, the brief remains `Asset mode: generated-images` and the manifest must identify which roles are existing inputs and which generated outputs.
- Define the `existing-assets` exception for first frontend work: it is valid only when user-provided assets or design-system assets fully cover every concrete visual role that would otherwise require imagegen, and the brief records proof for those roles.
- Require a Reference-Driven Asset Manifest when imagegen is used. The manifest must record each generated asset's ID, type or role, reference signal, generation prompt or spec, source input if any, output path, intended UI usage, and validation evidence.
- Require Generated Asset Evidence in `namba run` results when imagegen is used or required. Evidence must include manifest path, generated files, prompt summary, saved project asset paths, and rendered usage evidence for each asset's intended UI element.
- Require generated assets to be saved under a project asset path and wired into the implementation. A rendered icon, thumbnail, sprite, or small object image is valid evidence when it is the intended asset; evidence must not be limited to hero or section screenshots.
- Update the run execution prompt, post-run evidence validator, `$namba-run` skill, frontend implementer role card, frontend designer role card, and related tests.
- Rename or redefine `Generation status` so the frontend brief records plan readiness before execution, such as `Generation plan status: ready`, while completed generation remains a run-result evidence responsibility.

## Out Of Scope

- Generating actual visual assets for this SPEC. This SPEC changes the engine, prompts, roles, and validators; it does not implement an end-user screen.
- Replacing the existing Do-Not Design Contract with a separate design workflow.
- Changing non-frontend or `frontend-minor` advisory behavior except where shared wording or evidence summaries need to describe the new asset gate accurately.
- Adding network-dependent image generation tests.

## Constraints

- `Asset mode: not-applicable` is allowed only when generated imagery would harm clarity, introduce noise, or work against the UX, such as pure data tables, internal operations tools, CLI-like admin screens, and text-heavy settings screens. It must include explicit Not-applicable proof and cannot be silently omitted.
- CSS gradients, abstract shapes, empty placeholders, generic SaaS card walls, manually drawn decorative shapes, and token-only styling do not count as generated image substitutes.
- A design whose visual core is photo, product, illustration, icon, object, sprite, or other concrete visual content must not pass by applying only color tokens or CSS decoration.
- Existing-assets mode must not be used as a loophole for first frontend work unless the brief explicitly proves that user-provided or design-system concrete assets cover all needed visual roles and no imagegen generation is required.
- Standard UI chrome, such as common library icons for search, close, save, menu, and pagination, does not require imagegen by itself. Product-specific iconography, object thumbnails, sprites, textures, cutouts, badges, or illustrated markers require imagegen when they carry the screen's concrete visual quality and no sufficient existing asset is provided.
- Validators should provide actionable failure messages that name the missing section, missing proof, invalid substitute, or missing rendered usage evidence.
- Tests must remain deterministic and must not depend on live imagegen calls.
- The brief validator must own the pre-run decision. Runtime validation should enforce evidence against that decision instead of re-deciding whether imagegen was required after implementation output is produced.

## Acceptance Summary

- First frontend and first major screen work defaults to `Asset mode: generated-images` unless the brief records valid Not-applicable proof.
- Visual-asset-dependent frontend work requires imagegen and cannot pass with gradients, abstract decoration, placeholders, generic card walls, or token-only styling.
- `namba run` fails when imagegen is required but Generated Asset Evidence is missing or incomplete.
- Generated Asset Evidence supports both large and small UI assets, including icons, thumbnails, sprites, cutouts, textures, and object images, as long as the asset path and rendered usage evidence prove the intended UI usage.
- Mixed existing/generated asset usage is valid when generated roles stay in `Asset mode: generated-images`, existing roles are documented, and rendered usage evidence proves each generated asset appears in its intended UI element.
- The manifest and run-result evidence are repeatable per asset, so multiple small assets can be validated deterministically instead of relying on the first loose label match.
- Skill and role-card guidance directs designers and implementers to create, save, wire, and report generated assets before the run is accepted.
