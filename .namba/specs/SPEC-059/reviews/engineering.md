# Engineering Review

- Status: clear
- Last Reviewed: 2026-05-21
- Reviewer: Codex local engineering review
- Command Skill: `$namba-plan-eng-review`
- Recommended Role: `namba-planner`

## Focus

- Lock architecture, sequencing, failure modes, trust boundaries, and validation strategy before execution starts.

## Findings

- The implementation should be built as a read-only collector plus renderers, not as behavior embedded directly in command handlers.
- Existing structs provide strong anchors: `executionEvidenceManifest`, `validationReport`, `queueState`, queue runner evidence, and Codex diagnostics evidence.
- Parser failure handling is the key engineering risk. Every parser should return partial data and warnings instead of aborting the whole report.
- `status --json` should reuse the same collector and emit a compact subset, preventing two divergent status/report schemas.
- The command must not run GitHub, Codex, browser automation, or network calls; diagnostics should summarize only already-recorded evidence.
- Deterministic JSON is important for tests and CI. Use typed structs and stable slice ordering for issues, specs, blockers, and recent items.

## Decisions

- Add new report schema, collector, parser, and renderer files under `internal/namba`.
- Register `report` in the public top-level command list.
- Extend `runStatus` argument parsing only for `--json`; preserve existing no-arg behavior and existing invalid extra-arg behavior.
- Represent file and evidence states as `present`, `missing`, `not_applicable`, `corrupt`, or `unknown`.
- Represent report health as `ok`, `attention`, `blocked`, or `unknown`.
- Use fixture-driven unit tests with temporary `.namba` roots before wiring broad CLI output tests.

## Follow-ups

- Keep stale heuristics conservative: stale queue heartbeat candidates are strong; old incomplete SPECs should be `stale_candidate` or `attention`, not hard blocked by default.
- Ensure `--since` parsing is small and deterministic; if it creates scope creep, land it after the core schema but keep it in the command contract only if implemented and tested.
- Consider using the existing eval renderer style as a reference, but do not couple report JSON to eval result JSON.

## Recommendation

- Proceed to implementation. Engineering boundaries, parser behavior, and validation strategy are concrete enough for `namba run SPEC-059`.
