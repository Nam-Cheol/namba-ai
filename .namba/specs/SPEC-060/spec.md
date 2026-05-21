# SPEC-060

## Problem

NambaAI generates README and guide documents through `internal/namba/readme.go` and `namba sync`, but the current generated bundle reads more like a reference dump than a reusable visual documentation system. The renderer already emits important content such as hero imagery, language links, command guidance, and guide links, but the first screen does not consistently answer the new-user questions in a fixed order:

- What is this project and why should I trust it?
- What should I run first?
- Which Namba command fits my situation?
- Which document should I read next?

The change must improve the renderer, not manually rewrite generated files. Future runs of `namba sync` must keep `README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md`, `docs/getting-started*.md`, and `docs/workflow-guide*.md` visually clear, deterministic, and structurally consistent.

## Goal

Upgrade the generated README and guide bundle into a reusable GitHub-safe visual documentation system. The first screen should be optimized for orientation and action while preserving the existing documentation coverage and generated-doc contract.

## Implementation Surface

- Primary renderer: `internal/namba/readme.go`
- Preferred helper extraction: a dedicated renderer or component helper file under `internal/namba/` if it reduces duplication and keeps visual Markdown primitives reusable
- Docs configuration: `docsConfig`, `.namba/config/sections/docs.yaml`, and config loading only if safe defaults are useful for badges, CTA links, card grids, collapsible sections, or visual style toggles
- Sync path: `buildReadmeOutputs`, `renderReadmeRoot`, `renderReadmeGuide`, and the `namba sync` managed-output replacement contract
- Tests: `internal/namba/readme_sync_test.go`, `internal/namba/readme_contract_test.go`, `internal/namba/sync_test.go`, `sync_stability_test.go`, and existing root-level sync coverage in `namba_test.go`

## Visual Documentation Grammar

Generated README and guide documents should share these GitHub-safe primitives:

- Generated-doc header stays first.
- Hero block uses the existing configured README hero image when available. The current GitHub-compatible centered image helper may be reused, but no CSS, JavaScript, iframe, style attributes, or unsupported layout HTML may be introduced.
- Badge row communicates trust signals only: release, CI, security, license, and docs links when they are valid for the repository.
- CTA row is separate from the badge row and uses linked badges or plain GitHub Markdown links, not button-like unsupported HTML.
- Start-here path gives a compact step-by-step route for new users.
- Command-selection section maps user situations to `namba project`, `namba plan`, `namba harness`, `namba fix`, `namba queue`, `namba sync`, and `namba pr`.
- Card-style workflow summaries are implemented as Markdown tables or short structured lists, not CSS cards.
- Advanced or lengthy material moves into guide documents or `<details><summary>` sections. Core new-user information must remain visible outside collapsed blocks.
- Cross-document navigation appears consistently in README, getting-started, workflow-guide, release, CI, and security references using relative links for internal repository targets.

## User Journey Rules

- README owns orientation: project identity, trust signals, first action, command choice, and the next document links.
- `docs/getting-started*.md` owns installation, bootstrap, first run, and the shortest successful path.
- `docs/workflow-guide*.md` owns deeper command semantics, run modes, queue behavior, review readiness, PR and merge flow, and advanced reference material.
- Release, CI, and security links are surfaced as navigation targets without duplicating those documents.
- New users should be able to choose between setup, feature planning, harness planning, bug repair, queue operation, sync, and PR handoff without reading dense reference sections first.

## Localization Parity

English, Korean, Japanese, and Chinese generated outputs must use the same structural grammar:

- Same required section order for each document type.
- Same navigation slots and destinations, adjusted only for relative path depth and localized filenames.
- Same CTA purposes and command-selection rows.
- Same presence or absence of hero, badge, quick-start, workflow table, and advanced-details sections.
- Localized copy may differ in phrasing and examples, but it must not remove required structure.

## Constraints

- Do not manually edit generated README or guide files as the primary solution.
- Keep `.namba/` and the renderer/config as the source of truth.
- Keep output deterministic: no time-dependent content, environment-dependent ordering, network-dependent rendering, or locale-specific structural drift.
- Keep `namba sync` behavior compatible with existing managed-output replacement and no-op sync stability expectations.
- Use relative links for internal repository files and assets wherever practical.
- Preserve existing content coverage, including command skills, custom agents, run modes, queue flow, review readiness, release, CI, and security references.
- Use existing hero assets only; no new image generation is in scope for this SPEC.

## Acceptance Summary

Implementation is complete when `namba sync` regenerates the README and guide bundle with the visual grammar above, tests cover the renderer contract and generated output structure, and `go test ./...` passes.

## Context

- Project: namba-ai
- Project type: existing
- Language: go
- Mode: tdd
- Work type: plan
