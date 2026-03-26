# Acceptance

- [x] `namba init .` offers a guided setup flow in interactive terminals.
- [x] `namba init . --yes ...` supports the same choices through flags.
- [x] Generated scaffold includes `AGENTS.md`, `.agents/skills/`, `.codex/skills/`, `.codex/agents/`, `.codex/config.toml`, and `.namba/config/sections/*.yaml`.
- [x] Generated config stores methodology, language preferences, git mode, and Codex agent mode.
- [x] Secrets are not written into generated files.
- [x] README explains the Codex usage flow in Korean.
- [x] Tests covering the new init scaffold behavior are present.
- [x] Validation commands pass.

Note: checklist synced after rerunning validation in this shell (`go test ./...`, `gofmt -l "cmd" "internal" "namba_test.go"`, and `go vet ./...`).

