# SPEC-056 Plan

1. Refresh current project context with `namba project` if implementation starts from stale docs.
2. Extract reusable harness eval types and helpers from `internal/namba/harness_eval_test.go` into production eval files.
3. Add `namba eval` as a public read-only top-level command with help text and option parsing.
4. Define unified scenario, result, metric, and baseline JSON structs.
5. Convert the current harness eval fixture families into a minimum 24-scenario v1 corpus.
6. Add baseline comparison with clear regression diagnostics and stable exit codes.
7. Add Markdown and JSON renderers.
8. Update existing harness eval tests to call the same engine used by the CLI.
9. Add command-level tests for help, read-only behavior, invalid schema handling, result rendering, and regression failure.
10. Add CI invocation for `namba eval --suite harness --format json --baseline internal/namba/testdata/evals/harness/baseline.json --fail-on-regression`.
11. Update fixture README and user-facing help docs only where needed for the new command.
12. Run validation: `go test ./...`; recommended follow-up: `go vet ./...`, gofmt check, and the `namba eval` regression command.
