# Tech

## Runtime And Validation

- Language: go
- Framework: none
- Important roles: operator, maintainer
- Important runtimes/services: git, codex
- Validation commands: test=`go test ./...`, lint=`gofmt -l "cmd" "internal" "namba_test.go"`, typecheck=`go vet ./...`

## Analysis Contract

- Evidence: every generated finding cites one or more repository paths.
- Confidence: `high` means direct code/config support, `medium` means multiple weaker signals or doc-backed interpretation, and `low` means heuristic inference.
- Conflict handling: `mismatch-report.md` preserves stronger-vs-weaker source disagreements instead of flattening them into neutral prose.

## Systems

- `workspace` (go-service): `.namba/project/systems/workspace.md`

## Planning Signals

- Configured include paths: .
- Configured exclude paths: .git, .cache, .gocache, .idea, .vscode, .venv, venv, node_modules, dist, build, coverage, tmp, .tmp, logs, external, vendor, .namba/logs, .namba/project, .namba/worktrees
- `structure.md` is appendix material; use `product.md` and the per-system docs first.
- `mismatch-report.md` preserves code-vs-doc conflicts instead of flattening them into prose.
