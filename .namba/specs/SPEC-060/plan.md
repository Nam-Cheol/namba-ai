# SPEC-060 Plan

1. Refresh project context with `namba project` before implementation if project docs are stale.
2. Use the review files under `.namba/specs/SPEC-060/reviews/` to keep product, engineering, design, and readiness concerns visible before `namba run SPEC-060`.
3. Define reusable GitHub-safe Markdown primitives in or near the README renderer:
   - generated-doc header
   - hero image block
   - badge row
   - CTA row
   - cross-document navigation row
   - command-selection table
   - workflow card table
   - quick-start steps
   - advanced `<details><summary>` section
4. Apply the primitives incrementally to the root README renderer first, preserving existing content coverage and current hero-image behavior.
5. Apply the same navigation grammar and advanced-section strategy to `getting-started` and `workflow-guide` renderers.
6. Normalize English, Korean, Japanese, and Chinese output structure so each locale keeps the same section order, navigation slots, CTA purposes, command-selection rows, and advanced-details positions.
7. Extend docs configuration only if it is needed for safe defaults. Any new config must preserve current behavior when omitted.
8. Update existing README and generated-document tests instead of deleting them:
   - helper-level tests for GitHub-safe primitive rendering
   - output tests for hero, badge row, CTA row, command selection, quick start, workflow tables, collapsible advanced content, and cross-document navigation
   - locale parity tests across README and guide outputs
   - relative-link and unsupported-content guards
   - deterministic output tests
9. Run focused validation first, then full validation:
   - `go test ./internal/namba -run 'Readme|Sync|Frontend|Contract'`
   - `go test ./...`
10. Run `namba sync` after implementation so generated README, localized README, getting-started, and workflow-guide files match the renderer.
