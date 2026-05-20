# SPEC-054 Plan

1. Refresh project context with `namba project` and preserve any code-vs-doc
   drift in the project artifacts.
2. Locate the init scaffold sources and generated skill or workflow templates
   that define `namba-pr`, `namba-release`, PR body text, release note text,
   and GitHub Release body handoff behavior.
3. Inspect the concrete code surfaces first: `internal/namba/templates.go`,
   `internal/namba/pr_land_command.go`, `internal/namba/release.go`, and the
   related tests in `namba_test.go`, `internal/namba/templates_test.go`,
   `internal/namba/pr_land_command_test.go`, and
   `internal/namba/release_command_test.go`.
4. Trace the current language-selection path from explicit user request,
   init profile, and project configuration into scaffolded artifacts.
5. Update scaffold templates so new projects require concrete completed
   work, actual changes, evidence sources, and validation results in the
   generated PR body and GitHub release notes.
6. Preserve evidence fields in the generated workflow contract, including
   SPEC ID, PR number, short commit hash, validation evidence, and source
   references when available.
7. Keep `@codex review` absent unless the user explicitly opts in through the
   existing review-request path.
8. Add or update fixture and snapshot tests that initialize a project and
   verify scaffolded PR and release outputs are non-empty, non-generic,
   evidence-bearing, and in the expected language.
9. Run configured validation: `go test ./...`, `gofmt -l "cmd" "internal"
   "namba_test.go"`, and `go vet ./...`.
10. Run `namba sync` after implementation to refresh project artifacts and
   PR-ready documents.
