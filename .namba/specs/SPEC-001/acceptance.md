# Acceptance

- [x] `namba run` uses a runner abstraction instead of calling Codex directly from the orchestration path.
- [x] Serial execution writes `request.md`, `result.txt`, `execution.json`, and `validation.json` logs when execution succeeds.
- [x] Runner failures write `result.txt` and `execution.json` and skip validation.
- [x] Validation failures still write a validation report with the failing step recorded.
- [x] `run --parallel` reuses the same runner and validation helper path.
- [x] `--dry-run` still avoids runner and validator execution.
- [ ] `go test ./...` passes.

Note: checklist synced against the current implementation. The remaining validation item was not re-run in this shell.
