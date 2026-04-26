# SPEC-038 Plan

1. Refresh planning context with `.namba/project/product.md`, `.namba/project/tech.md`, mismatch and quality reports, and the relevant source files.
2. Add the `namba-coach` managed skill template.
   - Implement `renderCoachCommandSkill()` in `internal/namba/templates.go`.
   - Keep the skill body read-only and advisory.
   - Include the `$namba-help` boundary, canonical answer order, essential-clarification definition, routing table, wrong-command correction behavior, and todo-list ambiguity example.
3. Register the managed generated skill.
   - Add `namba-coach` to `managedCodexSkillNames()` in `internal/namba/codex.go`.
   - Add `.agents/skills/namba-coach/SKILL.md` to `codexSkillTemplates()`.
   - Update doctor or scaffold checks if they assert the expected command-entry skill set.
4. Expose `$namba-coach` consistently in command surfaces.
   - Update `renderAgents()` to include `$namba-coach` in Codex-native mode, workflow guidance, and command-entry skill lists.
   - Update `renderNambaSkill()` and `renderNambaSkillCommandMappingSection()` so the router knows when to choose coach.
   - Update `renderNambaSkillExecutionRulesSection()` so direct skill invocation preference includes `$namba-coach`.
   - Update `renderCodexUsageInitEnablesSection()`, `renderCodexUsageHowCodexUsesNambaSection()`, and `renderCodexUsageWorkflowCommandSemanticsSection()`.
5. Update generated documentation templates.
   - Update `internal/namba/readme.go` root README command skills and skill mapping sections.
   - Update managed project README and workflow guide copy in all generated languages where those sections are emitted.
   - Preserve the existing distinction between `$namba-help`, `$namba-create`, `namba plan`, and `namba harness`.
6. Add and update tests.
   - In `internal/namba/templates_test.go`, assert `renderCoachCommandSkill()` contains the read-only contract, clarification limits, routing rules, wrong-command correction, and todo-list example.
   - Add anchors proving `renderNambaSkill()` and `renderAgents()` expose `$namba-coach`.
   - In `internal/namba/readme_contract_test.go` and `internal/namba/readme_sync_test.go`, assert README, workflow guide, and Codex usage surfaces expose `$namba-coach` with the same role.
   - In `namba_test.go` or existing init/regen scaffold tests, assert `.agents/skills/namba-coach/SKILL.md` is generated and treated as a managed skill.
   - Add a regen cleanup regression proving `.agents/skills/namba-coach/SKILL.md` survives managed output cleanup once registered.
   - Do not expand public CLI help tests for `namba coach`, because this slice does not add a public CLI command.
7. Regenerate and sync generated outputs.
   - Run `namba regen` after template changes; this is the path that creates and preserves `.agents/skills/namba-coach/SKILL.md`.
   - Run `namba sync` after implementation validation so downstream docs, readiness, PR-ready artifacts, and project summaries reflect the regenerated managed skill.
8. Validate.
   - Run `gofmt -l "cmd" "internal" "namba_test.go"`.
   - Run `go vet ./...`.
   - Run `go test ./...`.
   - Confirm `namba regen` and `namba sync` complete without removing `.agents/skills/namba-coach/SKILL.md`.
