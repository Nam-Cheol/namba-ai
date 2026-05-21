# SPEC-060 Contract

## Generated Documentation Contract

| Boundary | Required behavior |
| --- | --- |
| Source of truth | README and guide outputs are generated from renderer/config code, not hand-maintained generated Markdown. |
| Output set | `namba sync` owns root README, localized README files, getting-started guides, and workflow-guide documents enabled by docs config. |
| GitHub-safe rendering | Output may use Markdown links, images, tables, badge images, and `<details><summary>`; it must not emit custom CSS, JavaScript, iframes, style attributes, or unsupported HTML buttons. |
| Internal links | Repository files and assets use relative links where practical. External release or CI links may remain absolute when they point to GitHub-hosted repository services. |
| Localization parity | English, Korean, Japanese, and Chinese outputs keep the same section order, navigation slots, CTA purposes, command-selection rows, and advanced-details positions. |
| Determinism | Renderer output is stable across repeated runs and does not depend on wall-clock time, map iteration, network calls, or local environment-specific ordering. |
| Compatibility | Existing generated-doc coverage stays present while dense reference material may move into guide docs or collapsed advanced sections. |

## Visual Grammar Contract

1. Header: generated-doc comment.
2. Hero: existing configured README hero image when available, with text fallback.
3. Trust: release, CI, security, license, and docs badge or link row.
4. Action: CTA row separate from trust badges.
5. Start: compact quick-start path.
6. Choice: command-selection table for `namba project`, `namba plan`, `namba harness`, `namba fix`, `namba queue`, `namba sync`, and `namba pr`.
7. Depth: workflow tables and advanced details for material that should not dominate the first screen.
8. Navigation: consistent relative cross-document links.

## Non-Goals

- Do not create a website or separate docs app.
- Do not generate new visual assets.
- Do not replace the `namba sync` managed-output contract.
- Do not remove existing command, queue, review, release, CI, or security documentation coverage.
