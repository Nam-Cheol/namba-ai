# SPEC-033 Execution-Evidence Contract

## Purpose

Define the typed execution-evidence layer that bridges NambaAI's existing planning/readiness contract with its existing structured runtime artifacts.

The code and tests remain authoritative. This document is the human-readable contract anchor.

## Phase Separation

NambaAI must keep two different advisory layers explicit:

1. planning/readiness evidence
   - pre-implementation
   - centered on `spec.md`, `plan.md`, `acceptance.md`, reviews, and harness evidence files
2. execution proof
   - post-run
   - centered on runtime artifacts under `.namba/logs/runs/`

This SPEC does not collapse those phases into one status.

## Primary Artifact

- Canonical artifact: `.namba/logs/runs/<log-id>-evidence.json`
- This manifest is an index over runtime artifacts, not a replacement payload format.
- The manifest should be emitted only by runtime paths that already produce structured run artifacts.

## Minimum Shape

The manifest must include:

- stable run identity
- `spec_id` when the run is tied to a SPEC
- generation timestamp
- execution mode or equivalent run-shape metadata
- typed references to:
  - request artifact
  - preflight artifact
  - execution artifact
  - validation artifact
- explicit state per reference:
  - `present`
  - `missing`
  - `not_applicable`

## Optional Extensions

### Browser Evidence

- Browser evidence is optional.
- Valid examples include screenshots, traces, DOM captures, or other browser-validation artifacts produced by relevant flows.
- Browser evidence must not be required for non-browser work.
- When browser verification is not relevant, the manifest should say so explicitly instead of implying failure.
- Browser references must use typed, repo-owned artifact paths rather than arbitrary free-form attachment conventions.

### Runtime Evidence

- Runtime/observability evidence is also optional and typed.
- Reuse existing structured artifacts and selected signal bundles where possible.
- Do not replace the `SPEC-028` progress event contract.
- Do not treat free-form logs as the only evidence contract.
- Runtime references must stay bounded to typed artifacts or explicit signal bundles that consumers can validate consistently.

## Consumer Rules

- `namba run` emits the manifest when the execution path supports it, and the implementation must define one canonical emission/finalization rule that covers important failure paths as well as success paths.
- `namba sync` and sync-generated summaries are the primary v1 operator-facing consumer path for execution proof.
- `reviews/readiness.md` remains plan-only and must not quietly become the primary post-run proof surface.
- The manifest is advisory by default in v1.
- Missing execution evidence must be visible when relevant, but v1 must not silently convert it into a universal hard gate.

## Invariants

- Keep `SPEC-032` route classification unchanged.
- Keep the harness evidence pack for planning/readiness unchanged.
- Keep browser evidence optional and capability-scoped.
- Keep legacy SPECs and historical runs compatible when the new manifest is absent.
