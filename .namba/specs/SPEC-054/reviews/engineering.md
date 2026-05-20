# Engineering Review

- Status: approved
- Last Reviewed: 2026-05-20
- Reviewer: Codex (`namba-planner` lens)
- Command Skill: `$namba-plan-eng-review`
- Recommended Role: `namba-planner`

## Focus

- Lock architecture, sequencing, failure modes, trust boundaries, and validation strategy before execution starts.

## Findings

1. High: The implementation must cover both scaffolded skill contracts and
   runtime body generation. Current evidence points to
   `internal/namba/templates.go`, `internal/namba/pr_land_command.go`, and
   `internal/namba/release.go` as the first code surfaces to inspect.
2. High: `buildPullRequestBody` currently emits references to generated
   summary and checklist files, but not enough concrete completed-work detail
   inside the PR body itself.
3. Medium: `renderReleaseNotes` already preserves commit-derived evidence,
   but the SPEC needs validation evidence and language behavior covered by
   tests so release notes do not remain Korean-only or validation-light by
   accident.
4. Medium: Generic-output prevention should be tested through required
   sections, evidence tokens, and non-empty content rather than brittle prose
   snapshots alone.

## Decisions

- Use targeted fixture or snapshot tests around init scaffolding, PR body
  rendering, scaffolded skill rendering, and release note handoff behavior.
- Keep GitHub mechanics unchanged unless a body string must be passed through
  more faithfully.
- Preserve current explicit review-request mechanics and marker de-duplication
  tests.

## Follow-ups

- Implementation should map the available evidence sources before changing
  output text: SPEC files, `.namba/project/change-summary.md`, PR metadata,
  commit refs, validation output, and release notes artifacts.
- Run repository validation after implementation: `go test ./...`,
  `gofmt -l "cmd" "internal" "namba_test.go"`, and `go vet ./...`.

## Recommendation

- Proceed. The SPEC identifies the right implementation boundaries and test
  strategy for a controlled template and handoff change.
