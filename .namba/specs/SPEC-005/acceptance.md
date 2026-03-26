# Acceptance

- [x] Users can install NambaAI without a local Go toolchain.
- [x] The default README install path uses release installer scripts, not `go install`.
- [x] The documented command form is the global `namba` command.
- [x] Windows installer downloads a release archive and registers the install directory on user PATH.
- [x] macOS/Linux installer downloads a release archive and registers the install directory on PATH.
- [x] A GitHub Actions release workflow packages supported platform binaries.
- [x] Release asset naming is covered by tests.
- [x] Validation commands pass.

Note: checklist synced after rerunning validation in this shell (`go test ./...`, `gofmt -l "cmd" "internal" "namba_test.go"`, and `go vet ./...`).

