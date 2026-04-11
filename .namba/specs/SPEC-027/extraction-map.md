# SPEC-027 Phase-1 Extraction Map

## Landed In Phase-1

- Runtime contract and request-building moved behind [`internal/namba/runtime_contract.go`](/mnt/c/study/mo-ai/namba-ai/internal/namba/runtime_contract.go)
- Preflight execution stayed isolated in [`internal/namba/runtime_harness.go`](/mnt/c/study/mo-ai/namba-ai/internal/namba/runtime_harness.go)
- Init-time repository scanning consolidated in [`internal/namba/init_scan.go`](/mnt/c/study/mo-ai/namba-ai/internal/namba/init_scan.go)

## Landed In The Next Slice

- Project-analysis inventory is now an explicit seam inside [`internal/namba/project_analysis.go`](/mnt/c/study/mo-ai/namba-ai/internal/namba/project_analysis.go)
  - file collection and system-root discovery are captured once in `analysisInventory`
  - system detection, conflict evidence, rendering, and quality evaluation now reuse that inventory instead of each path re-deriving the same repository view
  - conflict detection and quality-gate heuristics now run through explicit heuristic inputs instead of reading the full rendered `projectAnalysis` struct
  - quality-gate warnings now consume summarized system/conflict metrics instead of reading raw discovery structures directly
  - conflict heuristics now consume summarized runtime signals/evidence instead of reading inventory and systems directly inside the conflict detector
  - conflict heuristics now consume a precomputed runtime-support bundle, so runtime signal collection and evidence derivation stay outside `detectAnalysisConflicts`
- Parallel-run lifecycle now has an explicit prepare/stage/execute/finalize seam across [`internal/namba/parallel_run.go`](/mnt/c/study/mo-ai/namba-ai/internal/namba/parallel_run.go) and [`internal/namba/parallel_lifecycle.go`](/mnt/c/study/mo-ai/namba-ai/internal/namba/parallel_lifecycle.go)
  - git/runtime preflight and worker staging are isolated behind `prepareParallelRunPlan` and `stageParallelWorkers`
  - worker execution stays centralized in `parallelRunLifecycle`
  - merge blocking, merge execution, cleanup aggregation, and finished-report persistence now move through dedicated lifecycle helpers instead of one inline finalize block
  - worker prompt shaping and execution request construction now move through dedicated worker helpers instead of being duplicated between staging and execution
  - worker execution outcomes are now collected separately from shared result mutation, so goroutines no longer write directly into lifecycle state
  - terminal report writing is now orchestrated from `finalize`, while merge/block helpers only mutate lifecycle state and return errors
  - cleanup artifact handling now returns a cleanup outcome that report persistence applies separately, instead of cleanup helpers constructing report state inline
  - finished-report snapshots are now built before persistence, so the report writer only serializes the final snapshot instead of mutating report state itself
  - worker-artifact cleanup and `git worktree prune` execution now run through separate cleanup-phase helpers before their outcomes are recomposed
  - cleanup-phase aggregation now moves through `completeCleanupPhase`, so `finalize` no longer inlines cleanup outcome recomposition and report shaping
- Run-command execution setup in [`internal/namba/namba.go`](/mnt/c/study/mo-ai/namba-ai/internal/namba/namba.go) now has explicit `load -> materialize -> dispatch` helpers
  - top-level command routing and public command-usage lookup now reuse shared command definitions instead of keeping separate command-name switches in `Run` and `commandUsageText`
  - top-level `Run`/`help` handling now resolves through one invocation helper instead of parsing help, resolving usage text, and dispatching commands inline
  - top-level `usageText()` summary now reuses the same public command definitions instead of carrying a separate hard-coded command list
  - `worktree` subcommand routing now reuses shared subcommand definitions instead of keeping `new/list/remove/clean` inline inside one switch
  - `worktreeUsageText()` now reuses the same shared subcommand definitions instead of carrying a second hard-coded subcommand list
  - `fix --command` routing now reuses shared fix subcommand definitions instead of keeping `plan/run` inline inside `runFix`
  - `fixUsageText()` now reuses shared fix subcommand behavior summaries instead of carrying a separate nested command behavior block
  - `fixUsageText()` now also reuses the same shared flag-like usage-line helper pattern as `plan`/`harness` instead of carrying its own example lines
  - `planUsageText()` and `harnessUsageText()` now reuse one shared description-command usage builder instead of carrying duplicate help bodies
  - single-usage-line command help blocks now reuse one shared builder instead of each command carrying the same header/usage/behavior layout inline
  - top-level no-arg commands now reuse one shared help/extra-arg preamble helper instead of each handler repeating the same `--help` and `does not accept arguments` contract
  - config/spec loading and prompt construction move through `loadRunExecutionContext`
  - prompt markdown materialization moves through `materializeRunExecutionPrompt`
  - mode-based execution dispatch moves through `dispatchRunExecution`
  - `run` and `direct-fix` now share execution runtime loading and prompt write helpers instead of each path reloading quality/system/codex config separately
  - `direct-fix` now also has explicit `load -> materialize -> dispatch` helpers so prompt shaping, runner dispatch, and post-run sync no longer live inline in `executeDirectFix`
  - direct-fix prompt assembly is now split into repair-contract, project-context, and validation section helpers instead of one inline string-builder block
  - spec-package scaffolding now moves through `loadSpecPackageScaffoldContext`, `buildSpecPackageScaffoldOutputs`, and `materializeSpecPackageScaffoldOutputs` instead of one inline `createSpecPackage` block
  - `buildSpecDoc` and `buildSpecPlanDoc` now route through kind-local builders instead of one switch carrying every kind's full text inline
  - `buildSpecAcceptanceDoc` now routes through kind-local acceptance section helpers instead of keeping each mode branch inline in the builder bodies
