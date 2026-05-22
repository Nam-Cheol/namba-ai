# SPEC-065

## Problem

NambaAI currently has a deliberate `namba update` command, but Codex-facing maintenance commands do not tell an operator when the installed `namba` CLI appears behind the latest NambaAI release. That creates two bad failure modes:

- Humans may keep using stale workflow scaffolding because no maintenance command gives a lightweight version drift hint.
- Codex or automation could be tempted to treat update discovery as an action and run `namba update` without explicit user consent.

This SPEC adds advisory-only version drift messaging for human-readable maintenance surfaces. It must never auto-run updates, must never fail a maintenance command because update metadata is unavailable, and must preserve JSON and automation-oriented output contracts.

## Goal

Add non-blocking NambaAI version advisory messaging for installed `namba` CLI drift in Codex-facing maintenance commands, while keeping `namba update` an explicit user-approved command.

## Current Evidence

- `internal/namba/namba.go` registers `doctor`, `status`, `project`, `regen`, `sync`, and `update` as top-level commands.
- `runDoctor` prints local repo and toolchain readiness, including the detected Namba CLI path, but does not currently compare the installed CLI with published releases.
- `runStatus` supports human-readable text plus `--json`; `status --json` is a compact script-oriented contract backed by `renderStatusJSON`.
- `runProject`, `runRegen`, and `runSync` print human-readable completion and next-action style messages after local maintenance work.
- `internal/namba/self_update_command.go` owns `namba update`, release asset URLs, checksum verification, and the `updateRepo` constant set to `Nam-Cheol/namba-ai`.
- `internal/namba/version.go` exposes `Version()` and uses `dev` when no release version is injected.
- `internal/namba/readme.go`, `internal/namba/templates.go`, `README.md`, and generated docs already explain that `namba update` updates only NambaAI and is distinct from upstream `codex update`.

## Scope

- Add a cached latest-release resolver for the official NambaAI GitHub release metadata source:
  - repository: `Nam-Cheol/namba-ai`
  - source label: official GitHub Release metadata
  - expected lookup target: the GitHub latest release metadata endpoint or an equivalent official GitHub release metadata URL, not release asset redirects
- Store a user-local cache record with at least:
  - `source`
  - `repo`
  - `checked_at`
  - `latest_version`
  - `status`
- Use a bounded TTL for cache freshness and keep stale or invalid cache states advisory-only.
- Add an explicit refresh surface, preferably `namba doctor --check-update`, that may perform a live lookup when the user explicitly requests it.
- Surface advisory-only version state in human-readable `namba doctor` and `namba status`.
- Add next-step advisory text to human-readable `namba regen`, `namba project`, and `namba sync` when cached or explicitly refreshed data proves the installed CLI is behind.
- Update Codex-facing guidance, generated docs, and tests so Codex asks the user before running `namba update`.
- Add targeted tests for resolver behavior, cache states, text output, JSON stability, and maintenance next actions.

## Non-Goals

- Do not auto-run `namba update` from any command.
- Do not add a background updater, daemon, telemetry, or new analytics call.
- Do not call non-GitHub services for latest-release discovery.
- Do not change installer checksum verification or self-update replacement mechanics.
- Do not merge `namba update` with upstream `codex update`.
- Do not change `status --json`, report JSON, CI evidence schemas, or other stable machine-readable contracts unless a future SPEC explicitly scopes that.

## Required Behavior

- Behind state:
  - When installed `Version()` is a parseable release version and valid latest-release data shows a newer release, human-readable intended surfaces show a clear advisory.
  - The advisory recommends asking the user before running `namba update`.
  - The advisory names `namba update` as the NambaAI CLI update path and does not imply Codex itself will be updated.
- Up-to-date state:
  - When installed and latest versions are equal, no update prompt is emitted on `regen`, `project`, or `sync`.
  - `doctor` and human `status` may show neutral version state, but must not prompt an update.
- Dev-build state:
  - `Version()` equal to `dev`, blank, unparsable, or locally annotated build versions must not emit update prompts.
  - Dev state may show a neutral advisory in `doctor --check-update`, but must not ask the user to update from `dev`.
- Offline, cache-miss, stale-cache, lookup-failure, parse-failure, and cache-read or cache-write failures:
  - Commands continue successfully unless they would have failed for an existing unrelated reason.
  - Default human-readable maintenance commands stay quiet or neutral; they do not produce noisy warnings when no trustworthy behind state exists.
  - Explicit refresh may report the lookup failure as advisory text while exiting successfully if the only failure is update metadata retrieval.
- CI, non-TTY, piped, JSON, and automation-oriented flows:
  - No live network version check runs unless the user explicitly requests it with the refresh surface.
  - Script-oriented output remains stable and noise-free.
  - `namba status --json` preserves its current compact contract and does not include advisory chatter.

## Cache Contract

- Store the cache in a user-local cache location, not under the repository's `.namba` state.
- Prefer the platform-appropriate Go cache directory API and a Namba-specific subdirectory.
- Cache writes must be best-effort: inability to create, read, parse, or write the cache never fails the invoking command.
- The cache status should distinguish enough states for tests and output decisions, such as `hit`, `miss`, `stale`, `lookup_failed`, `parse_failed`, `dev`, `behind`, or `current`.
- The TTL must be bounded and documented in code or tests.
- Stale cache data must not trigger behind prompts unless the user explicitly requested refresh and the refresh succeeds.

## Output Contract

- `namba doctor`:
  - Default human-readable output may use fresh cache data when available and safe.
  - `namba doctor --check-update` explicitly refreshes latest-release metadata and reports advisory state without failing on lookup, parse, or cache errors.
- `namba status`:
  - Human-readable output may show cached advisory state when trustworthy behind data is available.
  - `namba status --json` remains unchanged and must not perform a live lookup.
- `namba regen`, `namba project`, and `namba sync`:
  - When trustworthy behind data is available, append a next-step advisory that says to ask the user before running `namba update`.
  - Do not emit update prompts for current, dev, unknown, stale, failed, JSON, CI, non-TTY, or automation-safe flows.
- Help and docs:
  - `doctor` help documents the explicit refresh option.
  - Codex-facing docs and skills tell Codex to ask the user before `namba update`.
  - Docs keep `namba update` distinct from upstream `codex update`.

## Implementation Notes

- Reuse the existing `updateRepo` constant or move the repository identity into a shared small version-advisory module to avoid duplicate repo strings.
- Prefer a dedicated file such as `internal/namba/version_advisory.go` with dependency-injected lookup/cache helpers so tests do not use the live network or real user cache.
- Reuse `App.downloadURL` or add a narrow metadata fetcher wrapper so tests can stub release metadata responses.
- Parse semantic release tags conservatively. Treat invalid, missing, pre-release, or unexpected tag data as no-prompt advisory states.
- Keep all network calls out of package tests by stubbing the fetcher.
- Keep advisory formatting centralized so doctor/status/regen/project/sync cannot drift in wording.

## Validation

- `go test ./...`
- `go vet ./...`
- `scripts/quality.sh`
