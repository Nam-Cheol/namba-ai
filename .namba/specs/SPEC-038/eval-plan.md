# SPEC-038 Eval Plan

## Template Tests

Add or update tests in `internal/namba/templates_test.go` to assert:

- `renderCoachCommandSkill()` renders front matter for `namba-coach`.
- The skill states that it is read-only and advisory.
- The skill states the canonical answer order: brief restatement, up to three essential questions when required, one primary command, optional single alternative, short reason.
- The skill defines essential clarification as information required to choose the correct workflow or make the handoff command usable.
- The skill asks only 1-3 essential clarification questions.
- The skill recommends one primary command and at most one alternative.
- The skill includes the `$namba-help` boundary.
- The skill includes wrong-command correction behavior.
- The skill includes routing for `namba plan`, `namba harness`, `$namba-create`, `namba fix`, `namba fix --command plan`, `namba run`, `$namba-help`, `namba sync`, `namba pr`, and `namba land`.
- The skill includes the todo-list ambiguity example.
- The skill includes a wrong-command example for `$namba-plan` vs `namba harness`.
- The skill includes a direct artifact creation example that routes to `$namba-create` instead of `namba harness`.
- The skill keeps post-implementation handoff stage-specific: `namba sync`, then `namba pr "<Korean title>"`, then `namba land`.
- `renderNambaSkill()` and `renderAgents()` expose `$namba-coach`.

## Registry And Scaffold Tests

Update existing init or regen scaffold coverage to assert:

- `managedCodexSkillNames()` includes `namba-coach`.
- `codexSkillTemplates()` includes `namba-coach/SKILL.md`.
- `namba init` or `namba regen` creates `.agents/skills/namba-coach/SKILL.md`.
- Regeneration does not delete `.agents/skills/namba-coach/SKILL.md`.
- Managed output cleanup does not remove `.agents/skills/namba-coach/SKILL.md` once `namba-coach` is registered.

## Documentation Tests

Update `internal/namba/readme_contract_test.go` and `internal/namba/readme_sync_test.go` to assert:

- README command-skill sections mention `$namba-coach`.
- Workflow guide sections mention `$namba-coach`.
- `.namba/codex/README.md` generated usage mentions `$namba-coach`.
- The role description is consistent: read-only coaching from current intent to the next Namba workflow.
- Docs distinguish `$namba-help` as "how Namba works and where docs live" from `$namba-coach` as "what should I run next for this current goal."

## Negative Tests

Do not add public CLI help tests for `namba coach`. The absence of a public command remains part of acceptance.

## Validation Commands

Run:

```text
gofmt -l "cmd" "internal" "namba_test.go"
go vet ./...
go test ./...
namba regen
namba sync
```
