# SPEC-057 Plan

## 1. Current CI Structure Analysis

- Inspect `.github/workflows/ci.yml`, `.github/workflows/secret-scan.yml`, and
  `.github/workflows/release.yml` before editing.
- Preserve the current `ci.yml` behavior:
  - checkout
  - setup Go from `go.mod`
  - Python unittest discovery
  - `go test ./...`
  - harness eval regression check
  - `go vet ./...`
  - gofmt check
- Keep `.github/workflows/secret-scan.yml` as the existing gitleaks lane.
- Do not change `.github/workflows/release.yml`; release build/publish is not
  part of this SPEC.
- Note the current repo shape in the implementation notes: Go packages are root,
  `cmd/namba`, and `internal/namba`; Python tests live under `tests/`;
  `scripts/` and `Makefile` are absent.

## 2. Quality Gates To Add

- Add `go test -race ./...` as a separate race-detection gate.
- Add `staticcheck ./...` as a separate static-analysis gate.
- Add `govulncheck ./...` as a separate vulnerability gate.
- Keep `go vet ./...` and gofmt as separate named steps.
- Add aggregate Go coverage generation with `go test ./... -coverprofile=coverage.out`.
- Add a coverage report step with `go tool cover -func=coverage.out`.
- Add an aggregate coverage threshold check. Initial threshold: 73.0 percent.
- Keep Python unittest and the existing harness eval regression check in CI.

## 3. Workflow Change Plan

- Split `ci.yml` into clearly named jobs or clearly separated steps so failures
  identify the gate:
  - baseline tests: Python unittest, Go unit tests, harness eval
  - formatting and vet: gofmt, go vet
  - static analysis: staticcheck
  - vulnerability: govulncheck
  - race: `go test -race ./...`
  - coverage: profile, report, threshold, artifact or summary
- Prefer separate jobs for long or independently actionable gates when runtime
  becomes noticeable; otherwise use separate named steps in a small number of
  jobs to avoid unnecessary setup overhead.
- Install staticcheck and govulncheck in CI with pinned or clearly declared
  versions during implementation. Avoid `latest` unless the team accepts
  version drift.
- Ensure each install step is separate from each run step so setup failures are
  distinguishable from analysis failures.
- Keep secret scanning in `secret-scan.yml`; at most reference it in docs.
- Avoid edits to `release.yml`.

## 4. Local Quality Script Design

- Create `scripts/quality.sh` with `set -euo pipefail`.
- Default mode should run the full local quality bar:
  - Python unittest discovery when `tests/` exists
  - gofmt check
  - `go vet ./...`
  - `go test ./...`
  - harness eval regression check
  - `go test -race ./...`
  - `staticcheck ./...`
  - `govulncheck ./...`
  - coverage profile, report, and threshold check
- Use a temp or configurable coverage output path so local runs do not leave
  accidental tracked artifacts.
- Define `COVERAGE_THRESHOLD=73.0` with environment override support.
- If `staticcheck` or `govulncheck` is missing, print the exact `go install`
  command and fail with a clear message.
- If a Makefile is added by other work before implementation, add only a thin
  `quality` target that delegates to `scripts/quality.sh`.

## 5. Coverage Threshold Setup

- Baseline evidence from planning: aggregate Go total is 73.7 percent.
- Start at 73.0 percent to block regressions without requiring immediate test
  expansion.
- Parse the `total:` line from `go tool cover -func=coverage.out`.
- Fail when measured coverage is lower than the threshold.
- Print both values in CI and local output:
  - measured coverage
  - threshold
- Document that package-specific thresholds are deferred to a follow-up SPEC.

## 6. Cache Strategy

- Keep `actions/setup-go` using `go-version-file: go.mod`.
- Add Go module/build cache if it is not already effective:
  - cache `~/go/pkg/mod`
  - cache `~/.cache/go-build`
  - key by runner OS plus `go.mod` hash, with restore keys for the same OS
- Do not cache analysis results in a way that can hide new staticcheck,
  govulncheck, race, or coverage failures.
- Rely on Go module/build cache to speed staticcheck and govulncheck tool
  installation. Cache installed tool binaries only if the implementation pins
  tool versions in the cache key.

## 7. Failure Developer Experience

- Use explicit job or step names:
  - Python unittest
  - Go unit tests
  - Harness eval regression
  - Go race tests
  - Go vet
  - gofmt
  - staticcheck
  - govulncheck
  - coverage threshold
- For gofmt failure, print files that need formatting.
- For coverage failure, print measured coverage, threshold, and coverage report
  location.
- For missing local tools, print install commands and do not continue into a
  confusing partial run.
- Upload `coverage.out` or publish the cover function report in the GitHub job
  summary so reviewers can inspect the regression source.

## 8. Staged Implementation Order

1. Add `scripts/quality.sh` first and run it locally enough to validate command
   order and messaging.
2. Add coverage parsing and threshold logic to the script.
3. Update `ci.yml` to call the same checks or mirror the same command sequence.
4. Add staticcheck and govulncheck installation with pinned or documented
   versions.
5. Add Go cache configuration.
6. Add coverage artifact or summary output.
7. Add minimal README/docs notes for local command, CI gate list, and threshold.
8. If a Makefile appears before implementation, add `make quality` as a thin
   wrapper only.
9. Run validation commands and record results.
10. Run `namba sync` after implementation so project artifacts reflect the new
    quality policy.

## 9. Verification Commands

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
- `namba sync` after implementation

## 10. Risks And Mitigations

- Risk: race tests increase CI runtime. Mitigation: isolate in a named job and
  rely on Go cache; consider making it parallel to baseline tests.
- Risk: staticcheck or govulncheck version drift causes surprise failures.
  Mitigation: pin versions or expose version variables in workflow/script.
- Risk: govulncheck needs vulnerability DB access in CI. Mitigation: keep it as
  a separate job or step with a clear failure name so network/tool issues do not
  obscure unit test status.
- Risk: aggregate coverage varies slightly by Go version or generated files.
  Mitigation: start at 73.0 percent against the measured 73.7 percent baseline
  and document the reason.
- Risk: local script becomes too slow for everyday use. Mitigation: keep the
  default full gate for release/PR confidence and optionally add documented
  flags only if implementation needs them.
- Risk: CI and local script diverge. Mitigation: either call the script from CI
  or keep command order and threshold variables intentionally mirrored.
