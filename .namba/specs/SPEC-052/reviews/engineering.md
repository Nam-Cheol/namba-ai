# Engineering Review

- Status: approved
- Last Reviewed: 2026-05-20
- Reviewer: namba-planner
- Command Skill: `$namba-plan-eng-review`
- Recommended Role: `namba-planner`

## Focus

- Lock architecture, sequencing, failure modes, trust boundaries, and validation strategy before execution starts.

## Findings

- The revised SPEC now points at the live command surface correctly. `internal/namba/namba.go` owns top-level dispatch and `runProject`, `internal/namba/self_update_command.go` owns `namba update`, and `internal/namba/update_command.go` is scoped to regen or managed-output ownership when generated Codex guidance changes.
- The project evidence gap is now resolved at the planning layer. The SPEC explicitly introduces `.namba/logs/project/codex-diagnostics-evidence.json` with schema `project-codex-diagnostics-evidence/v1`, and it assigns persistence ownership to `runProject` while keeping root-comparison inputs in `internal/namba/project_analysis.go`.
- The queue handoff artifact is now explicit enough for implementation. The canonical queue-owned evidence path is `.namba/logs/runs/<spec-id-lower>-queue-evidence.json` with schema `queue-runner-evidence/v1`, which matches the current queue handoff lane already used beside the normal run evidence manifest.
- The generated-doc target is now corrected. `.namba/codex/README.md` is named as the runtime guidance surface for evidence inspection, while `.namba/codex/output-contract.md` stays limited to response-shape guidance unless the final response contract itself changes.

## Decisions

- Treat the reusable Codex diagnostics payload as the source contract and wire project, run, queue, and update/version surfaces onto it afterward. That sequencing minimizes drift across evidence JSON, CLI advice, and generated docs.
- Keep Codex diagnostics advisory and field-granular. Missing Codex, version parse failure, doctor failure, timeout, unavailable workspace roots, and unavailable permission profile should remain structured states rather than workflow blockers.
- Extend queue evidence conservatively so the new diagnostics payload or queue-safe summary does not break the existing queue resume and desktop handoff assumptions around `queue-runner-evidence/v1`.

## Follow-ups

- [non-blocking] Add explicit command-level tests for the new project evidence artifact writer and for queue evidence carrying diagnostics pointers or summary data in addition to current validation fields.
- [non-blocking] After implementation, confirm generated Codex docs are refreshed through the managed output path and not by editing `.namba/codex/README.md` directly.

## Recommendation

- Approved for implementation. The live command surface, project evidence artifact, queue canonical evidence path, and generated-doc target are now clear enough to execute the feature with normal Go validation plus `namba sync` after implementation.
