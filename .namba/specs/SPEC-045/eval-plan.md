# SPEC-045 Evaluation Plan

## Evaluation Goals

- Prove that first frontend and first major screen work defaults to `Asset mode: generated-images` and `Imagegen requirement: required`.
- Prove that Generated Asset Evidence is enforced from the pre-run imagegen requirement decision.
- Prove that small generated assets are validated as first-class assets through repeatable per-asset blocks.
- Prove that existing-assets and not-applicable paths require explicit proof and cannot silently bypass imagegen.
- Prove that gradients, abstract shapes, placeholders, generic SaaS card walls, manually drawn decoration, and token-only styling do not count as generated image substitutes.
- Prove that generation plan readiness is separate from completed generation evidence.

## Regression Tests

- Scaffold and brief generation:
  - first frontend work seeds explicit decision fields
  - first major screen work starts with generated-images defaults
  - non-frontend and `frontend-minor` paths remain advisory

- Parser and readiness:
  - valid `Imagegen requirement: required` is detected before run
  - `covered-by-existing-assets` requires proof for all concrete visual roles
  - `not-applicable` requires proof that generated imagery would harm clarity, introduce noise, or work against UX
  - `Generation plan status: ready` is accepted as pre-run readiness
  - `Generation status: complete` is no longer required before run

- Per-asset manifest:
  - multiple generated assets parse deterministically
  - small assets such as product-specific icons, sprites, thumbnails, textures, cutouts, badges, and object images require saved path and rendered usage evidence
  - mixed existing/generated manifests distinguish existing inputs from generated outputs

- Execution:
  - missing `## Generated Asset Evidence` fails when imagegen is required
  - missing per-asset `Saved asset path` fails
  - missing per-asset `Rendered usage evidence` fails
  - small-asset evidence passes with component render, DOM/file-path linkage, preview, test-render, or screenshot proof

- Templates and generated surfaces:
  - `$namba-run` explains the generated-images default, proof exceptions, mixed-mode path, and per-asset evidence schema
  - `namba-designer` owns imagegen necessity, not-applicable proof, and manifest completeness
  - `namba-frontend-implementer` generates, saves, wires, renders, and reports assets in the intended UI elements

## Fixture Scenarios

- Marketing first screen with product-specific icon set and object thumbnails: generated-images required.
- Commerce product grid with existing brand photos plus generated comparison badges: mixed mode, still generated-images.
- Internal data table with standard sort/filter icons: not-applicable or existing-assets proof allowed.
- Game setup screen with generated sprites and cutouts: generated-images required with small-asset evidence.
- Text-heavy settings screen where imagery would add noise: not-applicable allowed only with proof.
- Runner output with generated file path but no rendered usage evidence: failed.

## Validation Commands

Run after implementation:

```bash
namba regen
namba sync
gofmt -l "cmd" "internal" "namba_test.go"
go test ./...
go vet ./...
git diff --check
```

## Manual Review Checks

- Verify validator errors name the missing field or asset ID.
- Verify managed `.agents` and `.codex` role-card outputs match template sources.
- Verify README or workflow guide changes, if any, explain the policy without bloating operator docs.
- Verify generated-images examples include both large and small asset use cases.
