# SPEC-045 Contract

## Invariants

- Extend the existing `frontend-major` asset gate; do not introduce a second frontend artifact outside `.namba/specs/<SPEC>/frontend-brief.md`.
- Keep the frontend brief as the canonical pre-run source for `Frontend implementation phase`, `Asset mode`, `Imagegen requirement`, and `Asset decision proof`.
- Enforce Generated Asset Evidence from the pre-run `Imagegen requirement: required` decision, not from post-hoc output prose alone.
- Treat `generated-images` as any imagegen-created UI asset, including icons, thumbnails, cutouts, textures, sprites, badges, object images, illustrations, hero media, and section imagery.
- Do not force imagegen for ordinary UI chrome already covered by a design system, icon library, or user-provided assets.
- Keep mixed existing/generated usage in `Asset mode: generated-images` whenever any uncovered concrete visual role still requires generation.
- Require `Asset mode: existing-assets` proof when user-provided or design-system assets fully cover all concrete visual roles.
- Require `Asset mode: not-applicable` proof when generated imagery would harm clarity, introduce noise, or work against the UX.
- Preserve advisory behavior for non-frontend and `frontend-minor` work.
- Keep tests deterministic and free of live imagegen or network dependency.

## Required Decision Fields

Every first frontend or first major screen brief must expose:

| Field | Required behavior |
| --- | --- |
| Frontend implementation phase | Distinguishes `first-frontend`, `first-major-screen`, `incremental`, and `not-applicable`. |
| Asset mode | Supports `generated-images`, `existing-assets`, and `not-applicable`. |
| Imagegen requirement | Supports `required`, `covered-by-existing-assets`, and `not-applicable`. |
| Asset decision proof | Required when imagegen is not required. |
| Generation plan status | Uses readiness language such as `ready`; completed generation belongs in run evidence. |

## Per-Asset Manifest Schema

The Reference-Driven Asset Manifest must support repeatable asset blocks with exact labels:

- Asset ID
- Asset type
- Asset source
- Required status
- Role in UI
- Reference signal
- Generation prompt/spec
- Source input path
- Saved asset path
- Intended UI usage
- Rendered usage evidence

Generated assets require non-pending `Generation prompt/spec`, `Saved asset path`, `Intended UI usage`, and `Rendered usage evidence`.

Existing-covered assets require non-pending `Source input path`, `Role in UI`, `Intended UI usage`, and proof that the asset covers a concrete visual role that would otherwise require imagegen.

## Run Evidence Schema

`namba run` output must include `## Generated Asset Evidence` when imagegen is required. The section must support repeatable per-asset blocks with:

- Manifest path
- Asset ID
- Generated file
- Saved asset path
- Prompt summary
- Intended UI usage
- Rendered usage evidence

The evidence validator must fail when any generated asset lacks saved-path or rendered-usage proof.

## Substitute Rejection

The gate must reject these as generated image substitutes:

- CSS gradients
- abstract shapes
- empty placeholders
- generic SaaS card walls
- manually drawn decorative shapes
- token-only styling

## Blocking Rules

- A first frontend or first major screen brief without a valid asset decision blocks execution.
- A required imagegen decision without Generated Asset Evidence blocks post-run validation.
- Missing per-asset saved path or rendered usage evidence blocks post-run validation.
- Existing-assets mode blocks unless all concrete visual roles are proven covered.
- Not-applicable mode blocks unless the proof explains why generated imagery would harm clarity, introduce noise, or work against the UX.
