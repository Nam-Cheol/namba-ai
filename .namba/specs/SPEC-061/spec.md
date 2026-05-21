# SPEC-061

## Problem

`namba queue` should keep processing the SPEC list selected by the operator until every selected SPEC is completed or a real implementation, validation, merge-conflict, or explicit operator-control blocker appears. Two reliability failures currently break that contract for initialized NambaAI projects, especially on Windows:

- Queue-owned state and evidence writes can appear in `git status --porcelain` and trip branch cleanliness checks even when the files were produced by queue itself.
- Namba's Codex runner appends the full execution prompt to the `codex exec` argv. Long SPEC requests can hit Windows command-line length limits, especially because Windows execution wraps Codex through `cmd /c codex ...`.

A third continuity gap appears when remote PR handoff is unavailable after implementation and validation have succeeded. The current queue path assumes push, PR creation, check polling, and PR merge. If remote handoff cannot be reached, queue needs an explicit local fallback path instead of stopping with an opaque PR blocker.

## Goal

Implement the smallest safe queue-continuity fix so initialized NambaAI projects can process a queued SPEC list across Windows and non-Windows environments without:

- treating queue-owned metadata as user dirty work,
- passing large Codex execution prompts through argv, or
- stopping on remote PR handoff unavailability when a local, evidence-backed fallback can finish the active SPEC safely.

## Context

- Project: namba-ai
- Project type: existing
- Language: go
- Mode: tdd
- Work type: fix

## Scope

In scope:

- Namba CLI queue conveyor behavior in `internal/namba/queue_command.go`.
- Queue branch-clean checks for queue-owned state and run evidence.
- Codex execution request transport in the execution layer used by queue, run, direct fix, and repair turns.
- Queue fallback handling when remote PR handoff is unavailable after the active SPEC has valid implementation and validation evidence.
- Targeted unit tests and the existing workflow fixture or equivalent e2e-style queue coverage.

Out of scope:

- A queue UX redesign or a new public queue mode unless tests prove the existing command surface cannot express the fix.
- Broadening global dirty-worktree ignores for `plan`, `pr`, `land`, or `release`.
- Treating failed validation, failed checks, negative review state, merge conflicts, or explicit pause/stop requests as continuity cases.
- Changing the queue state schema unless a narrowly tested additive field is required.

## Required Behavior

### Queue-owned cleanliness

- Queue may ignore only queue-owned runtime artifacts for the active queue operation.
- The allowlist is queue scoped, not global. A safe implementation should avoid widening the current global `hasWorkingTreeChanges` behavior. Non-queue clean checks must continue to report normal dirty state.
- Queue-owned artifacts include queue state/report files and active-run artifacts produced for the active SPEC, such as request, heartbeat, execution, validation, queue evidence, preflight, evidence manifest, and stdout/stderr stream files.
- Unrelated user edits, mixed user edits plus queue-owned files, modified SPEC docs, source files, config files, or arbitrary files under run logs for a non-active SPEC must still block when a clean branch is required.
- Operator-facing messages should distinguish queue-owned writes from user edits when a mixed dirty state blocks the queue.

### Codex request transport

- The full execution prompt must not be appended to `codex exec` or `codex exec resume` argv.
- Prefer stdin transport by passing `-` as the prompt argument and sending the request body through stdin. Local Codex help supports stdin for `codex exec` when prompt is omitted or `-`, and for `codex exec resume` when prompt is `-`.
- If stdin cannot be used by a supported runner path, use a temp request file and keep argv bounded to flags plus a small file reference.
- Preserve the existing capability-based flag/config resolution for approval policy, sandbox, model, profile, web search, add-dir, JSON output, and resume mode.
- Persist the existing request JSON and markdown artifacts so run evidence remains inspectable.

### Remote handoff fallback

- Remote PR handoff remains the preferred path.
- The local fallback is allowed only after implementation, sync, and validation evidence have succeeded for the active SPEC.
- Use fallback only for remote handoff unavailability, such as missing/unusable `gh`, authentication/network failure, push failure, or PR create/load failure. Do not use it for validation failure, failed checks, negative mergeability, review-required states, or local merge conflicts.
- The fallback must be explicit in queue state, report output, and durable evidence. It must state that the queue used a local branch/local base-branch merge fallback.
- The fallback should commit the active SPEC branch if needed, merge that branch into the configured local base branch, mark the SPEC as landed with local fallback evidence, and continue to the next selected SPEC.
- If the local merge cannot complete cleanly, queue must block with a concrete recovery action rather than hiding the conflict.

## Validation Strategy

- Add failing unit tests first for queue-scoped dirty classification, including queue-owned-only, user-only, and mixed dirty states.
- Add failing execution-layer tests proving long prompts are transported outside argv for both `codex exec` and `codex exec resume`, including a Windows command-line length simulation.
- Add queue state tests for remote-handoff-unavailable local fallback and for fallback conflict blocking.
- Add e2e-style coverage through the existing queue workflow fixture or an equivalent queue regression fixture. "Queue dry-run" in this SPEC means fixture-driven or `namba run --dry-run`-backed regression coverage that avoids launching real Codex/GitHub services; it does not require adding a new public `namba queue --dry-run` flag.
- Run configured validation: `go test ./...`, `gofmt -l "cmd" "internal" "namba_test.go"`, and `go vet ./...`.
