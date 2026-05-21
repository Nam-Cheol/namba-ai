# SPEC-058 Plan

1. Add a reusable E2E fixture helper under `internal/namba`.
   - Create isolated temp repos with minimal source files.
   - Initialize Namba through the real `init` command.
   - Provide deterministic stdout, stderr, clock, environment, and cwd handling.
2. Add a fake external command runner.
   - Model only the `git`, `gh`, `codex`, and shell calls the workflow needs.
   - Record calls for diagnostics.
   - Fail immediately on unexpected external calls.
   - Keep fake responses close to current production command shapes.
3. Implement the happy-path E2E scenario.
   - Run `init`, `project`, `plan`, `harness`, `run` or dry run, `queue`, `status`, `pr`, and `land` simulation through `App.Run`.
   - Assert project docs, SPEC artifacts, run evidence, queue state, and simulated PR metadata.
4. Implement the blocked unsafe scenario.
   - Simulate dirty workspace or unsafe branch state.
   - Assert the workflow blocks without continuing and writes or prints diagnosable recovery context.
5. Implement the missing evidence or validation failure scenario.
   - Simulate failed validation or missing run evidence.
   - Assert structured evidence status, queue blocker, and recovery action.
6. Add focused assertion helpers.
   - Require file existence with path-rich failure messages.
   - Decode JSON and compare specific fields.
   - Normalize volatile path, time, and SHA values where needed.
7. Validate locally.
   - Run focused test target for the new E2E suite.
   - Run `go test ./...`.
   - Run gofmt on touched Go files.
