# SPEC-048 Plan

Plan Review must complete before implementation starts.

1. Confirm baseline behavior and source-of-truth helpers:
   - `internal/namba/harness_contract.go`
   - `internal/namba/harness_contract_test.go`
   - `internal/namba/execution_evidence.go`
   - `internal/namba/execution_evidence_test.go`
   - `internal/namba/pr_land_command.go`
   - `internal/namba/pr_land_command_test.go`
   - `.codex/hooks/namba_codex_guard.py`
   - `tests/test_namba_codex_guard.py`
   - `.github/workflows/ci.yml`
2. Add `internal/namba/testdata/evals/harness/` with the required README and
   five JSON fixture files.
3. Define compact fixture schemas in tests and validate every fixture row has a
   name, expected result, and rationale.
4. Implement Go eval tests as the primary path:
   - route cases that assert observable command/request outcomes
   - evidence manifest raw-schema and builder-normalization cases
   - PR review opt-in cases
   - fixture schema validation
   - diagnostic formatting assertions
5. Add or refactor Python hook eval tests only if prompt refinement or guardrail
   cases must execute the real lifecycle hook subprocess.
6. Wire CI only as needed so deterministic eval tests run through existing
   quality commands.
7. Add fixture README guidance:
   - why the eval pack exists
   - how to run it
   - how to add cases
   - what each fixture file measures
   - what must not be added
   - how evals differ from ordinary unit tests
   - why external LLM and API calls are prohibited
8. Validate:
   - `go test ./...`
   - `python3 -m unittest discover -s tests` if Python evals are present
   - `go vet ./...`
   - existing formatting check
9. Manually flip one fixture expected value, confirm a useful failure, and
   revert the change.
10. Search for stale eval fixture references and confirm no runtime logs,
    generated eval reports, or temporary artifacts are committed.
11. Run `namba sync` after implementation validation.
