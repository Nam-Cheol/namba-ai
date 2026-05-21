# SPEC-060 Baseline

## Current Renderer Baseline

- `internal/namba/readme.go` owns README and guide generation through `buildReadmeOutputs`, `renderReadmeRoot`, and `renderReadmeGuide`.
- `docsConfig` currently supports managed README generation, profile selection, default language, additional languages, and optional `readme_hero_image`.
- `namba sync` replaces managed README and guide outputs through the existing managed-output session.
- Current output already includes generated-doc headers, optional hero image rendering, language links, status links, guide links, command guidance, run modes, queue flow, review readiness, release, CI, and security references.

## Existing Test Baseline

- `internal/namba/readme_sync_test.go` checks renderer sections, localized outputs, guide content, and Namba CLI documentation coverage.
- `internal/namba/readme_contract_test.go` checks checked-in generated documentation drift.
- `internal/namba/sync_test.go` checks sync behavior and readiness refresh interactions.
- `sync_stability_test.go` checks no-op sync stability.
- `namba_test.go` checks CLI-level init and sync generated document presence.

## Current Gaps

- The renderer lacks a reusable visual Markdown primitive layer for badge rows, CTA rows, command-selection tables, workflow-card tables, and advanced details.
- Locale outputs can satisfy content checks while still drifting structurally.
- GitHub-safe allowlist and unsupported-content guards are not explicit enough for the new visual documentation contract.
- Existing first-screen content does not consistently enforce the orientation order: identity, trust, action, command choice, next document.
