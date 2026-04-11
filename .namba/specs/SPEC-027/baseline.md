# SPEC-027 Phase-1 Baseline Evidence

## Canonical Command

Use this source-aligned benchmark command as the phase-1 evidence anchor:

```bash
go test ./internal/namba -run '^$' -bench 'BenchmarkSpec027(LegacyInitDetection|SharedInitScan|RunPreflightSetup|ProjectCommand|SyncCommand)$' -benchmem -count=1
```

Observed on this branch:

- Date: 2026-04-10
- Host: Linux amd64
- CPU: AMD Ryzen 9 7945HX

## Measured Results

| Benchmark | Result |
| --- | --- |
| `BenchmarkSpec027LegacyInitDetection` | `819131 ns/op`, `73545 B/op`, `1565 allocs/op` |
| `BenchmarkSpec027SharedInitScan` | `454099 ns/op`, `41512 B/op`, `1028 allocs/op` |
| `BenchmarkSpec027RunPreflightSetup` | `5879 ns/op`, `4640 B/op`, `47 allocs/op` |
| `BenchmarkSpec027ProjectCommand` | `1000868 ns/op`, `390358 B/op`, `2838 allocs/op` |
| `BenchmarkSpec027SyncCommand` | `2044234 ns/op`, `785339 B/op`, `5471 allocs/op` |

## Hotspot Interpretation

The measured optimization target for phase-1 is the init helper repository-scan path. Against the legacy repeated-walk baseline, the shared scan improved:

- runtime by about `44.6%`
- memory by about `43.6%`
- allocations by about `34.3%`

The broader `project`, `sync`, and `run` setup baselines are recorded here as stability anchors for follow-up slices. Phase-1 was not intended to optimize those paths yet.
