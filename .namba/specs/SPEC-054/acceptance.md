# Acceptance

- [x] Newly initialized projects receive scaffolded PR workflow templates or
  skill contracts that require concrete completed work in `namba pr` bodies.
- [x] Newly initialized projects receive scaffolded release workflow templates
  or skill contracts that require GitHub release notes to describe actual
  changes, evidence sources, and validation results.
- [x] GitHub Release bodies produced through the scaffolded release workflow
  cannot be empty or only generic placeholder text.
- [x] Generated artifacts preserve available change evidence such as SPEC ID,
  PR number, short commit hash, validation evidence, and source references.
- [x] Output language follows an explicit user language request when present;
  otherwise it follows the language stored by `namba init` or project
  configuration.
- [x] `@codex review` is included only when the existing explicit opt-in review
  path is used.
- [x] Fixture or snapshot tests cover the newly scaffolded PR body and release
  notes behavior.
- [x] Regression tests cover `buildPullRequestBody`, scaffolded
  `namba-pr`/`namba-release` skill rendering, and release note handoff behavior
  at the narrowest practical layer.
- [x] Validation commands pass: `go test ./...`, `gofmt -l "cmd" "internal"
  "namba_test.go"`, and `go vet ./...`.
