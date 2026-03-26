# Acceptance

- [x] Codex exec uses `-a` and `-s` instead of `--full-auto`.
- [x] Empty system config values default to `on-request` and `workspace-write`.
- [x] Custom system config values are forwarded to the Codex command line.
- [x] Invalid approval or sandbox values fail before Codex is launched.
- [x] Execution logs record the effective approval and sandbox values.
- [x] `go test ./...` passes.

Note: checklist synced after rerunning validation in this shell (`go test ./...`, `gofmt -l "cmd" "internal" "namba_test.go"`, and `go vet ./...`).

