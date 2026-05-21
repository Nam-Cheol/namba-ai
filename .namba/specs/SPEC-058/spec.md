# SPEC-058

## Problem

NambaAI already has focused unit tests for individual CLI commands and an opt-in live Codex smoke test. It does not yet have a default CI-safe E2E fixture that proves the real CLI workflow still works when a user moves from bootstrap through planning, execution evidence, queue state, and PR or land handoff simulation.

The missing coverage creates a blind spot: command-level tests can pass while the end-to-end orchestration contract breaks across `init`, `project`, `plan`, `harness`, `run`, `queue`, `status`, `pr`, and `land`.

## Goal

Add deterministic Go E2E fixture coverage for the NambaAI CLI workflow without calling live Codex, live GitHub, external APIs, or the network.

## Context

- Project: namba-ai
- Project type: existing
- Language: go
- Mode: tdd
- Work type: plan

## Scope

- Create Go test fixtures that build independent temporary repositories with `t.TempDir`.
- Exercise production command handlers through `NewApp(...).Run(ctx, args)`.
- Replace only external process boundaries with fake command runner or adapter seams.
- Cover user-visible flow across:
  - `namba init`
  - `namba project`
  - `namba plan`
  - `namba harness`
  - `namba run` or dry run
  - `namba queue start`, `queue status`, and resume or blocked-state transitions
  - `namba status`
  - PR creation simulation
  - land or release readiness simulation
  - run and queue evidence file checks
  - blocked or unsafe scenario checks

## Out Of Scope

- Live Codex execution.
- Live GitHub API calls.
- Network access.
- Real GitHub tokens.
- Real PR creation.
- Real release publish.
- Flaky dependency on host git remotes or external services.

## Required Scenarios

1. Happy path: a fresh temp repo completes the main workflow and produces expected project docs, SPEC artifacts, run evidence, queue state, simulated PR metadata, and readiness output.
2. Blocked unsafe path: a dirty or unsafe planning or queue state is blocked with durable, diagnosable state and without accidentally continuing.
3. Missing evidence or validation failure path: missing or failed validation evidence is reported through structured run or queue evidence with clear recovery context.

## Implementation Notes

- Prefer `internal/namba/e2e_workflow_fixture_test.go` so tests can use package-private seams already present on `App`.
- Use a small fake external runner that records command invocations and rejects unexpected `git`, `gh`, `codex`, shell, or network-like calls.
- Keep assertions mostly JSON and field based; use focused stdout substring assertions only where the CLI user contract is textual.
- Normalize volatile fields such as temp paths, timestamps, and fake SHAs in helper assertions.
