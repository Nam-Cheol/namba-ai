# SPEC-060 Eval Plan

## Validation Goals

- Prove generated README and guide outputs use the new reusable GitHub-safe visual documentation grammar.
- Prove localized README and guide outputs keep equivalent structure.
- Prove `namba sync` remains deterministic and managed-output compatible.
- Prove existing documentation coverage is preserved while dense material is reorganized into guide or advanced sections.

## Focused Checks

- Renderer helper tests for badge rows, CTA rows, command-selection tables, workflow-card tables, advanced details, and cross-document navigation.
- Output tests for hero, badge row, CTA row, command selection, quick-start path, card-style workflow or navigation section, collapsible advanced section, and cross-document navigation.
- Locale parity tests for English, Korean, Japanese, and Chinese README and guide outputs.
- Link tests that internal repository targets and assets use relative paths where practical.
- Safety tests that output does not contain custom CSS, JavaScript, iframe tags, style attributes, or unsupported HTML buttons.
- Determinism tests that repeated `buildReadmeOutputs` or `namba sync` output is stable.

## Command Validation

- Focused test target:
  `go test ./internal/namba -run 'Readme|Sync|Frontend|Contract'`
- Full validation:
  `go test ./...`
- Formatting check:
  `gofmt -l cmd internal namba_test.go`
- Typecheck:
  `go vet ./...`

## Negative Validation

- Temporarily inject an unsupported tag into a renderer helper and confirm the safety test fails.
- Temporarily reorder a localized section and confirm the locale parity test fails.
- Temporarily switch an internal guide link to an absolute repository URL and confirm the relative-link test fails.
- Do not commit temporary failure injections.
