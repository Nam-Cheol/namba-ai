# SPEC-057 Contract

## CI Quality Contract

| Boundary | Required behavior |
| --- | --- |
| Existing tests | CI continues to run Python unittest discovery, `go test ./...`, harness eval regression, `go vet ./...`, and gofmt check. |
| Race detection | CI exposes `go test -race ./...` as an independently named gate. |
| Static analysis | CI exposes `staticcheck ./...` as an independently named gate. |
| Vulnerability check | CI exposes `govulncheck ./...` as an independently named gate. |
| Coverage | CI generates a Go coverage profile, prints a report, and fails below the aggregate threshold. |
| Threshold | Initial aggregate coverage threshold is 73.0 percent, derived from the measured 73.7 percent baseline. |
| Local parity | `scripts/quality.sh` runs the same core checks locally or clearly names any mode split. |
| Secret scanning | Existing secret scan workflow remains intact. |
| Release | Release workflow semantics remain unchanged. |
| Cache | Go module/build cache improves runtime without hiding analysis or coverage failures. |

## Developer Experience Contract

- Each failure must identify the failed gate without requiring log archaeology.
- gofmt failures print the files that need formatting.
- coverage failures print measured coverage, threshold, and report location.
- missing local staticcheck or govulncheck prints exact install guidance.
- coverage artifacts or job summaries are available for PR review.

## Deferred Contract

- Package-specific coverage thresholds for `internal/namba`, harness runtime,
  queue, evidence, and eval surfaces are out of scope for SPEC-057.
- A later SPEC may add package-specific thresholds after stronger eval/E2E
  fixture coverage exists.
