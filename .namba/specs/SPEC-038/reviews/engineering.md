# Engineering Review

- Status: clear
- Last Reviewed: 2026-04-26
- Reviewer: namba-planner
- Command Skill: `$namba-plan-eng-review`
- Recommended Role: `namba-planner`

## Focus

- Re-verify the revised SPEC artifacts for `namba regen` vs `namba sync` semantics, managed registry/template ownership, regen cleanup regression coverage, no public `namba coach` CLI leakage, and the `core_harness_change` route.

## Findings

- The revised source artifacts now describe the regen/sync split consistently. `spec.md`, `contract.md`, `acceptance.md`, `plan.md`, and `eval-plan.md` all state that `namba regen` is the command that creates and preserves `.agents/skills/namba-coach/SKILL.md`, while `namba sync` refreshes downstream README, workflow, project, readiness, and support artifacts after implementation. That matches the current code: `runRegen` writes `AGENTS.md` plus `codexScaffoldFiles(...)` through `isRegenManagedPath`, while `runSync` only replaces outputs covered by `isReadmeManagedPath`, `isProjectAnalysisManagedPath`, `isSpecReviewReadinessManagedPath`, and `isSyncProjectSupportManagedPath`. `namba sync` is correctly framed as downstream refresh, not repo-skill backfill.
- Managed ownership responsibilities are now explicit and implementation-correct. `internal/namba/codex.go` splits the contract between `managedCodexSkillNames()` and `codexSkillTemplates(profile)`, and regen ownership/cleanup depends on both: `isManagedRepoSkillPath(...)` is driven by the managed-name registry, while `codexScaffoldFiles(...)` materializes the template output map. `internal/namba/output_session.go` then stamps manifest owner/checksum state for written outputs. The revised plan and eval plan correctly require adding `namba-coach` to both lists; adding only the template would leave cleanup and manifest ownership behavior inconsistent.
- Regen cleanup regression coverage is now called out at the right level. The repo already has adjacent evidence in `internal/namba/create_workflow_test.go`: regen removes stale managed artifacts and preserves non-regen managed outputs. The revised acceptance and eval plan correctly add one coach-specific regression proving `.agents/skills/namba-coach/SKILL.md` survives managed cleanup once registered. That is the right targeted guard for this slice instead of relying only on broad scaffold tests.
- The no-public-CLI boundary is now tight across the source artifacts. `spec.md`, `contract.md`, `acceptance.md`, `plan.md`, `eval-plan.md`, and `baseline.md` all explicitly forbid a public `namba coach` command or help surface. That matches the current code layout: public command registration lives in `internal/namba/namba.go`, and `internal/namba/help_contract_test.go` enumerates public help contracts without any `coach` entry. Any addition of `namba coach` command registration, usage text, or help-contract coverage would be a regression for this SPEC.
- The `core_harness_change` route is correctly represented after the revisions. `harness-request.json` sets `request_kind` to `core_harness_change`, `touches_namba_core` to `true`, and `adaptation_mode` to `extend_domain`. That aligns with the actual implementation path: this work belongs in Namba core managed renderers, registries, and generated docs. The readiness route remaining `namba plan` is appropriate; this should not be redirected into `$namba-create`, a user-authored skill path, or widened `namba harness` semantics.
- The remaining engineering risk is consistency drift across repeated generated surfaces. The revised plan and eval plan name the right touch points in `internal/namba/templates.go` and `internal/namba/readme.go`, including `renderAgents()`, `renderNambaSkill()`, `renderNambaSkillCommandMappingSection()`, `renderNambaSkillExecutionRulesSection()`, the Codex usage sections, and generated README/workflow command-skill sections. Implementation should keep the wording disciplined so `$namba-help` stays explanation/docs-first while `$namba-coach` stays current-goal-to-next-command coaching.

## Decisions

- Keep the engineering review at `clear`.
- Keep the slice as a managed repo-skill addition only: `$namba-coach` exists, `namba coach` does not.
- Treat dual registration, regen preservation, and regen-first sequencing as acceptance-critical behavior, not incidental implementation detail.

## Follow-ups

- Add one explicit coach-specific regen cleanup regression near the existing regen/managed-output tests.
- Keep generated-doc assertions strong enough to catch partial rollout across repeated command-skill lists and language variants.
- Leave doctor/help expansion out of scope unless the contract is explicitly widened later.

## Recommendation

- Clear for implementation. The revised artifacts now correctly anchor regen vs sync semantics, managed registry/template responsibilities, cleanup regression coverage, the no-public-CLI boundary, and the `core_harness_change` route. Land this as a Namba core registry/template/doc update with regen-first validation.
