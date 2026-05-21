# SPEC-058 Contract

## E2E Fixture Contract

| Boundary | Required behavior |
| --- | --- |
| Repository isolation | Every E2E scenario creates its own temp repository and does not depend on the developer's real workspace state. |
| CLI path | Tests call `App.Run` for the relevant `namba` commands so command parsing, root discovery, scaffold writing, and workflow state transitions stay on the production path. |
| External commands | Live `codex`, `gh`, network, token, PR creation, release publish, and remote operations are forbidden in the default test suite. |
| Fake runner | Fake external command handling is deterministic, records calls, and fails on unexpected command shapes with clear scenario context. |
| Evidence | Run, validation, queue, and PR or land simulation artifacts are checked through structured JSON or targeted file assertions. |
| Diagnostics | Failure messages identify the scenario, command phase, expected artifact, and relevant field or command mismatch. |
| Cross-platform | Tests avoid POSIX-only behavior unless routed through existing cross-platform shell command helpers. |
| CI | The suite runs under the normal `go test ./...` baseline without extra environment variables. |

## Scenario Contract

- Happy path proves a new temp repo can move through the representative Namba workflow using fake Codex and fake GitHub surfaces.
- Blocked unsafe path proves dirty or unsafe state blocks before irreversible workflow progress.
- Missing evidence or validation failure path proves absent or failed evidence is visible in durable state and user-facing diagnostics.

## Non-Goals

- Do not create real GitHub pull requests.
- Do not publish releases.
- Do not call live Codex.
- Do not require network access.
- Do not replace focused unit tests with a brittle monolithic golden test.
