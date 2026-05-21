# SPEC-058 Eval Plan

## Validation Goals

- Prove the added E2E fixture suite catches workflow integration drift without external services.
- Prove fake Codex and fake GitHub surfaces remain deterministic and close to production command boundaries.
- Prove failure diagnostics point to the broken phase quickly.

## Static Review

- Inspect the new test helper for:
  - independent temp repos
  - no network use
  - no live Codex or live GitHub calls
  - deterministic timestamps and fake SHAs
  - unexpected command rejection
  - cross-platform command handling

## Command Validation

- Focused test target:
  `go test ./internal/namba -run TestE2EWorkflowFixture`
- Full validation:
  `go test ./...`
- Formatting:
  `gofmt -l internal`

## Negative Validation

- Force an unexpected fake command in a local test edit and confirm the failure names the scenario and command.
- Force missing validation evidence in the fixture and confirm the queue or run evidence reports a failed or blocked state.
- Do not commit temporary failure injections.

## CI Evidence

- The new tests run in the existing CI baseline because they are part of `go test ./...`.
- No new secrets, tokens, network configuration, or live smoke environment variables are required.
