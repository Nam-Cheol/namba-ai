# SPEC-033

## Goal

Define a typed execution-evidence contract for NambaAI's core harness so plan and readiness surfaces can reference real runtime proof without treating browser tooling or runtime observability as universal required evidence.

## Command Choice Rationale

- This work changes Namba-owned runtime, harness validator, readiness, and generated guidance surfaces.
- Under the typed harness contract introduced by `SPEC-032`, that is `core_harness_change` and must route through `namba plan`.
- This is not a reusable domain harness package and not direct artifact generation.

## Context

- Project: namba-ai
- Project type: existing
- Language: go
- Mode: tdd
- Work type: plan

## Verified Local Context

- `SPEC-020` established `namba harness` as the planning surface for reusable harness work and required explicit trigger/eval strategy, but it did not define a post-run execution-evidence contract.
- `SPEC-016` expanded runtime observability and durable execution metadata for standalone runs.
- `SPEC-028` introduced a machine-readable JSONL event stream for parallel progress under `.namba/logs/runs/`.
- `SPEC-032` introduced typed harness classification, `harness-request.json`, and an advisory harness evidence contract centered on `contract.md`, `baseline.md`, `eval-plan.md`, and optional `harness-map.md`.
- `internal/namba/execution.go` already writes structured run artifacts such as request, preflight, execution, and validation JSON.
- `internal/namba/parallel_progress.go` already writes structured progress events for parallel runs.
- `internal/namba/spec_review.go` and `internal/namba/harness_contract.go` currently surface harness evidence mainly by validating the presence of required human-readable artifacts.
- `.namba/config/sections/codex.yaml` and generated `.codex/config.toml` already opt into Playwright MCP, but current guidance treats browser verification as relevant only when the task actually needs it.

## Problem

NambaAI now has two useful but disconnected layers:

1. planning-side harness evidence
   - typed harness classification exists
   - advisory harness readiness exists
   - required evidence files exist
2. runtime-side structured evidence
   - run request/preflight/execution/validation artifacts exist
   - parallel progress events exist

The missing piece is a stable typed bridge between them.

That gap creates four problems:

1. plan and readiness surfaces can confirm that a harness package is documented, but they cannot reference a canonical execution-proof artifact after a run completes
2. runtime artifacts exist, but consumers must know artifact-specific file names and semantics instead of following one contract
3. browser validation artifacts such as screenshots, traces, or DOM captures have no canonical optional slot in the core harness contract
4. current advisory readiness is pre-implementation, while execution proof is post-run; the repository does not yet define how those two layers relate without collapsing them into one ambiguous status

## Desired Outcome

- NambaAI defines one typed execution-evidence manifest that indexes the runtime artifacts produced by the existing run pipeline.
- The manifest lives alongside existing run artifacts and references them instead of duplicating their content.
- The manifest covers, at minimum:
  - request
  - preflight
  - execution
  - validation
  - progress when the run mode emits it
- Browser evidence is modeled as an optional extension:
  - present only when browser verification is actually relevant and artifacts exist
  - absent or `not_applicable` for non-browser work
- Runtime observability is modeled as an optional typed extension that reuses existing structured artifacts and selected signal bundles rather than redefining the `SPEC-028` event contract
- Plan readiness stays pre-implementation and advisory
- Execution proof becomes a separate post-run advisory layer that is surfaced through one explicit v1 operator-facing consumer path: `namba sync` and sync-generated summaries
- `reviews/readiness.md` stays plan-only and does not become the primary post-run proof surface
- `SPEC-032` routing and harness-classification semantics remain unchanged
- The operator-facing labels remain distinct:
  - `readiness` answers "can we implement?"
  - `execution proof` answers "what run evidence exists?"

## Scope

- Define a typed execution-evidence manifest contract for runs that already emit structured runtime artifacts.
- Persist that manifest under `.namba/logs/runs/` using the existing run/log identity model.
- Index existing runtime artifacts rather than inventing a second full payload format.
- Define per-artifact status semantics such as `present`, `missing`, and `not_applicable`.
- Add optional extension slots for:
  - browser evidence references
  - selected runtime/observability evidence references
- Ensure the manifest can represent:
  - default/solo/team-style runs that produce request/preflight/execution/validation artifacts
  - parallel runs that also produce progress JSONL
- Surface the latest execution-proof status through `namba sync` and generated summaries separately from pre-implementation readiness so the two phases are visible without being conflated.
- Define one manifest ownership/finalization rule so the canonical execution-evidence artifact still exists on important failure paths such as:
  - preflight failure
  - execution failure before validation
  - validation failure after retries
- Update generated guidance and stable docs where the new contract becomes user-visible.
- Add regression coverage for manifest generation, optional-extension behavior, and compatibility with older runs/specs that lack the new manifest.

## Non-Goals

- Do not reopen or redesign the `SPEC-032` routing/classification foundation.
- Do not claim that runtime observability is missing as a foundation problem; `SPEC-028` already establishes that layer for parallel progress.
- Do not make Playwright or browser artifacts mandatory for every harness or run.
- Do not make browser tooling a repo-level hard dependency for non-browser work.
- Do not redesign `namba run` mode semantics.
- Do not silently convert advisory readiness into execution-time hard gating.
- Do not require external SaaS observability backends, remote Codex app features, or user-local Codex settings to satisfy the contract.

## Design Constraints

- Keep `.namba/` as the source of truth for repository-owned planning state and runtime artifacts.
- Reuse current structured runtime artifacts as the authoritative evidence source; the manifest is an index and interpretation layer, not a replacement.
- Keep pre-implementation readiness and post-run execution evidence as separate phases with explicit linkage.
- Make browser evidence capability-scoped and optional.
- Keep optional browser/runtime extensions limited to typed, repo-owned artifacts or explicit signal bundles; do not allow arbitrary free-form attachment paths to define proof semantics.
- Preserve compatibility for existing SPECs and existing run artifacts that do not yet emit the new manifest.
- Keep the schema generic enough that future evidence producers can attach browser, UI, or richer runtime proofs without breaking the base contract.
