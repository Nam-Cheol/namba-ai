# SPEC-057 Baseline

## Current Workflow Baseline

- `.github/workflows/ci.yml`
  - Python unittest discovery: `python3 -m unittest discover -s tests`
  - Go unit tests: `go test ./...`
  - Harness eval regression check:
    `go run ./cmd/namba eval --suite harness --format json --baseline internal/namba/testdata/evals/harness/baseline.json --fail-on-regression`
  - Go vet: `go vet ./...`
  - gofmt check over tracked Go files
- `.github/workflows/secret-scan.yml`
  - gitleaks action with full checkout history
- `.github/workflows/release.yml`
  - tag/workflow-dispatch matrix build and publish flow
  - out of scope for semantic changes

## Test Structure Baseline

- Go packages from `go list ./...`:
  - `github.com/Nam-Cheol/namba-ai`
  - `github.com/Nam-Cheol/namba-ai/cmd/namba`
  - `github.com/Nam-Cheol/namba-ai/internal/namba`
- Python unittest tests exist under `tests/`.
- No `scripts/` directory exists.
- No `Makefile` exists.

## Coverage Baseline

- Command used:
  `go test ./... -coverprofile=/private/tmp/namba-ai-coverage.out`
- Report command:
  `go tool cover -func=/private/tmp/namba-ai-coverage.out`
- Observed aggregate total: 73.7 percent statements.
- Planned starting threshold: 73.0 percent aggregate total statements.

## Local Tool Baseline

- `staticcheck` was not found in local `PATH`.
- `govulncheck` was not found in local `PATH`.
- The implementation should make both missing-tool cases actionable rather than
  silently skipping the checks.
