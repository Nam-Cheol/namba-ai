# SPEC-037 Hook Runtime

## 1. Overview

Introduce a framework-agnostic Hook Runtime for NambaAI's execution pipeline.
The runtime allows repository-local commands to run at well-defined lifecycle
events during `namba run SPEC-XXX`, records every executed hook result as
runtime evidence, and keeps the contract independent from any one execution
engine.

Namba's existing execution path already produces structured artifacts under
`.namba/logs/runs/` for request, preflight, execution, validation, and typed
execution evidence. This SPEC extends that evidence layer with hook results
instead of creating a second logging model.

The v1 runtime is intentionally conservative:

- Hooks are registered by config in `.namba/hooks.toml`.
- Hook failures are non-blocking by default.
- Execution stops only when the failing hook has `continue_on_failure = false`.
- Hook stdout and stderr are persisted to files and referenced from the
  evidence manifest.
- Namba hook events are Namba-owned lifecycle events, not copied Codex hook
  event names.

## 2. Goals / Non-Goals

### Goals

- Define a Namba-owned Hook Runtime aligned with:
  `request -> preflight -> execution -> validation -> evidence manifest`.
- Add hook event points for preflight, execution, patch/bash/MCP tool
  observations, validation, and terminal failures.
- Define `.namba/hooks.toml` registration with the required fields:
  `event`, `command`, `cwd`, `timeout`, `enabled`, and
  `continue_on_failure`.
- Record every executed hook result in
  `.namba/logs/runs/<log-id>-evidence.json`.
- Persist hook stdout/stderr as separate artifacts referenced by the evidence
  manifest.
- Keep the runtime extensible for Codex, Claude Code, MCP-based engines, and
  future Namba runners through a normalized Namba event model.

### Non-Goals

- Do not implement Codex hook configuration compatibility in v1.
- Do not mirror Codex event names such as `PreToolUse`, `PostToolUse`,
  `PermissionRequest`, `UserPromptSubmit`, or `Stop`.
- Do not include auto-review behavior, permission systems, or approval flows.
- Do not design multiple environment profiles or remote-environment hook
  routing.
- Do not add implementation code in this SPEC.
- Do not require all runners to emit patch/bash/MCP tool observations on day
  one. Runners that cannot observe a tool boundary simply do not emit that
  tool-boundary event.

## 3. Reference Basis

This design uses stable, verifiable Codex v0.124 hook capabilities as a
reference point without coupling Namba to Codex internals.

Verified sources as of 2026-04-25:

- OpenAI Codex changelog for Codex CLI 0.124.0, dated 2026-04-23:
  hooks are stable, can be configured inline in `config.toml` and managed
  `requirements.toml`, and can observe MCP tools, `apply_patch`, and
  long-running Bash sessions.
  Source: https://developers.openai.com/codex/changelog
- OpenAI Codex Hooks documentation:
  hooks run deterministic scripts during the Codex lifecycle; command hooks
  have timeouts; commands run with the session working directory; matcher
  filters include tool names for `PreToolUse`, `PermissionRequest`, and
  `PostToolUse`; supported tool-name matching includes `Bash`, `apply_patch`,
  and MCP tool names; `PostToolUse` runs after supported tools including Bash,
  `apply_patch`, and MCP tool calls.
  Source: https://developers.openai.com/codex/hooks
- OpenAI Codex Hooks documentation also states that post-tool hooks cannot undo
  side effects that already happened and that shell interception is not
  universal. Namba therefore treats tool-boundary hooks as evidence-producing
  observations, not as transactional rollback points.
  Source: https://developers.openai.com/codex/hooks

Namba-specific local basis:

- `internal/namba/execution.go` writes request, preflight, execution,
  validation, and validation-attempt artifacts under `.namba/logs/runs/`.
- `internal/namba/execution_evidence.go` emits the typed
  `.namba/logs/runs/<log-id>-evidence.json` manifest and keeps it advisory by
  default.
- `internal/namba/runtime_contract.go` defines `executionRequest`,
  `preflightReport`, execution modes, runner selection, and quality validation
  command configuration.

## 4. Hook Runtime Architecture

The Hook Runtime is a Namba-owned component invoked by the run lifecycle. It
has four responsibilities:

