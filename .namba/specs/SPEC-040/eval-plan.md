# SPEC-040 Evaluation Plan

## Evaluation Goals

- Prove that Namba adapts useful upstream patterns without importing upstream skills or third-party automation flows.
- Prove that existing Namba skills become clearer and safer without bloating command-entry skill bodies.
- Prove that PR/review handoff guidance captures thread state, validation evidence, and CI/check evidence.
- Prove that helper-script candidates are evaluated deterministically before implementation.
- Prove that harness/MCP and frontend validation quality criteria are concrete enough for review.

## Regression Tests

- Template rendering tests:
  - affected render functions in `internal/namba/templates.go` produce the required trigger and behavior guidance
  - generated skill content remains compact and excludes install-flow instructions
  - `namba-pr` includes check inspection and failure-log evidence guidance
  - `namba-review-resolve` includes thread outcome, original-thread reply, validation, and CI/check evidence guidance
  - `namba-harness` includes helper-candidate and harness/MCP evaluation criteria

- Update/regen tests:
  - `namba regen` preserves user-authored skills and updates managed Namba skills from templates
  - no `.codex/skills` or `$CODEX_HOME/skills` install path is reintroduced
  - no new standalone Namba skill appears in the managed skill registry

- Helper-candidate tests, if scripts are implemented:
  - each helper supports `--help`
  - PR-thread and CI helpers can run against recorded JSON fixtures without network
  - frontend helper can run against a local dummy server and capture screenshot/console evidence
  - failing CI logs are bounded and external checks are reported as URLs only

- Review-loop behavior tests, if behavior is executable:
  - unresolved review threads are classified with exactly one allowed outcome
  - resolved state is used before deciding whether to reply or resolve
  - review request marker is present exactly once

- Harness/MCP quality tests:
  - sample harness evaluation scenarios are independent, read-only, realistic, verifiable, and stable
  - tool-quality guidance includes workflow-first design, concise output, actionable errors, and bounded context behavior

## Manual Review Checks

- Verify the generated `SKILL.md` files remain readable and do not become long procedural manuals.
- Verify all added guidance maps back to an existing Namba skill or workflow; nothing creates a new skill.
- Verify excluded upstream flows appear only as rejection criteria, not as instructions to execute.
- Verify Korean PR defaults and existing Namba collaboration policy remain unchanged.

## Validation Commands

Run after implementation:

```bash
namba regen
namba sync
gofmt -l "cmd" "internal" "namba_test.go"
GOCACHE=/tmp/namba-go-cache go test ./...
GOCACHE=/tmp/namba-go-cache go vet ./...
git diff --check
```

## Exit Criteria

- All acceptance criteria in `acceptance.md` are met.
- Review readiness clearly reflects any remaining product, engineering, or design concerns.
- Generated and source diffs are explainable in one PR.
- No excluded third-party install or app-automation flow is introduced.
