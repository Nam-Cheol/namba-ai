# SPEC-068 Plan

1. Add Bubble Tea v2 and Lip Gloss v2 dependencies with the v2 import paths.
2. Extract init domain execution from terminal rendering:
   - keep profile detection, overrides, validation, file writes, manifest writes, and generated content behavior in domain-oriented functions;
   - return a structured write summary that can be rendered by either TUI or fallback output.
3. Implement the `namba init` interactive TUI:
   - define a model with `Init`, `Update`, and `View`;
   - map up/down/left/right/Enter/Esc/q to the existing flow semantics;
   - run the wizard with AltScreen only for interactive terminals.
4. Implement progress/result TUI presentation for `namba update` and `namba regen`:
   - update remains a self-update command, not a wizard;
   - regen remains a managed-output regeneration command, not a wizard.
5. Preserve fallback output for CI, non-TTY, `--yes`, and tests that use buffer-backed stdin/stdout.
6. Add or update targeted tests:
   - Bubble Tea model update tests for navigation, confirm, back/escape, and quit;
   - non-TTY fallback tests for init, update, and regen;
   - regression tests for init generated files and regen ownership semantics;
   - safe update tests with mocked download and executable replacement boundaries.
7. Run configured validation and smoke checks:
   - `go test ./...`
   - `gofmt -l "cmd" "internal" "namba_test.go"`
   - `go vet ./...`
   - smoke `namba init`, `namba update`, and `namba regen` with safe fixtures or mocked I/O.
8. Run `namba sync` after implementation if project docs or PR-ready artifacts need refreshing.
