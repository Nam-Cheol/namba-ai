# Engineering Review

- Status: clear
- Last Reviewed: 2026-04-23
- Reviewer: namba-planner
- Command Skill: `$namba-plan-eng-review`
- Recommended Role: `namba-planner`

## Focus

- Lock architecture, sequencing, failure modes, trust boundaries, and validation strategy before execution starts.

## Findings

- The command surface should land as a new top-level `codex` command with an `access` subcommand, following the existing `worktree` pattern. That keeps help routing, error handling, and future Codex admin commands consistent instead of creating a one-off parser shape inside `init`.
- The shared access layer is the architectural center of this slice. It should own preset labels, preview text, raw flag normalization, and precedence rules so `runInitWizard`, `parseInitArgs`, and `namba codex access` cannot drift into separate interpretations of the same access settings.
- Zero-argument `namba codex access` is now a locked inspect-only read path in both TTY and non-TTY contexts. Interactive prompting belongs in `namba init` or explicit flag-driven mutation flows, which keeps the command deterministic and script-safe.
- The post-init mutation path should reuse the managed-output pipeline, but with a narrow output set. Today the access values feed `.namba/config/sections/system.yaml` and `.codex/config.toml`; calling full `runRegen` would still protect non-managed files, but it broadens churn and failure surface without adding correctness for this command.
- Failure handling needs to be explicit at the trust boundary: resolve the effective profile in memory, validate the new access pair before any write, fail outside a Namba repo via `requireProjectRoot`, and treat invalid managed config as a hard stop rather than partially updating source-of-truth config and leaving generated Codex config stale.
- Validation should lock no-op and refresh semantics. If the effective access pair is unchanged, the command should avoid write/manifest churn; if instruction-surface outputs change, it should emit the same session-refresh notice path already used by managed-output regeneration.

## Decisions

- Keep `.namba/config/sections/system.yaml` as the only mutable source of truth for this slice; `.codex/config.toml` remains generated output.
- Implement `namba codex access` as a project-root command surface with read-only help behavior, not as a hidden `init` variant or a direct YAML edit helper.
- Keep zero-argument `namba codex access` inspect-only in both interactive and non-interactive contexts; do not auto-prompt or mutate without explicit mutation flags.
- Keep the mutation path command-local and bounded: normalize and validate inputs first, then regenerate only the affected managed outputs and report refresh guidance.

## Follow-ups

- Add TDD coverage for inspect-only zero-argument behavior across TTY and non-TTY contexts, preset-vs-flag precedence, unchanged-value no-op behavior, invalid-config rejection before mutation, existing-repo safety checks, and preservation of non-managed files.
- Validate implementation with the repo-standard commands: `go test ./...`, `gofmt -l "cmd" "internal" "namba_test.go"`, and `go vet ./...`.

## Recommendation

- Clear with follow-ups. The slice is ready if the command shape, bounded regeneration path, and partial-failure semantics above are locked before implementation starts.
