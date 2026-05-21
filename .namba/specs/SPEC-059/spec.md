# SPEC-059

## Problem

NambaAI already records useful local workflow signals under `.namba`, but there is no single operator-facing command that turns those files into a health picture for harness execution and evidence quality.

The current state is fragmented:

- `namba status` only prints a shallow repository summary.
- run evidence, validation reports, queue state, review readiness, harness evidence, release notes, and Codex diagnostics live in different `.namba` paths.
- missing or malformed state is easy to overlook until a queue, release, or implementation handoff blocks later.
- automation cannot consume a stable health schema for CI or eval checks.

This keeps NambaAI too close to a task-running CLI. The requested change moves it toward an observable local harness system: a maintainer should be able to ask "what is the health of this Namba workspace?" and see success/failure trends, blockers, missing evidence, stale candidates, readiness, and diagnostics without reading every `.namba` file by hand.

## Goal

Add a read-only observability report surface for NambaAI that aggregates local `.namba` state into both:

- a human-readable report for operators and maintainers; and
- a stable JSON document for CI, eval, and future automation.

## Context

- Project: namba-ai
- Project type: existing
- Language: go
- Mode: tdd
- Work type: plan
- Primary runtime: CLI under `cmd/namba` and `internal/namba`
- Source of truth: local `.namba` files

## Scope

Implement:

- `namba report`
- `namba report --format markdown`
- `namba report --format text`
- `namba report --format json`
- `namba report --json`
- `namba status --json`

Aggregate at least:

- run success and failure counts
- blocked command count
- blocked reason breakdown
- missing evidence count
- validation failure count
- review required count
- review ready count
- queue pending, running, blocked, and done counts
- SPEC completeness summary
- stale SPEC or stale queue item candidates
- release readiness summary
- recent diagnostics evidence summary

Data sources:

- `.namba/logs/runs/*-evidence.json`
- `.namba/logs/runs/*-execution.json`
- `.namba/logs/runs/*-validation.json`
- `.namba/logs/runs/*-validation-attempt-*.json`
- `.namba/logs/runs/*-parallel.json`
- `.namba/logs/runs/*-parallel-progress.jsonl`
- `.namba/logs/runs/*-queue-evidence.json`
- `.namba/logs/runs/*-heartbeat.json`
- `.namba/logs/queue/state.json`
- `.namba/logs/queue/report.md`
- `.namba/specs/SPEC-*`
- `.namba/specs/SPEC-*/reviews/readiness.md`
- `.namba/project/release-checklist.md`
- `.namba/releases/*.md`
- `.namba/logs/project/codex-diagnostics-evidence.json`
- `.namba/manifest.json`

## Non-Goals

- Do not send remote telemetry.
- Do not introduce a SaaS dashboard.
- Do not add a database.
- Do not add a daemon or live watcher.
- Do not integrate with an external observability vendor.
- Do not change queue execution behavior as part of report generation.
- Do not make `namba status` text output incompatible with existing users.

## Proposed Design

### CLI Contract

`namba report` is a read-only command.

```sh
namba report
namba report --format markdown
namba report --format text
namba report --format json
namba report --json
namba report --spec SPEC-059
namba report --since 7d
namba report --fail-on blocked
namba status --json
```

Defaults:

- `namba report` emits markdown-style terminal text.
- `--json` is an alias for `--format json`.
- `namba status` without flags keeps its current text output.
- `namba status --json` emits a compact compatible subset of the report schema.

Exit behavior:

- default report generation exits `0` when a report is produced, even when health is `attention` or `blocked`;
- invalid flags or unsupported formats exit with usage errors;
- `--fail-on attention|blocked|stale` lets CI opt into non-zero exit behavior.

### JSON Contract

Use a stable top-level schema:

```json
{
  "schema_version": "namba-report/v1",
  "generated_at": "2026-05-21T00:00:00+09:00",
  "project": {},
  "summary": {},
  "runs": {},
  "queue": {},
  "specs": {},
  "release": {},
  "diagnostics": {},
  "issues": [],
  "warnings": []
}
```

The schema must be append-only within `namba-report/v1`: existing field names and meanings stay stable, while new optional fields may be added.

### Human Report Contract

The default report should lead with operator decisions:

- health state
- top counts
- top issues
- queue state
- latest run and validation evidence
- review readiness
- stale candidates
- diagnostics
- next recommended action

### Parser Contract

Create tolerant local parsers that never assume every `.namba` file exists or is valid. Each parser returns structured partial data plus warnings instead of panicking or aborting the full report.

Normalize file states as:

- `present`
- `missing`
- `not_applicable`
- `corrupt`
- `unknown`

Normalize health states as:

- `ok`
- `attention`
- `blocked`
- `unknown`

### Safety Contract

- Report generation must not mutate project files.
- Report generation must not invoke network calls.
- Report generation must not run GitHub, Codex, browser, or external CLIs.
- All file paths in output should be workspace-relative where possible.
- Paths escaping `.namba` should be treated as suspicious warnings.

## Current Baseline Observed During Planning

At planning time in this repository:

- `.namba/specs` contains 58 SPEC directories before SPEC-059.
- 44 readiness summaries exist before SPEC-059.
- 32 readiness summaries are ready and 12 require follow-up.
- older `SPEC-001` through `SPEC-014` do not have review bundles.
- `SPEC-029` is missing core SPEC files.
- `.namba/logs/project/codex-diagnostics-evidence.json` exists.
- run evidence manifests are not currently present in `.namba/logs/runs`.
- `.namba/logs/queue/state.json` is not currently present.

## Implementation Notes

- Add report command registration beside existing top-level commands in `internal/namba/namba.go`.
- Keep schema and renderers separate from parsers, similar to the existing eval command shape.
- Prefer small typed structs over generic maps for stable JSON output.
- Reuse existing execution evidence, queue state, validation report, diagnostics, and review parsing types where possible.
- Add parser helpers that can read from an injected filesystem or temporary root so unit tests can build compact fixtures.

## Risks

- Treating missing evidence as hard failure would make a fresh workspace look broken. Missing evidence should be visible but not always fatal.
- Existing historical SPECs may not match modern scaffold expectations. The report should distinguish old missing review bundles from current active blockers.
- `status --json` must not accidentally widen into a second full report schema that drifts from `namba report --json`.

## Handoff

Implementation should start with parser and schema tests, then add renderers and CLI integration. Keep the command read-only throughout the slice.
