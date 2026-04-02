# Acceptance

- [ ] Root README is rewritten as a first-session onboarding surface rather than a loose command list.
- [ ] Root README clearly differentiates `namba project`, `namba plan`, `namba harness`, and `namba fix` in user-facing language that explains when each command should be chosen.
- [ ] Root README explains the overall delivery path from install/init through planning, execution, sync, PR handoff, and merge in a way a first-time user can follow without prior Namba context.
- [ ] The command skill section explains the purpose of each major command-entry skill in practical terms, including `$namba`, `$namba-project`, `$namba-plan`, `$namba-harness`, `$namba-fix`, `$namba-run`, `$namba-sync`, `$namba-pr`, `$namba-land`, `$namba-regen`, and `$namba-update`.
- [ ] The documentation makes it obvious when the plan-review skills are used and what problem they solve.
- [ ] Workflow guide clearly separates:
  - lifecycle commands such as `update`, `regen`, `sync`, `pr`, and `land`
  - planning commands such as `project`, `plan`, `harness`, and `fix`
  - execution modes such as default, `--solo`, `--team`, and `--parallel`
- [ ] Workflow guide review-readiness section includes `namba harness` alongside `namba plan` and `namba fix --command plan` where applicable.
- [ ] No checked-in generated README/workflow guide retains stale phrasing that implies only `namba plan` and `namba fix` create SPEC packages when `namba harness` is also supported.
- [ ] `internal/namba/readme.go` remains the source of truth for README/workflow-guide output, and the implementation does not rely on hand-editing generated docs to close gaps.
- [ ] Root README and workflow guide have clearly different responsibilities so the same onboarding explanation is not duplicated in slightly different wording across both surfaces.
- [ ] Generated docs are updated through renderer/source changes plus `namba sync`, not by hand-editing generated output alone.
- [ ] README and workflow guide changes are reflected consistently across the supported generated language variants where the renderer manages them.
- [ ] Regression coverage explicitly protects against renderer/generated-doc drift for the new command-choice guidance, skill-purpose guidance, and `namba harness` exposure.
- [ ] Validation commands pass
- [ ] Tests covering the new renderer/doc behavior are present
