# Acceptance

- [ ] Existing CI behavior remains present: Python unittest discovery,
  `go test ./...`, harness eval regression check, `go vet ./...`, and gofmt
  check.
- [ ] `go test -race ./...` runs in CI as an independently visible gate.
- [ ] `staticcheck ./...` runs in CI as an independently visible gate.
- [ ] `govulncheck ./...` runs in CI as an independently visible gate.
- [ ] CI generates a Go coverage profile and coverage report.
- [ ] CI enforces an aggregate Go coverage threshold.
- [ ] The initial aggregate threshold is documented as 73.0 percent, based on
  the measured 73.7 percent planning baseline.
- [ ] Coverage output is available through a CI artifact or job summary.
- [ ] `scripts/quality.sh` exists and runs the same core checks locally.
- [ ] Local quality output gives actionable setup instructions when
  `staticcheck` or `govulncheck` is missing.
- [ ] Existing `.github/workflows/secret-scan.yml` remains intact.
- [ ] Existing `.github/workflows/release.yml` semantics are not changed.
- [ ] Go module/build cache strategy is present in CI or explicitly justified.
- [ ] Minimal README/docs notes list the local quality command, CI quality
  gates, and coverage threshold policy.
- [ ] No Makefile is created solely for this SPEC; if a Makefile exists by then,
  `make quality` is only a thin wrapper.
- [ ] Validation commands for implementation are recorded in the final run
  evidence: `scripts/quality.sh`, Python unittest discovery, Go tests, race
  tests, vet, gofmt check, staticcheck, govulncheck, coverage generation, and
  coverage threshold check.