1. Load hook registrations from `.namba/hooks.toml`.
2. Resolve enabled hooks for a triggered Namba event.
3. Execute matching hook commands with a bounded timeout and normalized input
   context.
4. Persist hook stdout, stderr, and result metadata into run evidence.

### Runtime Boundaries

- The Hook Runtime belongs to Namba's execution orchestration layer, not to a
  specific runner implementation.
- Runners may emit normalized tool observations to Namba through a typed
  observation sink:
  `patch_applied`, `bash_completed`, or `mcp_tool_completed`.
- Namba maps those observations to `after_patch`, `after_bash`, and
  `after_mcp_tool`.
- If a runner cannot observe a tool boundary, Namba must not fabricate that
  event from logs or free-form model output.
- All pipeline-stage events are emitted by Namba itself because preflight,
  execution, validation, and evidence finalization are Namba-owned phases.

### Per-Run Lifecycle Owner

Implementation must introduce one per-run Hook Lifecycle owner instead of
scattering hook calls across every early return branch in `executeRun`.

The lifecycle owner is responsible for:

- buffering hook results until the evidence manifest is written
- writing hook stdout/stderr artifacts as each hook finishes
- preserving the primary Namba failure status when hooks also fail
- deciding whether a hook failure is advisory or blocking
- triggering `on_failure` at most once
- finalizing the execution evidence manifest after all relevant hook results
  are recorded

The lifecycle owner must be the only code path that appends hook results to the
manifest. Existing request, preflight, execution, validation, progress, and
runtime evidence references remain owned by the existing execution-evidence
builder.

### Runner Observation Contract

Tool-boundary events require a typed observation contract before
implementation. A runner observation must include:

- `observation_type`: `patch_applied`, `bash_completed`, or
  `mcp_tool_completed`
- `tool_name`: canonical runner tool name, such as `apply_patch`, `Bash`, or an
  MCP tool id
- `tool_use_id`: runner-provided invocation id when available
- `started_at` and `ended_at` when the runner exposes them
- `status`: runner-reported tool outcome when available
- `exit_code`: required for bash/shell observations when available
- `command`: required for bash/shell observations when available
- `cwd`: command working directory when available
- `input_summary`: bounded summary of tool input when available
- `output_artifacts`: paths to runner-provided output artifacts when available

If the active runner does not support observations, Namba records no
tool-boundary hook results and must not imply that `after_patch`,
`after_bash`, or `after_mcp_tool` hooks were evaluated. The run evidence may
include a non-hook capability note such as `tool_observations: unsupported`,
but disabled or unsupported tool observations must not produce fake hook
results.

### Parallel Mode Scope

For v1, Namba-owned lifecycle hooks run at worker scope for actual execution
workers:

- worker `executeRun` emits `before_preflight`, `after_preflight`,
  `before_execution`, `after_execution`, `before_validation`,
  `after_validation`, and worker-level `on_failure`
- worker hook results are recorded in the worker log id evidence manifest
- aggregate parallel lifecycle code does not run hook commands in v1
- aggregate parallel evidence may reference worker manifests through existing
  parallel progress/evidence links, but it does not duplicate worker hook
  results

This keeps v1 aligned with the code that already owns request, preflight,
execution, validation, and repair attempts. Aggregate parallel hooks can be a
future extension only after the aggregate lifecycle has its own hookable phase
contract.

### Event Dispatch Contract

Each hook invocation receives a normalized JSON context through stdin. The
exact serialization can be implementation-defined, but the context must include
at least:

- `schema_version`: hook context schema version, initially `namba-hook/v1`
- `event`: the event name
- `log_id`: current run log id
- `run_id`: current run id when available, otherwise the log id
- `spec_id`: target SPEC id
- `execution_mode`: default, solo, team, or parallel
- `work_dir`: active execution working directory
- `project_root`: repository root
- `artifacts`: known request/preflight/execution/validation/evidence paths
- `stage_status`: current phase status when known
- `triggered_at`: timestamp for the event trigger

Tool-boundary events add tool-specific context as described in the event model.

### Operator Outcome Glossary

- `status`: hook process outcome only. Valid values are `succeeded`, `failed`,
  `timeout`, and `error`.
- `stage_status`: Namba pipeline phase status, such as preflight passed/failed,
  execution succeeded/failed, or validation passed/failed.
