# SPEC-023

## Problem

`namba` CLI는 현재 command마다 help/argument parsing 계약이 제각각이라, 사용자가 도움말이나 option probing을 기대한 입력이 read-only로 끝나지 않고 실제 mutation 경로로 들어갈 수 있다.

현재 저장소에서 확인되는 구체적인 문제:

- `internal/namba/namba.go`의 `runPlan`, `runHarness`, `runFix`는 `--help`/`-h`를 개별적으로 처리하고, `internal/namba/spec_command_test.go`도 이 세 command에 대해서만 read-only help를 고정하고 있다.
- 반면 `internal/namba/namba.go`의 `runProject`와 `runSync`는 인자를 전혀 파싱하지 않으므로 `namba project --help`나 `namba sync --help`도 실제 문서 생성/갱신 경로로 들어간다.
- `internal/namba/pr_land_command.go`, `internal/namba/release_command.go`, `internal/namba/self_update_command.go`, `internal/namba/namba.go`의 `runExecute`, `runWorktree`, `parseInitArgs`는 `--help`를 usage가 아니라 `unknown flag` 또는 다른 오류로 처리한다.
- `runPR`, `runLand`, `runProject`, `runSync`, `runExecute` 등 여러 command는 help 여부를 판단하기 전에 project root, git repo, GitHub auth, validation, file write 같은 작업 전제조건을 먼저 확인한다.
- `parseDescriptionCommandArgs`는 자유 텍스트 description 안에 `--help`가 들어와도 무조건 help probing으로 해석하므로, "`namba fix --help`가 실제로 SPEC을 생성했다" 같은 문제를 그대로 description으로 넣고 planning하려는 사용 사례를 안전하게 지원하지 못한다.

그 결과 사용자는 아래처럼 의도와 실제 결과가 어긋나는 경험을 하게 된다.

- `namba plan --help`, `namba harness --help`, `namba fix --help`는 일부 고쳐져 있지만, command 전반에서 동일한 기대를 가질 수 없다.
- `namba project --help`나 `namba sync --help`는 도움말을 기대한 입력인데도 실제 저장소 상태를 바꿀 수 있다.
- `namba land --help`나 `namba pr --help`는 usage가 아니라 repo/auth 관련 오류로 끝날 수 있다.
- flag-like text를 포함한 description은 option probing과 실제 요청 텍스트를 안전하게 구분하기 어렵다.

즉 현재 문제는 "특정 command가 우연히 SPEC를 만든다"에 그치지 않고, `namba` 전체 CLI가 "help는 read-only", "malformed args는 non-mutating failure"라는 기본 계약을 아직 전역 불변식으로 가지지 못한 상태라는 점이다.

## Goal

`namba`의 모든 top-level command에서 help/argument 계약을 일관되게 정리해, `--help`/`-h`/`help <command>`는 항상 read-only help 경로로 끝나고, unknown flag나 malformed invocation은 어떤 mutation도 일으키지 않도록 CLI를 harden한다.

## Context

- Project: namba-ai
- Project type: existing
- Language: go
- Mode: tdd
- Work type: plan
- Affected commands: `init`, `doctor`, `status`, `project`, `regen`, `update`, `plan`, `harness`, `fix`, `run`, `sync`, `pr`, `land`, `release`, `worktree`
- Primary implementation areas: `internal/namba/namba.go`, `internal/namba/pr_land_command.go`, `internal/namba/release_command.go`, `internal/namba/self_update_command.go`, shared CLI parsing helpers, usage/help renderers, and regression tests under `internal/namba/*_test.go`
- Existing partial contract to preserve: `plan`, `harness`, `fix` already have read-only help tests; this SPEC should lift that behavior into a top-level CLI invariant instead of leaving it command-specific.

## Desired Outcome

- `namba <command> --help`, `namba <command> -h`, and `namba help <command>` are all supported as read-only help flows.
- `namba help <command>` is semantically aligned with `<command> --help`; implementation can share identical output or an equivalent command-specific usage surface as long as the behavior is consistent and testable.
- Help parsing happens before project-root detection, git-repo checks, GitHub auth, validation, file writes, SPEC generation, sync, update downloads, PR merge, or any other mutation path.
- Commands that do not normally accept positional args or extra flags do not silently ignore them; they either show help or return a non-mutating usage error.
- Commands that accept free-form descriptions support a delimiter form such as `--`, so help-like tokens can be passed as literal description text when needed.
- Unknown flag, missing value, and malformed subcommand handling is consistent across the CLI and never mutates repo state.
- Help output and malformed-invocation error output are clearly distinguishable, so users and agents can tell whether they requested documentation or triggered a usage failure.
- The CLI surface becomes predictable enough that Codex and human users can safely probe commands with `--help` without accidental side effects.

## Target User

- Maintainers using `namba` interactively while discovering or debugging command behavior.
- Codex sessions that routinely probe CLI surfaces with `--help` before deciding how to act.
- Users creating new planning requests whose description may legitimately contain strings like `--help`, `--command`, or other flag-like text.

## Scope

- Define a shared top-level help/argument contract for all first-class `namba` commands.
- Add per-command help rendering where missing and normalize existing help entry points.
- Refactor command dispatch so help detection happens before environmental preconditions or mutations.
- Harden no-arg/read-only commands so stray args are rejected rather than ignored.
- Add delimiter-aware parsing for free-form description commands such as `plan`, `harness`, and `fix`.
- Add regression tests that prove:
  - help probing is read-only
  - malformed invocations are non-mutating
  - flag-like description text is preserved when explicitly requested
  - commands no longer ignore unexpected args and perform work anyway
- Update checked-in docs or command guidance where the CLI contract is described, but only as a consequence of renderer/source changes.

## Non-Goals

- Do not redesign the semantic behavior of `namba plan`, `namba fix`, `namba project`, `namba pr`, or `namba land` beyond help/argument contract hardening.
- Do not change which commands create SPECs versus execute work; only make those transitions safer and more explicit.
- Do not introduce shell-style argument parsing complexity beyond what is needed for stable help flows and delimiter-based literal descriptions.
- Do not paper over the issue by editing docs alone; the CLI behavior and regression tests must become the source of truth.

## Design Constraints

- Treat "help is read-only" and "argument errors are non-mutating" as CLI-wide invariants, not ad hoc command exceptions.
- Parse help before `requireProjectRoot`, git checks, GitHub auth, validation, file writes, or network access.
- Preserve existing valid command behavior for normal non-help invocations.
- Preserve support for literal user text that looks like a flag when the user intentionally passes it after a delimiter.
- Keep usage output short, command-specific, and stable enough for regression testing.
- Keep help output and malformed-invocation output observably different so the CLI interaction remains legible to both humans and Codex.
- Prefer shared parsing/rendering helpers over copying bespoke `--help` checks into every command.
