# Data Flow

1. The analyzer scopes repository inputs using `.namba/config/sections/analysis.yaml`.
2. System boundaries are inferred before summarization so multi-system repos are not flattened into one tree dump.
3. Evidence-backed findings are rendered into `product.md`, `tech.md`, and per-system docs.
4. Mismatches and thin-output signals are emitted as first-class artifacts.
5. Validation still runs through the configured commands (`go test ./...`, `gofmt -l "cmd" "internal" "namba_test.go"`, `go vet ./...`).

## System State Signals

### workspace

- Generated project state is persisted under `.namba`, including manifests and project documents. Confidence: high. Evidence: `.namba/manifest.json`.

