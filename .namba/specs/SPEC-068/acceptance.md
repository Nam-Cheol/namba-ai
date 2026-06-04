# Acceptance

- [x] `namba init` interactive mode uses Bubble Tea v2 with a model-driven `Init`, `Update`, and `View`.
- [x] `namba init` interactive mode uses Lip Gloss v2 for all TUI visual styling.
- [x] `namba init` interactive mode runs in Bubble Tea AltScreen.
- [x] `namba init` supports up, down, left, right, Enter, Esc, and q navigation.
- [x] `namba init` preserves the current selection flow, defaults, validation, generated files, and manifest behavior.
- [x] `namba init --yes`, CI, and non-TTY flows keep a non-interactive fallback.
- [x] After the interactive TUI exits, stdout prints a plain text summary of created and skipped files.
- [x] `namba update` presents progress and final result state through TUI rendering when interactive, without becoming a wizard.
- [x] `namba update` preserves release download, checksum verification, binary extraction, replacement, Windows scheduling, and error semantics.
- [x] `namba regen` presents progress and final result state through TUI rendering when interactive, without becoming a wizard.
- [x] `namba regen` preserves managed-output ownership rules and session-refresh guidance.
- [x] Domain init logic is separated from TUI rendering logic.
- [x] Tests include Go unit tests for the changed behavior.
- [x] Tests include non-TTY fallback coverage for init, update, and regen.
- [x] Tests include Bubble Tea model update coverage for navigation, confirm, back/escape, and quit behavior.
- [x] Smoke validation covers `namba init`, `namba update`, and `namba regen` using safe fixtures or mocked I/O for unsafe external effects.
- [x] Configured validation passes: `go test ./...`, `gofmt -l "cmd" "internal" "namba_test.go"`, and `go vet ./...`.
