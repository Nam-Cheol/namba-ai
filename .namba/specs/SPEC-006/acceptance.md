# Acceptance

- [x] `namba init .` creates `AGENTS.md`, `.agents/skills/`, `.codex/skills/`, and `.namba/` Codex-ready state.
- [x] Repo-local skill `namba` exists and documents how Namba commands are interpreted inside Codex.
- [x] `AGENTS.md` tells Codex to treat `namba run SPEC-XXX` as Codex-native in-session execution.
- [x] `namba doctor` reports Codex-native repository readiness.
- [x] README explains that `namba init .` makes the project usable from Codex.
- [x] README explains the supported status line customization path.
- [x] Validation commands pass.

Note: checklist synced after rerunning validation in this shell (`go test ./...`, `gofmt -l "cmd" "internal" "namba_test.go"`, and `go vet ./...`).

