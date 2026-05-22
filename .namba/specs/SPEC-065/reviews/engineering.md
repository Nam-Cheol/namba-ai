# Engineering Review

- Status: clear
- Last Reviewed: 2026-05-22
- Reviewer: `namba-planner`
- Command Skill: `$namba-plan-eng-review`
- Recommended Role: `namba-planner`

## Focus

- Lock architecture, sequencing, failure modes, trust boundaries, and validation strategy before execution starts.

## Findings

- The safest implementation shape is a narrow version-advisory module with injected lookup/cache dependencies, plus small output hooks in existing command handlers.
- `status --json` is a hard contract and should be protected with regression tests before human `status` output changes.
- `runDoctor` currently accepts no arguments through `handleNoArgTopLevelCommand`; implementation must deliberately extend doctor flag parsing for `--check-update` without weakening usage errors for unknown flags.
- `runRegen`, `runProject`, and `runSync` already print human-readable completion messages, which makes them suitable for a centralized advisory append helper.
- `runUpdate` already owns checksum-verified self-update behavior and must remain separate. The advisory resolver should not reuse asset download URLs or update execution paths.
- User-local cache behavior needs dependency injection in tests so package tests do not touch a developer's real cache or the live network.

## Decisions

- Add tests before implementation for resolver states and command output behavior.
- Keep live latest-release lookup behind the explicit refresh path and safe interactive checks only if implementation can prove they are not CI, non-TTY, piped, JSON, or automation-oriented.
- Use official GitHub Release metadata for `Nam-Cheol/namba-ai`; do not use redirected asset download URLs as a metadata source.
- Treat every cache/network/parse failure as advisory data loss, not command failure.
- Keep comparison conservative: only prompt when both installed and latest versions are valid release versions and latest is newer.

## Follow-ups

- Add test hooks or helper constructors to isolate cache paths and release metadata responses.
- Add tests for `doctor --check-update`, unknown doctor flags, default doctor output, human status, and `status --json`.
- Add tests proving `regen`, `project`, and `sync` next-action advisory appears only in behind state.
- Run `go test ./...`, `go vet ./...`, and `scripts/quality.sh`.

## Recommendation

- Cleared for implementation.
