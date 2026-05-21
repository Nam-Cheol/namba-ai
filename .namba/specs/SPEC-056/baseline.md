# SPEC-056 Baseline Evidence

## Existing Deterministic Eval Surface

The repository already contains a small deterministic harness eval pack:

- `internal/namba/harness_eval_test.go`
- `internal/namba/testdata/evals/harness/route_cases.json`
- `internal/namba/testdata/evals/harness/mention_plugin_cases.json`
- `internal/namba/testdata/evals/harness/prompt_refinement_cases.json`
- `internal/namba/testdata/evals/harness/guardrail_cases.json`
- `internal/namba/testdata/evals/harness/evidence_manifest_cases.json`
- `internal/namba/testdata/evals/harness/pr_review_cases.json`

The current corpus has 39 fixture rows across those files.

## Existing Strengths

- Fixtures are local and deterministic.
- Existing tests already avoid LLM, GitHub, Codex CLI, network, and user-local configuration.
- Cases cover routing, mention ambiguity, prompt refinement, dangerous command guardrails, evidence manifests, and PR review opt-in.
- Harness request metadata already has typed fields in `internal/namba/harness_contract.go`.

## Existing Gaps

- There is no public `namba eval` command.
- Fixture schemas are split by file and not represented as one operator-facing scenario schema.
- JSON and Markdown eval reports do not exist.
- Baseline comparison does not exist.
- CI does not run an explicit harness eval regression command.
- Some test helpers choose evaluation paths from expected fixture categories, which is useful for tests but not acceptable for a public evaluator.

## Public Behavior To Preserve

- `namba plan` remains the feature planning command.
- `namba harness` remains the harness-oriented SPEC command.
- `$namba-create` remains the direct artifact creation path.
- `namba run` semantics and execution modes remain unchanged.
- `namba pr --review` remains explicit opt-in only.

## Validation Baseline

Configured validation from project docs:

- test: `go test ./...`
- lint: `gofmt -l "cmd" "internal" "namba_test.go"`
- typecheck: `go vet ./...`
