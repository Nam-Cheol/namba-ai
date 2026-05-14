# SPEC-045 Plan

1. Map the current frontend asset gate.
   - Read `internal/namba/frontend_brief.go`, `internal/namba/execution.go`, `internal/namba/namba.go`, and `internal/namba/templates.go`.
   - Confirm how `frontend-major`, asset modes, negative-first contract issues, readiness summaries, execution prompts, and post-run evidence validators currently interact.
   - Identify generated skill and role-card ownership so source templates and generated `.agents` or `.codex` artifacts stay in sync.

2. Tighten the frontend brief contract.
   - Add explicit decision fields: `Frontend implementation phase`, `Asset mode`, `Imagegen requirement`, and `Asset decision proof`.
   - Make first frontend implementation and first major screen work default to `Asset mode: generated-images`.
   - Add policy language that `generated-images` includes any imagegen-created UI asset: hero or section imagery, icons, small object images, product thumbnails, cutouts, textures, sprites, badges, illustrations, and similar concrete assets.
   - Distinguish ordinary existing UI chrome from product-specific small assets. Common design-system icons do not require imagegen by themselves; product-specific icons, thumbnails, sprites, cutouts, textures, badges, or object images do when they carry visual identity or reference fidelity.
   - Define mixed-mode handling: screens may use existing assets and generated assets together, but any generated-required role keeps the mode on `generated-images` and must be manifest-backed.
   - Define the first-frontend `existing-assets` exception: existing assets are sufficient only when the brief proves they fully cover every concrete visual role that would otherwise require imagegen.
   - Require Reference-Driven Asset Manifest entries to use repeatable per-asset blocks with exact labels: Asset ID, Asset type, Asset source, Required status, Role in UI, Reference signal, Generation prompt/spec, Source input path, Saved asset path, Intended UI usage, and Rendered usage evidence.
   - Preserve `not-applicable` only with explicit proof that imagery would harm the UX.
   - Rename or redefine `Generation status` as pre-run plan readiness, such as `Generation plan status: ready`, so the brief does not claim assets are already generated before `namba run`.

3. Strengthen run prompting and evidence validation.
   - Update the run execution prompt to require imagegen review before first frontend coding and to require generated asset work before final layout when the policy calls for it.
   - Update pre-run validation so imagegen-required decisions are derived from the brief contract, then update post-run validation so that decision fails without `## Generated Asset Evidence`.
   - Require `## Generated Asset Evidence` to mirror the manifest with repeatable per-asset labels: Asset ID, Generated file, Saved asset path, Prompt summary, Intended UI usage, and Rendered usage evidence.
   - Require evidence for saved asset paths and rendered usage in the intended UI element, not only hero or section usage.
   - Accept validator-readable rendered usage proof for small assets, including screenshot callouts, component render output, DOM/file-path linkage, storybook or preview evidence, or test-render evidence tied to the intended UI element.
   - Reject gradients, abstract shapes, empty placeholders, generic SaaS card walls, manually drawn decoration, and token-only styling as substitutes.

4. Update skills and role cards.
   - Update `$namba-run` guidance with the default generated-images gate, not-applicable proof, and Generated Asset Evidence requirements.
   - Update `namba-designer` guidance so design review owns imagegen necessity, manifest completeness, not-applicable proof, and substitute rejection.
   - Update `namba-frontend-implementer` guidance so implementers generate or integrate imagegen assets, save them under project asset paths, render them in the intended UI elements, and report evidence.
   - Regenerate templated assets if the authoritative templates own the role cards or skill text.

5. Add deterministic regression tests.
   - Cover first frontend default `generated-images` behavior.
   - Cover explicit `Frontend implementation phase` and `Imagegen requirement` parsing.
   - Cover valid and invalid `not-applicable` proof.
   - Cover imagegen-required briefs failing without Generated Asset Evidence.
   - Cover Generated Asset Evidence for small assets such as icons, sprites, thumbnails, or object images.
   - Cover multi-asset manifest and multi-asset run evidence parsing.
   - Cover missing per-asset saved path and missing per-asset rendered usage evidence failures.
   - Cover existing-assets proof when all visual roles are covered, and mixed existing/generated proof when generation fills only some roles.
   - Cover rejected substitutes: CSS gradients, abstract shapes, placeholders, generic card walls, and token-only styling.
   - Cover updated prompt and template text for skill and role guidance.

6. Validate and sync.
   - Run `go test ./...`.
   - Run `gofmt -l "cmd" "internal" "namba_test.go"` and ensure it prints no changed Go files.
   - Run `go vet ./...`.
   - Run `namba sync` if generated docs, role cards, skills, or project artifacts are affected.
   - Inspect the diff for generated ownership comments and unintended churn.