- `failure_status`: terminal Namba run status when a failure path is active,
  such as `preflight_failed`, `execution_failed`, `validation_failed`, or
  `hook_failed`.
- `blocking`: whether this hook result stopped the Namba run.
- `failure_action`: what Namba did after a hook failure. Valid values are
  `continued`, `stopped`, or `not_applicable`.
- `error_summary`: short troubleshooting summary for `error` status, including
  malformed config, spawn failure, or cwd resolution failure.

Operator reading order for hook evidence is:
`event -> hook_name -> status -> exit_code -> blocking/failure_action -> error_summary -> stdout_path/stderr_path`.

### Ordering

For v1, enabled hooks for the same event run serially in deterministic
`hook_name` order. This keeps evidence ordering stable and avoids concurrency
semantics before Namba has a use case that needs them.

## 5. Hook Event Model

All events are advisory by default. A hook failure only stops execution when
that hook's registration has `continue_on_failure = false`.

| Event | Trigger timing | Input context | Default behavior on hook failure |
| --- | --- | --- | --- |
| `before_preflight` | After the run request is resolved and the log id is known, before preflight checks execute. | Base context plus request path when already written or planned request metadata when not yet written. | Record result and continue. If `continue_on_failure=false`, stop before preflight, write evidence, then trigger `on_failure`. |
| `after_preflight` | Immediately after preflight report is written, regardless of pass/fail. | Base context plus preflight artifact path, preflight pass/fail state, and preflight error summary when present. | Record result and continue if preflight passed. If the hook is blocking, stop before execution and trigger `on_failure`. If preflight failed, preserve the preflight failure as the primary run failure. |
| `before_execution` | After preflight passes and before the selected runner starts execution turns. | Base context plus runner, model/profile metadata when configured, delegation mode, and execution artifact target path. | Record result and continue. If blocking, stop before runner invocation, write evidence, then trigger `on_failure`. |
| `after_execution` | After the runner finishes all execution turns or fails before validation starts. | Base context plus execution artifact path, turn summary, runner success/failure state, and execution error summary when present. | Record result and continue to validation only when execution succeeded. If blocking, stop before validation and trigger `on_failure`. Existing runner failure remains the primary failure. |
| `after_patch` | After a runner reports a completed patch/file-edit operation. For Codex v0.124 this maps to observable `apply_patch` post-tool behavior when available. | Base context plus `tool_name`, `tool_use_id` when available, patch target metadata when available, tool input summary, tool status, and related artifact paths. | Record result and continue. A blocking hook can stop subsequent execution, but it cannot undo the patch. |
| `after_bash` | After a runner reports a completed shell/Bash command, including non-zero exits when the runner exposes them. | Base context plus `tool_name`, `tool_use_id` when available, shell command, command cwd, exit code, duration when available, and output artifact references when available. | Record result and continue. A blocking hook can stop subsequent execution, but it cannot undo command side effects. |
| `after_mcp_tool` | After a runner reports a completed MCP tool invocation. | Base context plus MCP tool name, server/tool identifier when available, tool input summary, tool result status, and output artifact references when available. | Record result and continue. A blocking hook can stop subsequent execution, but it cannot undo completed external tool side effects. |
| `before_validation` | After execution succeeds and before Namba starts the validation command pipeline. | Base context plus execution artifact path, validation attempt number, validation root, and configured validation commands. | Record result and continue. If blocking, stop before validation, write evidence, then trigger `on_failure`. |
| `after_validation` | After each validation report is written, including failed attempts and final success/failure. | Base context plus validation artifact path, attempt number, pass/fail state, failing command summaries, and retry state. | Record result and continue according to validation outcome. If blocking, stop before repair/next step and trigger `on_failure`. Existing validation failure remains the primary failure when validation itself failed. |
| `on_failure` | After Namba determines a terminal failure or a blocking hook stops the run. It runs once per terminal failure path and never recursively triggers itself. | Base context plus `failure_phase`, `failure_status`, primary error summary, blocking hook result when applicable, and available artifact paths. | Record result. It cannot replace or hide the primary failure. If its own `continue_on_failure=false` fails, record that failure but do not trigger another `on_failure`. |

### Tool-Observation Availability

