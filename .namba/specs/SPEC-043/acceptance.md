# Acceptance

- [x] `docs/getting-started.md`, `docs/getting-started.ko.md`, `docs/getting-started.ja.md`, and `docs/getting-started.zh.md` use a shared readable structure with language navigation, a compact table of contents, quick-start commands, install/update/uninstall guidance, and clear next-reading links.
- [x] `docs/workflow-guide.md`, `docs/workflow-guide.ko.md`, `docs/workflow-guide.ja.md`, and `docs/workflow-guide.zh.md` use a shared readable structure with command-selection tables, planning and execution flow summaries, review readiness guidance, queue guidance, PR/land flow, generated assets, collaboration defaults, and release flow.
- [x] The four locale sets are meaningfully equivalent: every supported language explains what NambaAI is, which command to use, how to start, how the workflow proceeds, and where deeper references live.
- [x] Badges, status links, card-like sections, tables, and any HTML are GitHub-compatible and improve scanning rather than adding decoration only.
- [x] Existing command names and behavioral claims remain aligned with code, CLI help, `.namba/project` docs, and existing tests.
- [x] README generation ownership is respected: generated README files are not hand-edited unless the owning renderer or config is updated and `namba sync` is run.
- [x] Supporting reference docs remain discoverable from the manual pages, and any changes to `docs/codex-upstream-reference.md` or `docs/moai-adk-codex-migration-analysis.md` preserve their reference or migration-analysis purpose.
- [x] Relative links, language-switch links, anchors, and code fences are checked across changed docs.
- [x] Validation commands pass: `go test ./...`, `gofmt -l "cmd" "internal" "namba_test.go"`, and `go vet ./...`.
- [x] If no automated markdown or link checker is available in the repo, the implementation records the manual link-check method used.

## Implementation Notes

- Source of truth updated: `internal/namba/readme.go`; generated manual outputs refreshed with `go run ./cmd/namba sync`.
- No repository markdown link checker was found. Manual link check used Perl extraction across the eight changed manual files and verified every relative link target exists; a separate `rg` pass inspected tables and code fences.
- Validation passed locally: `go test ./...`, `gofmt -l "cmd" "internal" "namba_test.go"`, and `go vet ./...`.
