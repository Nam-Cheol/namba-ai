# SPEC-057 Eval Plan

## Validation Goals

- Prove CI still runs the existing quality gates.
- Prove new gates are visible and independently actionable.
- Prove local quality command matches the CI quality policy.
- Prove coverage regression is blocked at the documented aggregate threshold.

## Static Checks

- Inspect `.github/workflows/ci.yml` and confirm named gates exist for:
  - Python unittest
  - Go unit tests
  - Harness eval regression
  - gofmt
  - Go vet
  - staticcheck
  - govulncheck
  - Go race tests
  - coverage report
  - coverage threshold
- Inspect `scripts/quality.sh` and confirm it runs or clearly maps the same
  gates.
- Confirm README/docs mention:
  - local quality command
  - CI gate list
  - aggregate coverage threshold and baseline rationale

## Command Validation

- `python3 -m unittest discover -s tests`
- `go test ./...`
- `go test -race ./...`
- `go vet ./...`
- gofmt check over tracked Go files
- `staticcheck ./...`
- `govulncheck ./...`
- `go test ./... -coverprofile=coverage.out`
- `go tool cover -func=coverage.out`
- `scripts/quality.sh`

## Negative Validation

- Temporarily raise the coverage threshold above the measured baseline in a
  local dry run or targeted script test and confirm the threshold check fails
  with measured and expected values.
- Temporarily run the local script in an environment without staticcheck or
  govulncheck and confirm it prints install guidance.
- Do not commit temporary threshold or PATH mutations.

## CI Review Evidence

- GitHub Actions should show distinct failure locations for staticcheck,
  govulncheck, race tests, and coverage threshold.
- Coverage report should be available through artifact upload or job summary.
- Secret scan remains in the existing secret-scan workflow.
- Release workflow diff is empty unless an implementation note explicitly
  justifies a non-semantic change.