`after_patch`, `after_bash`, and `after_mcp_tool` are emitted only when the
active runner supplies normalized tool observations. For Codex-backed runs,
Codex v0.124 provides a verified basis for observing `apply_patch`, Bash, and
MCP tool calls, but Namba must keep the integration behind the normalized
runner boundary.

## 6. Hook Registration Config

The registration file is:

`.namba/hooks.toml`

The file contains named hook tables. The table key is the stable `hook_name`
stored in evidence. This avoids adding a separate required `name` field while
still making evidence records stable.

Example shape:

```toml
[hooks.generated_docs_check]
event = "after_patch"
command = "go test ./..."
cwd = "."
timeout = 120
enabled = true
continue_on_failure = true

[hooks.validation_guard]
event = "before_validation"
command = "./scripts/check-validation-inputs.sh"
cwd = "."
timeout = 30
enabled = true
continue_on_failure = false
```

### Required Fields

| Field | Type | Required | Meaning |
| --- | --- | --- | --- |
| `event` | string | yes | One of the v1 event names listed in the Hook Event Model. |
| `command` | string | yes | Shell command to execute for the hook. Namba passes normalized event context through stdin. |
| `cwd` | string | yes | Working directory for the hook. Relative paths resolve from project root. |
| `timeout` | integer | yes | Maximum runtime in seconds. |
| `enabled` | boolean | yes | Disabled hooks are ignored and do not produce hook results. |
| `continue_on_failure` | boolean | yes | When true, failure is recorded but does not stop execution. When false, failure stops the run after evidence is recorded. |

### Config Rules

- Unknown events are config errors.
- Empty `command` is a config error.
- `timeout` must be greater than zero.
- `cwd` must resolve inside the repository root for v1.
- Hook names are derived from the TOML table key under `[hooks.<hook_name>]`.
- Hook names must be non-empty and stable after normalization.
- Multiple hooks may register for the same event.
- Disabled hooks are not executed and are not represented as hook results.
- Missing `.namba/hooks.toml` is a no-op and must not fail a run.
- Malformed `.namba/hooks.toml` is a run-start config error. Once a log id is
  available, Namba records a hook config error in evidence, triggers
  `on_failure`, and stops before execution begins.

## 7. Hook Execution Lifecycle

Every event follows the same lifecycle:

1. Event is triggered by Namba or by a normalized runner observation.
2. Hook Runtime loads `.namba/hooks.toml`.
3. Runtime selects registrations where `event` matches the triggered event.
4. Runtime filters out registrations where `enabled = false`.
5. Runtime sorts selected hooks by `hook_name`.
6. Runtime constructs the normalized hook input context.
7. Runtime executes each hook command in its configured `cwd`.
8. Runtime enforces `timeout`.
9. Runtime captures stdout, stderr, exit code, start time, end time, and
   duration.
10. Runtime writes stdout and stderr artifacts under the current run log
    namespace.
11. Runtime constructs a hook result.
12. Runtime appends the hook result to the current evidence manifest state.
13. Runtime applies `continue_on_failure` policy if the hook did not succeed.
14. If a blocking hook fails, Namba stops the relevant phase, finalizes
    evidence, and triggers `on_failure` once.

### Output Artifact Paths

Hook output artifacts should live under the run namespace, for example:

- `.namba/logs/runs/<log-id>-hooks/<event>/<hook-name>-stdout.txt`
- `.namba/logs/runs/<log-id>-hooks/<event>/<hook-name>-stderr.txt`

The implementation may add collision-safe suffixes when the same hook is
executed more than once in a run, such as repeated `after_validation` attempts.
The evidence record is the source of truth for the exact paths.

## 8. Evidence Schema Extension

Extend `.namba/logs/runs/<log-id>-evidence.json` with a top-level `hooks`
array. The existing execution evidence remains the base manifest; hook evidence
is an extension on that manifest.

Each executed hook result must include the following fields:

| Field | Type | Required | Meaning |
| --- | --- | --- | --- |
| `event` | string | yes | Event that triggered the hook. |
| `hook_name` | string | yes | Stable hook name from `[hooks.<hook_name>]`. |
| `command` | string | yes | Command string from config. |
| `cwd` | string | yes | Resolved working directory used for execution. |
| `started_at` | string | yes | RFC3339 timestamp. |
| `ended_at` | string | yes | RFC3339 timestamp. |
| `duration_ms` | integer | yes | Non-negative wall-clock duration in milliseconds. |
| `exit_code` | integer | yes | Process exit code, or `-1` when no process exit code exists because of timeout or spawn failure. |
| `status` | string | yes | One of `succeeded`, `failed`, `timeout`, or `error`. |
| `stdout_path` | string | yes | Relative path to captured stdout artifact. Empty output still gets a file. |
| `stderr_path` | string | yes | Relative path to captured stderr artifact. Empty output still gets a file. |

