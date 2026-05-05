# SPEC-042 Plan

1. Refresh and lock planning context.
   - Run `namba project` if implementation starts after meaningful code drift.
   - Re-read `.namba/project/product.md`, `.namba/project/tech.md`, mismatch and quality reports, and the command implementation files touched by this SPEC.
   - Run the product, engineering, and design review tracks under `.namba/specs/SPEC-042/reviews/`.

2. Add the public queue command contract.
   - Register `namba queue` in `publicTopLevelCommandDefinitions()`.
   - Add usage text and help tests for `queue start`, `queue status`, `queue resume`, `queue pause`, and `queue stop`.
   - Implement argument parsing for inclusive ranges and explicit SPEC lists.
   - Reject target SPECs that do not already exist.
   - Define queue-specific semantics for `--auto-land`, `--skip-codex-review`, cooperative pause, stop, and no-auto-land waiting.

3. Build durable queue state.
   - Add queue state structs and atomic read/write helpers under `internal/namba`.
   - Store runtime state and reports under `.namba/logs/queue/`.
   - Separate queue-level state from per-SPEC phase state, including expected branch, last observed head SHA, current run log id, pause/stop request flags, and last safe checkpoint.
   - Add branch resolution for existing SPECs and block on multiple `spec/<SPEC-ID>-*` matches.
   - Add lock or active-state protection so only one queue and one active SPEC can run.
   - Cover corrupted, missing, stale, stopped, paused, and blocked state files with tests.

4. Implement status detection and resume checkpoints.
   - Add helpers that derive per-SPEC phase from local SPEC files, review readiness, run evidence, validation reports, branch state, PR state, checks, and merge state.
   - Recompute external truth before irreversible actions.
   - Persist a checkpoint before and after every phase.
   - Add tests for idempotent resume after interruption at review, run, PR, checks, land, and main refresh.

5. Wire review and implementation phases.
   - Run the three review tracks in parallel where the configured Codex runner or skill path supports it.
   - Refresh readiness once, serially, after all review tracks finish; block if required review evidence is missing or unclear.
   - Treat `clear-with-followups` as passable only when every follow-up bullet is tagged `[non-blocking]` or `[post-implementation]`.
   - Invoke existing team implementation behavior equivalent to `namba run --team SPEC-XXX`.
   - Reuse existing validation and repair loops instead of adding a second validator.

6. Wire PR, checks, and land phases.
   - Reuse existing PR creation and PR lookup behavior where it can be made active-SPEC-aware; otherwise add queue-specific helpers that share the same safety gates.
   - Add a queue-scoped way to skip creating a new `@codex review` marker comment without changing global config.
   - Detect existing PRs and existing review marker comments without duplication.
   - Choose and implement the required-check proof strategy before wiring auto-land: prefer explicit GitHub required-check data, and use the stricter "all surfaced checks must be green" fallback only when that fallback is recorded in queue evidence and covered by tests.
   - Inspect checks, draft state, base branch, mergeability, and review or merge-state ambiguity; block if the selected check evidence strategy cannot produce a trustworthy result.
   - Stop at a waiting or blocked state unless `--auto-land` is explicitly present and every gate passes.
   - Reuse land behavior to merge and refresh local `main`.

7. Deliver operator reports.
   - Add concise terminal output for start, status, pause, stop, resume, skip, blocked, and done outcomes.
   - Include active SPEC, operator-facing state, internal detail, queue progress, PR URL, blocker or wait reason, evidence path, and next safe command.
   - Keep default `status` compact by summarizing completed/skipped counts and writing full detail to the report artifact.
   - Ensure blocked reports distinguish validation failure, failed checks, non-mergeable PR, dirty Git state, and ambiguous GitHub state.
   - Ensure `pause_requested`, `paused`, `stopped`, `waiting_for_checks`, `ready_to_land`, and `waiting_for_land` are not conflated in copy.

8. Regression coverage and validation.
   - Add unit tests with fake Git, GitHub, runner, and validator command hooks; do not require live GitHub for normal tests.
   - Add command-level tests for help, parsing, state persistence, skip semantics, blocked semantics, and resume semantics.
   - Run `go test ./...`.
   - Run `gofmt -l "cmd" "internal" "namba_test.go"`.
   - Run `go vet ./...`.
   - Run `namba sync` if generated docs, README surfaces, or project artifacts changed.