- Sync-command support-doc orchestration in [`internal/namba/namba.go`](/mnt/c/study/mo-ai/namba-ai/internal/namba/namba.go) now separates output planning from output materialization
  - `buildSyncProjectSupportOutputs` owns `.namba/project/*.md` content and path assembly
  - `materializeSyncProjectSupportOutputs` remains the thin bridge to `writeOutputs`, preserving manifest/session-refresh behavior
  - `buildChangeSummaryDoc` and `buildReleaseNotesDoc` now assemble document-local section helpers instead of carrying all section text inline
  - `buildPRChecklistDoc` and `buildReleaseChecklistDoc` now assemble document-local checklist helpers instead of carrying all checklist text inline
- Template assembly in [`internal/namba/templates.go`](/mnt/c/study/mo-ai/namba-ai/internal/namba/templates.go) now routes the first planning-role card/custom-agent pairs through shared role profiles
  - `planner`, `plan-reviewer`, and `product-manager` now keep role-card wording and custom-agent instruction bodies in one profile definition while assembly stays shared
  - `frontend-architect`, `designer`, and `mobile-engineer` now keep role-card wording and custom-agent instruction bodies in one profile definition while assembly stays shared
  - `backend-architect` and `reviewer` now keep role-card wording and custom-agent instruction bodies in one profile definition while assembly stays shared, while preserving the narrower wording differences between role-card and custom-agent responsibilities where they matter
  - `backend-implementer`, `data-engineer`, `security-engineer`, `test-engineer`, `devops-engineer`, and `implementer` now route through shared workspace-write role profiles, while preserving the narrower role-card vs custom-agent wording split for `test-engineer`
  - `renderNambaSkill` now assembles its command-mapping and execution-rules sections through dedicated local helpers instead of carrying that long-form skill body inline
  - `renderCodexUsage` now assembles its init-enables and how-Codex-uses front sections through dedicated local helpers instead of carrying that front long-form body inline
  - `renderCodexUsage` now assembles its workflow-command-semantics section through a dedicated local helper instead of carrying that contract-heavy block inline
  - `renderCodexUsage` now assembles its agent-roster, delegation-heuristics, and plan-review-readiness mid sections through dedicated helpers instead of carrying that middle long-form body inline
  - `renderCodexUsage` now assembles its output-contract, git-collaboration, Claude-mapping, and important-distinction tail sections through dedicated helpers instead of carrying that final long-form body inline
- README rendering in [`internal/namba/readme.go`](/mnt/c/study/mo-ai/namba-ai/internal/namba/readme.go) now starts moving document-local sections behind explicit helpers
  - the Namba CLI getting-started guide now assembles install, bootstrap, basic-flow, and next-docs sections through dedicated section helpers instead of keeping every localized block inline in one renderer
  - `renderReadmeGuidePrelude` now owns the generated-header, localized title, language-link, and doc-link scaffolding for Namba CLI guides instead of each guide rebuilding the same prelude inline
  - localized Namba CLI workflow-guide text remains bounded inside the guide renderer while shared prelude assembly moves into the new helper seam
  - the managed-project getting-started guide now assembles open, refresh-context, work-package, review-readiness, implement, and handoff sections through dedicated section helpers instead of keeping each localized block inline
  - the Namba CLI workflow guide now assembles PR/merge, generated-assets, collaboration-defaults, and release-flow tail sections through dedicated helpers instead of keeping those localized tails inline
  - the Namba CLI workflow guide now assembles its localized update/regen/sync/pr/land command-differences block through a dedicated helper instead of repeating that first lifecycle section inline in each language branch
  - the Namba CLI root README now assembles the localized quick-start block through a dedicated helper instead of keeping install/init/basic-flow text inline in each language branch
  - the Namba CLI root README now assembles the localized command-skills, skill-mapping, and custom-agents blocks through dedicated helpers instead of keeping those contract-heavy sections inline in each language branch
  - the Namba CLI root README now assembles the localized read-more and technical-snapshot tail blocks through dedicated helpers instead of keeping those closing sections inline in each language branch
  - the managed-project root README now assembles the localized work-summary, quick-start, read-more, and current-defaults blocks through dedicated helpers instead of keeping those repeated sections inline in each language branch
  - the managed-project root README now assembles the localized what-you-can-do block through a dedicated helper instead of keeping that repeated task-routing section inline in each language branch
  - the managed-project root README now assembles the English command-skills, skill-mapping, and custom-agents block through dedicated helpers instead of keeping that command-surface section inline in the default branch
  - the managed-project workflow guide now assembles the localized collaboration-rules tail through a dedicated helper instead of keeping the profile-derived handoff block inline in each language branch
  - the managed-project workflow guide now assembles the localized key-locations and work-order sections through dedicated helpers instead of keeping those repeated structure/order blocks inline in each language branch
  - the managed-project workflow guide now assembles the localized run-modes section through a dedicated helper instead of repeating the same execution-mode block inline across language branches
  - the managed-project workflow guide now assembles the localized review-readiness section through a dedicated helper instead of repeating the same advisory review guidance block inline across language branches
  - the managed-project workflow guide now assembles the localized role-routing section through a dedicated helper instead of repeating the same specialist-delegation block inline across language branches
  - the managed-project workflow guide now assembles the localized `namba plan` / `namba fix` section through a dedicated helper instead of repeating that command-selection block inline across language branches
  - the managed-project workflow guide now assembles the English-only planning-commands prelude through a local helper instead of keeping that command-selection block inline in the default branch
  - the managed-project workflow guide now assembles the English-only PR-and-merge-flow tail through a local helper instead of keeping that final handoff block inline in the default branch
  - the managed-project workflow guide now routes localized and English section ordering through a shared assembly helper instead of keeping branch-specific nested append/index chains inline
  - the managed-project workflow guide now assembles its localized and English prelude through a dedicated helper instead of repeating the generated-header, guide-link, and project-intro block inline in each language branch
  - the English managed-project root README now assembles command-skills and custom-agents through local helpers while skill-mapping reuses the Namba CLI root helper because the wording contract matches exactly