Optional future-compatible fields may include:

- `attempt`: validation attempt or repeated tool-observation sequence number
- `tool_name`: normalized tool name for tool-boundary events
- `tool_use_id`: runner-provided tool invocation id when available

New v1 hook-result records must also include these operator fields:

- `blocking`: boolean; true when the hook stopped execution
- `failure_action`: `continued`, `stopped`, or `not_applicable`
- `error_summary`: short cause for `error`, `timeout`, or config failure
- `scope`: `worker` for v1 run hooks; future values may include `aggregate`

These fields are not a substitute for the required fields above. They make the
manifest self-explanatory for operators. Older evidence manifests remain
compatible when they do not contain these fields, but newly emitted v1 hook
results must include them.

### Manifest Example

```json
{
  "hooks": [
    {
      "event": "before_validation",
      "hook_name": "validation_guard",
      "command": "./scripts/check-validation-inputs.sh",
      "cwd": "/repo",
      "started_at": "2026-04-25T12:00:00+09:00",
      "ended_at": "2026-04-25T12:00:01+09:00",
      "duration_ms": 812,
      "exit_code": 0,
      "status": "succeeded",
      "blocking": false,
      "failure_action": "not_applicable",
      "stdout_path": ".namba/logs/runs/spec-037-hooks/before_validation/validation_guard-stdout.txt",
      "stderr_path": ".namba/logs/runs/spec-037-hooks/before_validation/validation_guard-stderr.txt"
    },
    {
      "event": "after_bash",
      "hook_name": "log_shell",
      "command": "./scripts/log-shell.sh",
      "cwd": "/repo",
      "started_at": "2026-04-25T12:00:02+09:00",
      "ended_at": "2026-04-25T12:00:03+09:00",
      "duration_ms": 301,
      "exit_code": 1,
      "status": "failed",
      "blocking": false,
      "failure_action": "continued",
      "stdout_path": ".namba/logs/runs/spec-037-hooks/after_bash/log_shell-stdout.txt",
      "stderr_path": ".namba/logs/runs/spec-037-hooks/after_bash/log_shell-stderr.txt"
    },
    {
      "event": "before_execution",
      "hook_name": "required_policy_check",
      "command": "./scripts/policy-check.sh",
      "cwd": "/repo",
      "started_at": "2026-04-25T12:00:04+09:00",
      "ended_at": "2026-04-25T12:00:05+09:00",
      "duration_ms": 402,
      "exit_code": 2,
      "status": "failed",
      "blocking": true,
      "failure_action": "stopped",
      "error_summary": "blocking hook exited non-zero",
      "stdout_path": ".namba/logs/runs/spec-037-hooks/before_execution/required_policy_check-stdout.txt",
      "stderr_path": ".namba/logs/runs/spec-037-hooks/before_execution/required_policy_check-stderr.txt"
    }
  ]
}
```

### Evidence Invariants

- Every enabled hook that starts execution must have one hook result.
- Hook results must be recorded on success, non-zero exit, timeout, spawn
  failure, and blocking failure.
- Hook stdout/stderr paths must point to artifacts under `.namba/logs/runs/`.
- The evidence manifest must still be finalized when a hook blocks before
  preflight, before execution, before validation, or during failure handling.
- Hook evidence is advisory metadata unless the hook itself is configured with
  `continue_on_failure = false`.
- Additive hook evidence remains compatible with historical
  `execution-evidence/v1` manifests. Implementation may keep the schema version
  at `execution-evidence/v1` for additive `hooks` support, or bump to
  `execution-evidence/v2` only if strict schema validation requires it. In both
  cases, readers must tolerate absent `hooks` on older manifests.

## 9. Failure Handling

### Failure Classification

A hook is considered failed when any of the following occurs:

