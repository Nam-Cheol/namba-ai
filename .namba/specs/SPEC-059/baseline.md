# SPEC-059 Baseline

## Repository Baseline

Observed during planning on 2026-05-21:

- 58 SPEC directories existed before SPEC-059.
- 44 readiness summaries existed before SPEC-059.
- 32 readiness summaries were review-ready.
- 12 readiness summaries required follow-up.
- `SPEC-001` through `SPEC-014` had no review bundles.
- `SPEC-029` was missing core SPEC files.
- 11 harness request files existed before SPEC-059.
- no run evidence manifests were present under `.namba/logs/runs`.
- no active queue state existed at `.namba/logs/queue/state.json`.
- project Codex diagnostics evidence existed under `.namba/logs/project/codex-diagnostics-evidence.json`.

## Existing Code Baseline

- `namba status` is registered in `internal/namba/namba.go` and currently accepts no arguments.
- `namba eval` already has separate JSON and markdown rendering paths that can guide report renderer shape.
- execution evidence manifests are represented by `executionEvidenceManifest`.
- queue state is represented by `queueState`.
- validation output is represented by `validationReport`.
- diagnostics evidence is represented by `codexDiagnosticsEvidence`.

## Behavioral Baseline

The new report must improve visibility without changing:

- queue execution semantics;
- run execution semantics;
- release execution semantics;
- existing `namba status` text output;
- existing `namba eval` output.
