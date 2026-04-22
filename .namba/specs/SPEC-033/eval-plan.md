# SPEC-033 Eval Plan

## Goal

Define the regression shape that proves the new execution-evidence contract is working without reopening the existing harness-classification foundation.

## Fixture Matrix

| Fixture | Expected execution-evidence result |
| --- | --- |
| Standard SPEC run with request/preflight/execution/validation artifacts | manifest exists with base artifact references and no required browser extension |
| Preflight failure before execution starts | manifest still exists and marks preflight as `present` while later stages are `missing` or `not_applicable` according to the contract |
| Execution failure before validation starts | manifest still exists and marks execution as `present` while validation is `missing` or `not_applicable` according to the contract |
| Validation failure after retries | manifest still exists and records validation evidence plus failure-oriented terminal status without disappearing |
| Parallel SPEC run | manifest exists and references the existing progress JSONL artifact |
| Non-browser run | browser extension is `not_applicable` or absent, not failed |
| Browser-verified run with canonical browser artifacts | manifest includes typed browser references |
| Legacy run created before this contract | consumers degrade safely when the manifest is absent |

## Validator Expectations

- The manifest must not duplicate full runtime payloads that already exist in separate artifacts.
- Base artifact references must include explicit state, not implied truthiness.
- The canonical manifest must still be emitted on important failure paths, not only on success.
- Browser evidence must never be required globally.
- Parallel progress linkage must reuse `SPEC-028` artifacts instead of redefining them.
- Consumers that surface execution evidence must keep it distinct from pre-implementation readiness.

## Regression Coverage

- Manifest generation tests for standard runtime paths.
- Failure-path tests for preflight failure, execution failure before validation, and validation failure after retries.
- Parallel manifest tests proving progress linkage is optional and mode-aware.
- Browser-extension tests for `present` vs `not_applicable`.
- Compatibility tests for historical runs/specs without the new manifest.
- Consumer tests proving readiness/sync summaries do not silently collapse planning readiness and execution evidence into one state.

## Exit Signal

This SPEC is ready to execute when the planned implementation can prove one thing clearly:

NambaAI can point to a canonical, machine-readable execution-evidence artifact without pretending browser verification is universal and without reopening `SPEC-032` classification semantics.