- The command exits with a non-zero exit code.
- The command exceeds `timeout`.
- The command cannot be spawned.
- The configured `cwd` cannot be resolved or entered.
- The hook config is malformed enough that the hook cannot be executed.
- A hook config file is syntactically invalid or contains an unsupported event.

### Default Policy

The v1 default policy is non-blocking:

- Record the hook result.
- Keep the original Namba phase moving.
- Do not reinterpret a pipeline stage as failed solely because an advisory hook
  failed.

### Blocking Policy

When `continue_on_failure = false` and the hook fails:

- Record stdout, stderr, and hook result first.
- Mark the hook result as failed, timed out, or errored.
- Stop the active Namba phase at the next safe boundary.
- Finalize evidence with available request/preflight/execution/validation
  artifact references.
- Trigger `on_failure` once with the blocking hook result in context.
- Preserve completed side effects. `after_patch`, `after_bash`, and
  `after_mcp_tool` cannot roll back work that already happened.

### `on_failure` Policy

- `on_failure` runs after terminal Namba failures and after blocking hook
  failures.
- It receives the primary failure status and available artifacts.
- It must not recursively trigger itself.
- Its failure must be recorded, but it must not hide the original failure.

### Config Error Policy

- If `.namba/hooks.toml` is absent, the Hook Runtime is inactive, produces no
  hook results, and does not fail the run.
- If `.namba/hooks.toml` is present but syntactically malformed, Namba records
  one config-error hook result when a run log id exists, triggers `on_failure`,
  and stops before preflight.
- If a registration uses an unknown event, empty command, non-positive timeout,
  or out-of-root cwd, Namba treats that registration as a config error. The run
  stops before execution begins after evidence is recorded.
- Disabled malformed hook entries are still parsed as part of the config file;
  syntax-level TOML errors fail the config regardless of `enabled`.
- Config-error hook results use `status: "error"`, `exit_code: -1`,
  `blocking: true`, `failure_action: "stopped"`, and an `error_summary` that
  names the invalid field or parse failure.

## 10. Acceptance Criteria

- [ ] `.namba/hooks.toml` is the v1 hook registration location.
- [ ] Hook registrations are named by `[hooks.<hook_name>]`, and that key is
  recorded as `hook_name` in evidence.
- [ ] The following event names are supported exactly:
  `before_preflight`, `after_preflight`, `before_execution`,
  `after_execution`, `after_patch`, `after_bash`, `after_mcp_tool`,
  `before_validation`, `after_validation`, and `on_failure`.
- [ ] Each event has a deterministic trigger point in the Namba run lifecycle
  or in a normalized runner tool-observation boundary.
- [ ] Hook input context includes run identity, SPEC identity, execution mode,
  working directory, artifact paths, event name, and event-specific data.
- [ ] Enabled hooks for one event execute in deterministic `hook_name` order.
- [ ] Disabled hooks are ignored and do not produce hook result records.
- [ ] Hook command execution captures stdout, stderr, exit code, start time,
  end time, and duration.
- [ ] Hook stdout and stderr are persisted under `.namba/logs/runs/` and
  referenced by evidence paths.
- [ ] `.namba/logs/runs/<log-id>-evidence.json` includes a top-level `hooks`
  array containing every executed hook result.
- [ ] Every hook result includes `event`, `hook_name`, `command`, `cwd`,
  `started_at`, `ended_at`, `duration_ms`, `exit_code`, `status`,
  `stdout_path`, and `stderr_path`.
- [ ] Hook failures are non-blocking by default.
- [ ] A hook failure stops execution only when `continue_on_failure=false`.
- [ ] Blocking hook failures are recorded before execution stops.
- [ ] `on_failure` runs once for terminal failures and does not recursively
  trigger itself.
- [ ] `after_patch`, `after_bash`, and `after_mcp_tool` are emitted only from
  normalized runner observations; Namba does not infer them from free-form logs.
- [ ] The design remains framework-agnostic and does not depend on Codex
  config files, Codex event names, or Codex permission/approval behavior.
- [ ] Regression coverage proves success, advisory failure, blocking failure,
  timeout, missing config, malformed config, stdout/stderr artifact persistence,
  evidence manifest extension, and failure-path finalization.
- [ ] Existing request, preflight, execution, validation, and execution
  evidence artifacts remain compatible with historical runs.
- [ ] Validation commands pass.
