# SPEC-051 Plan

1. Refresh the local project context enough to confirm the planning branch contract and validation commands.
2. Reproduce the issue with a long `namba fix --command plan` description and inspect `internal/namba/planning_start.go` plus adjacent planning command tests.
3. Add a focused helper or limit inside planning branch slug generation so `spec/SPEC-XXX-` remains intact and only the trailing slug is capped.
4. Add targeted regression tests proving long `namba plan` and `namba fix --command plan` descriptions produce bounded branch names, including a Korean description case.
5. Verify short branch descriptions still produce the same branch name as before.
6. Run the configured validation commands and targeted Go tests.
7. Sync artifacts with `namba sync` after implementation.
