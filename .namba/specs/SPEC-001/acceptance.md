# Acceptance

- [ ] `namba run` uses a runner abstraction instead of calling Codex directly from the orchestration path.
- [ ] Serial execution writes `request.md`, `result.txt`, `execution.json`, and `validation.json` logs when execution succeeds.
- [ ] Runner failures write `result.txt` and `execution.json` and skip validation.
- [ ] Validation failures still write a validation report with the failing step recorded.
- [ ] `run --parallel` reuses the same runner and validation helper path.
- [ ] `--dry-run` still avoids runner and validator execution.
- [ ] `go test ./...` passes.