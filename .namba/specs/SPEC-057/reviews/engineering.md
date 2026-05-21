# Engineering Review

- Status: clear
- Last Reviewed: 2026-05-21
- Reviewer: namba-planner
- Command Skill: `$namba-plan-eng-review`
- Recommended Role: `namba-planner`

## Focus

- Lock architecture, sequencing, failure modes, trust boundaries, and validation strategy before execution starts.

## Findings

- Current CI already has a strong baseline path, so implementation should extend
  rather than replace it.
- Staticcheck, govulncheck, race, and coverage should be independently named in
  CI. Separate jobs are preferred if runtime grows, but separate named steps are
  acceptable if setup overhead would dominate.
- `staticcheck` and `govulncheck` were not installed locally during planning.
  The local script must fail with install instructions instead of silently
  skipping these checks.
- Coverage parsing should use the `total:` line from `go tool cover -func` and
  compare numerically against 73.0 percent.
- Because `govulncheck` may depend on external vulnerability database access in
  CI, its failure must be distinguishable from unit test failure.
- Go caching should cover module and build caches, but not cache analysis
  results in a way that hides failures.

## Decisions

- Use `scripts/quality.sh` as the local source of truth for the full quality
  sequence.
- Keep coverage artifacts outside tracked source by default for local runs.
- Do not edit `release.yml` except if implementation proves a no-op formatting
  change is unavoidable, which should be avoided.
- Pin or explicitly declare staticcheck and govulncheck versions in CI and docs.

## Follow-ups

- During implementation, decide whether CI calls `scripts/quality.sh` directly
  or mirrors its steps. Direct calls maximize parity; mirrored steps improve
  GitHub Actions failure grouping.
- If CI calls the script directly, keep critical subcommands clearly grouped in
  logs or add script section banners.

## Recommendation

- Clear for implementation.
