# SPEC-062 Harness Map

## Current Boundaries

| Surface | Current location | SPEC-062 boundary direction |
| --- | --- | --- |
| CLI command glue | `cmd/namba`, `internal/namba/namba.go` | Keep thin and behavior-compatible |
| Execution evidence | `internal/namba/execution_evidence.go` | Candidate evidence contract package |
| Eval engine | `internal/namba/eval_*` | Candidate eval package with stable schema helpers |
| Report and observability | `internal/namba/report.go` | Candidate observability package |
| Hook runtime | `internal/namba/hook_runtime.go` | Candidate hook runtime package |
| Queue state and runtime | `internal/namba/queue_command.go` | Candidate queue package after schema tests exist |
| GitHub handoff | PR and land command files | Candidate handoff package if dependency direction stays clean |
| Release workflow | release command files | Keep behavior unchanged; refactor only if isolated and covered |
| README and docs rendering | `internal/namba/readme.go` | Candidate docs rendering utility boundary |

## Architecture Check Intent

The architecture check should prevent higher-level packages from importing CLI command glue or unrelated command implementations. Acceptable dependencies should flow from command glue into focused internal packages, not from focused packages back into command parsing.

## Refactor Safety

- Move code only after schema and output contract tests exist.
- Keep public command behavior and output stable.
- Run eval and report JSON validation before and after moves.
- Preserve generated docs behavior through renderer tests and `namba sync`.
