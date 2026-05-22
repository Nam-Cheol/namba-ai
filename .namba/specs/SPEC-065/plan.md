# SPEC-065 Plan

1. Refresh implementation context before coding:
   - Re-read `internal/namba/namba.go` command routing and output for `doctor`, `status`, `project`, and `sync`.
   - Re-read `internal/namba/update_command.go`, `internal/namba/self_update_command.go`, and `internal/namba/version.go`.
   - Re-read docs and Codex-facing templates that mention `namba update` or `codex update`.
2. Add tests first for the version advisory model:
   - version comparison for behind, current, newer-local, dev, blank, and unparsable installed versions;
   - release metadata parsing for valid latest tag, missing tag, invalid JSON, unexpected tag, prerelease or ignored states if applicable;
   - cache hit, cache miss, cache stale, cache parse failure, cache read failure, and cache write failure;
   - lookup failure and offline behavior remain advisory-only.
3. Add tests for safe execution contexts:
   - no live lookup in default `status`, `status --json`, CI, non-TTY, piped, and automation-oriented flows;
   - explicit refresh through `doctor --check-update` may call the resolver;
   - all lookup/cache failures return success when no unrelated command failure occurs.
4. Add the resolver and cache implementation:
   - use only official GitHub Release metadata for `Nam-Cheol/namba-ai`;
   - store `source`, `repo`, `checked_at`, `latest_version`, and `status` in a user-local cache;
   - enforce a bounded TTL;
   - keep cache and network errors best-effort and non-blocking.
5. Wire advisory state into human-readable command output:
   - `namba doctor` shows neutral or behind version state from safe cache data;
   - `namba doctor --check-update` refreshes metadata and reports the advisory result;
   - `namba status` may show trustworthy cached behind state;
   - `namba status --json` remains compact and unchanged;
   - `namba regen`, `namba project`, and `namba sync` append next-step advisory text only for trustworthy behind state.
6. Centralize wording:
   - say the installed NambaAI CLI appears behind the latest release;
   - recommend asking the user before running `namba update`;
   - state that `namba update` updates only NambaAI, not upstream Codex;
   - avoid prompts for current, dev, unknown, stale, failed, or parse-failed states.
7. Update help, docs, and generated guidance sources:
   - `doctorUsageText`;
   - README and generated docs through `internal/namba/readme.go` when those outputs are managed;
   - `.agents/skills` and `internal/namba/templates.go` guidance that mentions `$namba-update` or `namba update`;
   - tests that lock these documentation contracts.
8. Run focused tests after each slice:
   - version-advisory unit tests;
   - command output tests for doctor/status/regen/project/sync;
   - help and docs contract tests.
9. Run full validation:
   - `go test ./...`
   - `go vet ./...`
   - `scripts/quality.sh`
10. Run `namba sync` after implementation if generated docs, readiness summaries, README bundles, or project support artifacts changed.
