# SPEC-037 Plan

1. Confirm the current execution pipeline and artifact boundaries:
   - inspect `internal/namba/execution.go`
   - inspect `internal/namba/execution_evidence.go`
   - inspect `internal/namba/runtime_contract.go`
2. Introduce a hook config contract for `.namba/hooks.toml`:
   - named `[hooks.<hook_name>]` registrations
   - required fields: `event`, `command`, `cwd`, `timeout`, `enabled`, `continue_on_failure`
   - deterministic validation and ordering rules
3. Add the Hook Runtime boundary:
   - load config per triggered event
   - filter enabled hooks
   - execute commands with timeout and cwd resolution
   - capture stdout, stderr, exit code, timing, and status
   - apply `continue_on_failure`
   - keep one per-run lifecycle owner responsible for buffering hook results,
     running `on_failure` once, and finalizing evidence after hook results are recorded
4. Wire Namba-owned lifecycle events into the existing run flow:
   - `before_preflight`
   - `after_preflight`
   - `before_execution`
   - `after_execution`
   - `before_validation`
   - `after_validation`
   - `on_failure`
5. Define the runner-observation seam for tool-boundary events:
   - `after_patch`
   - `after_bash`
   - `after_mcp_tool`
   - emit only from normalized runner observations, never from free-form output inference
   - add a typed observation sink/callback contract before implementing any
     tool-boundary hook event
   - record no hook result when the active runner does not support tool observations
6. Extend execution evidence:
   - add a top-level `hooks` array to `.namba/logs/runs/<log-id>-evidence.json`
   - persist stdout/stderr artifacts under `.namba/logs/runs/<log-id>-hooks/`
   - preserve current request/preflight/execution/validation evidence compatibility
   - include operator fields for `blocking`, `failure_action`, `error_summary`, and `scope`
7. Implement failure handling:
   - advisory hook failures continue by default
   - `continue_on_failure=false` stops at the next safe boundary
   - blocking failures are recorded before stop
   - `on_failure` runs once and never recursively triggers itself
   - missing `.namba/hooks.toml` is a no-op
   - malformed `.namba/hooks.toml` stops before preflight after config-error evidence is recorded
8. Implement parallel v1 scope explicitly:
   - run hooks at worker scope through worker `executeRun`
   - record worker hook results in worker evidence manifests
   - do not run aggregate parallel hooks in v1
   - keep aggregate evidence from duplicating worker hook results
9. Add regression tests for:
   - missing config
   - malformed config
   - event filtering and disabled hooks
   - deterministic hook ordering
   - stdout/stderr artifact persistence
   - successful hook evidence
   - advisory hook failure
   - blocking hook failure before preflight/execution/validation
   - timeout status and `exit_code=-1`
   - on-failure evidence finalization
   - tool-boundary event emission through normalized runner observations
   - unsupported runner observations produce no fake hook results
   - worker-scope parallel evidence ownership
10. Update docs and generated Namba guidance if the user-facing run contract changes.
11. Run validation commands.
12. Sync artifacts with `namba sync`.
