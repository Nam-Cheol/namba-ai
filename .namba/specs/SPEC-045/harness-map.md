# SPEC-045 Harness Map

## Core Relationship

This SPEC evolves the existing Namba frontend execution harness by making imagegen asset decisions explicit, repeatable, and enforceable for first frontend and first major screen work.

- Existing contract source: `.namba/specs/<SPEC>/frontend-brief.md`
- Existing run surface: `namba run SPEC-XXX`
- Existing template source: `internal/namba/templates.go`
- Delivery model: update contract and validators first, regenerate managed skill/role-card surfaces, then validate with deterministic tests
- Explicit boundary: no live imagegen calls in tests and no new standalone design workflow

## Adaptation Map

| Current surface | SPEC-045 target | Adaptation |
| --- | --- | --- |
| `frontend-brief.md` asset manifest | repeatable per-asset manifest | Add exact-label asset blocks for large and small generated assets. |
| `Asset mode: generated-images` | pre-run imagegen requirement decision | Add `Imagegen requirement` so runtime enforcement follows the brief decision. |
| `Generation status: complete` | `Generation plan status: ready` | Separate pre-run plan readiness from completed generation evidence. |
| `Generated Asset Evidence` | per-asset run evidence | Require saved path, prompt summary, intended UI usage, and rendered usage evidence for each generated asset. |
| `$namba-run` guidance | stricter first-frontend asset gate | Explain defaults, exceptions, mixed mode, and substitute rejection. |
| `namba-designer` role | asset-decision owner | Own imagegen necessity, ordinary UI chrome exceptions, mixed-mode proof, and manifest completeness. |
| `namba-frontend-implementer` role | rendered asset evidence owner | Generate, save, wire, render, and report assets in the intended UI element. |

## Rejection Map

| Candidate shortcut | Namba decision |
| --- | --- |
| CSS gradient as generated image replacement | Reject. Not a generated asset. |
| Abstract decorative shape as asset evidence | Reject unless it is an actual imagegen output with a concrete visual role. |
| Empty placeholder image | Reject. It proves absence, not asset fidelity. |
| Generic SaaS card wall | Reject as visual fallback for asset-dependent surfaces. |
| Token-only brand-color imitation | Reject when the visual core depends on concrete photo, product, illustration, icon, object, sprite, or similar asset content. |
| Existing-assets mode without role coverage proof | Reject for first frontend and first major screen work. |

## Planned Namba Surfaces

- `internal/namba/frontend_brief.go`
- `internal/namba/frontend_brief_test.go`
- `internal/namba/execution.go`
- `internal/namba/execution_test.go`
- `internal/namba/namba.go`
- `internal/namba/templates.go`
- `internal/namba/templates_test.go`
- `internal/namba/spec_command_test.go`
- `internal/namba/spec_review_test.go`
- `.agents/skills/namba-run/SKILL.md`
- `.codex/agents/namba-designer.md`
- `.codex/agents/namba-designer.toml`
- `.codex/agents/namba-frontend-implementer.md`
- `.codex/agents/namba-frontend-implementer.toml`

## Helper Candidate Map

| Candidate | Purpose | Minimum proof |
| --- | --- | --- |
| Repeatable asset-block parser | Parse manifest and Generated Asset Evidence blocks by `Asset ID`. | Fixture tests for multiple assets, missing labels, and small-asset evidence. |
| Imagegen requirement resolver | Decide whether run evidence is required from the frontend brief. | Fixture tests for required, existing-covered, not-applicable, and mixed-mode cases. |
| Substitute detector | Flag banned substitute language in brief or run evidence. | Fixture tests for gradient, placeholder, abstract shape, card wall, and token-only claims. |

Helper candidates should remain small and deterministic. They should not call imagegen, browsers, or network services during unit tests.
