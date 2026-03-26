# Acceptance

- [x] README is rewritten in Korean for OSS users and includes emoji section headers.
- [x] README includes a direct `go install` command for NambaAI.
- [x] The Go module path matches the GitHub repository path used in README installation examples.
- [x] Windows CLI startup forces UTF-8 console output before running the app.
- [x] The UTF-8 setup path is covered by tests.
- [x] Validation commands pass.

Note: checklist synced after rerunning validation in this shell (`go test ./...`, `gofmt -l "cmd" "internal" "namba_test.go"`, and `go vet ./...`).