## Next Safe Seams

- [`internal/namba/namba.go`](/mnt/c/study/mo-ai/namba-ai/internal/namba/namba.go)
  - command parsing and top-level orchestration beyond the current shared command-definition slice
  - consider shrinking repeated no-arg command entry guards only when the resulting helper stays narrower than the command-specific side effects
  - other nested dispatchers and subcommand help surfaces beyond the current `worktree`/`fix` subcommand-definition slices
  - consider broader usage-line sharing only when command-specific placeholders stay narrow enough to avoid wording drift
  - consider broader planning-command help sharing only if command-specific wording stays narrow enough to avoid drift
  - consider broader simple-help sharing only when command-specific examples and behavior lines still read clearly without over-generalizing the helper surface
  - consider sharing prompt-section helpers across direct-fix and run execution only when the wording contract truly aligns
  - consider sharing scaffold acceptance sections across kinds only when the wording contract truly aligns
  - consider sharing support-doc section helpers only when wording contracts truly align
- [`internal/namba/project_analysis.go`](/mnt/c/study/mo-ai/namba-ai/internal/namba/project_analysis.go)
  - keep heuristic inputs narrow and avoid re-coupling quality gates to renderer-only analysis fields
  - consider splitting conflict heuristics from runtime-evidence collection if project-analysis growth continues
  - consider isolating runtime signal collection from evidence ranking if conflict logic grows beyond the current README mismatch contract
  - consider moving `analysisInventory` into its own file once downstream callers stabilize
- [`internal/namba/parallel_run.go`](/mnt/c/study/mo-ai/namba-ai/internal/namba/parallel_run.go)
  - keep lifecycle entrypoint thin and avoid reintroducing orchestration details
- [`internal/namba/parallel_lifecycle.go`](/mnt/c/study/mo-ai/namba-ai/internal/namba/parallel_lifecycle.go)
  - consider separating merge/report persistence from cleanup artifact handling if lifecycle growth continues
  - consider extracting cleanup-phase orchestration if worker cleanup and prune state need richer reporting
  - consider extracting report persistence if more commands begin sharing run artifacts
- [`internal/namba/readme.go`](/mnt/c/study/mo-ai/namba-ai/internal/namba/readme.go)
  - continue separating README section models from renderer text assembly, starting with the remaining Namba CLI root/workflow renderers and the managed-project root/workflow renderers
- [`internal/namba/templates.go`](/mnt/c/study/mo-ai/namba-ai/internal/namba/templates.go)
  - separate long-form template bodies from assembly helpers
  - extend role-profile extraction only where role-card wording and custom-agent instruction wording stay aligned enough to avoid drift

## Critical Path

1. Keep runtime contract helpers centralized
2. Continue shrinking repeated repository inventory logic
3. Split `parallel_run.go` lifecycle concerns before broader renderer moves
4. Split monolith renderers only after inventory and lifecycle seams are stable

## Deferred From Phase-1

- `project_analysis.go` inventory/render split
- `parallel_lifecycle.go` cleanup artifact extraction and report-writer reuse beyond the current lifecycle seam
- `namba.go` orchestration reduction beyond the run-execution slice
- `namba.go` prompt-section sharing beyond the current direct-fix extraction slice
- `namba.go` cross-kind scaffold-text sharing beyond the current doc/plan/acceptance routing slice
- `namba.go` support-doc helper sharing beyond the current document-local renderer slices
- renderer decomposition in `readme.go` and `templates.go`
- broader `runSync` pipeline redesign
