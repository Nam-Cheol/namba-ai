# SPEC-057 Strengthen CI quality gates

## Problem

NambaAI currently has a useful baseline CI workflow, but the main CI gate is
still mostly a single happy-path regression lane. Harness reliability can drift
through issues that are not always caught by `go test ./...`: data races,
formatting drift, vet/staticcheck findings, known Go vulnerabilities, and
coverage regressions. Local validation is also spread across separate commands,
so maintainers do not have one command that mirrors the CI quality bar before a
PR.

The current repository evidence is:

- `.github/workflows/ci.yml` runs Python unittest discovery, `go test ./...`,
  the harness eval regression check, `go vet ./...`, and a gofmt check.
- `.github/workflows/secret-scan.yml` already runs gitleaks and should remain a
  separate secret-scan workflow.
- `.github/workflows/release.yml` is tag/workflow-dispatch release packaging and
  publishing; this SPEC must not change release semantics.
- Go module: `github.com/Nam-Cheol/namba-ai`, Go `1.22`, packages at root,
  `cmd/namba`, and `internal/namba`.
- Python unittest tests exist under `tests/` and are already wired into CI.
- No `scripts/` directory exists yet.
- No `Makefile` exists, so this SPEC should not require one.
- Existing Namba quality config lists `go test ./...`, gofmt over `cmd`,
  `internal`, and `namba_test.go`, and `go vet ./...`.
- Aggregate Go coverage measured during planning is 73.7 percent total
  statements using `go test ./... -coverprofile=/private/tmp/namba-ai-coverage.out`
  and `go tool cover -func`.
- `staticcheck` and `govulncheck` were not found in the local `PATH` during
  planning, so the implementation must make local setup failure actionable.

## Goal

Strengthen CI and local quality validation so CI catches harness-trust failures
early while developers can run the same quality bar locally with one command.

## Scope

- Improve `.github/workflows/ci.yml` or equivalent existing CI workflow.
- Add `scripts/quality.sh` as the local quality entry point.
- Add a thin `make quality` target only if a Makefile already exists.
- Add minimal README or docs notes for the local command, CI gate list, and
  coverage threshold.
- Preserve the existing Python unittest behavior.
- Preserve the existing secret-scan workflow.
- Touch `.namba` validation settings only if an existing related setting should
  reference the new quality policy.

## Non-Goals

- Large README rewrites or documentation structure changes.
- Redesigning the `.namba` internal validation system.
- Changing release workflow semantics.
- Creating a Makefile when none exists.
- Package-specific coverage thresholds.
- Broad test refactors or forcing aggregate coverage to 80 percent.
- Paid external services.
- Production deployment pipeline changes.

## Required Quality Gates

- `python3 -m unittest discover -s tests`
- `go test ./...`
- `go test -race ./...`
- Harness eval regression check already present in CI.
- `go vet ./...`
- gofmt check over tracked Go files.
- `staticcheck ./...`
- `govulncheck ./...`
- Go coverage profile and aggregate coverage report.
- Aggregate coverage threshold gate based on the 73.7 percent baseline.

## Coverage Policy

Start with an aggregate Go coverage threshold of 73.0 percent. This is slightly
below the measured 73.7 percent baseline, which gives a small buffer for Go
version/reporting noise while still blocking meaningful coverage regression.
Package-specific thresholds for `internal/namba`, harness runtime, queue,
evidence, or eval packages are intentionally deferred to a later SPEC after
more eval/E2E fixture coverage exists.

## Acceptance

- CI failure causes are clear from job or step names.
- Staticcheck, govulncheck, race test, and coverage threshold failures are
  independently visible.
- Coverage threshold and baseline are documented.
- CI uploads or summarizes the coverage report.
- `scripts/quality.sh` exists and runs the same core validation bar locally.
- Existing Python unittest and secret scan behavior remain intact.
- Release workflow behavior is not changed unnecessarily.
- Go module/build cache strategy is included.
- Implementation validation commands are explicit in the SPEC plan.
