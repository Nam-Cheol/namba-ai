# Engineering Review

- Status: clear
- Last Reviewed: 2026-05-20
- Reviewer: Codex local review
- Command Skill: `$namba-plan-eng-review`
- Recommended Role: `namba-planner`

## Focus

- Lock architecture, sequencing, failure modes, trust boundaries, and validation strategy before execution starts.

## Findings

- The proposed implementation is conservative: add a read-only top-level command, productionize existing deterministic fixture helpers, and keep public command behavior unchanged.
- The main engineering risk is accidentally preserving test-only logic where expected fixture labels choose the actual evaluation path. The SPEC now calls this out explicitly.
- `go:embed` is the right default-corpus strategy because installed binaries should run `namba eval` outside a source checkout.
- Exit codes, schema validation, baseline comparison, and renderer tests are concrete enough for implementation.

## Decisions

- Add new eval files under `internal/namba/` instead of folding the command into the already-large `namba.go`.
- Keep fixture files under `internal/namba/testdata/evals/harness/` and add embedded defaults.
- Require `go test ./...` as minimum validation.

## Follow-ups

- Implementation should verify the new command is read-only with a temp project test.
- CI should use the built binary or `go run ./cmd/namba` consistently to avoid path differences.

## Recommendation

- Clear for implementation.
