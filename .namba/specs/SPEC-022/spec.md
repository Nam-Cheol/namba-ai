# SPEC-022

## Problem

`namba plan` is currently grounded on repository context that is too shallow to produce consistently strong, implementation-ready SPECs for real multi-app or multi-runtime codebases.

Local evidence in the current repository shows the gap clearly:

- `.agents/skills/namba-plan/SKILL.md` only tells the planner to create the next feature SPEC, prefer the CLI, seed review artifacts, and keep the scope concrete. It does not define noise exclusion, source-priority rules, evidence/confidence, conflict reporting, multi-app separation, or thin-output handling.
- `internal/namba/namba.go` `runProject` currently builds `.namba/project/*` from a README copy plus a small set of helper generators, so the planning surface depends on documentation that is intentionally shallow.
- `internal/namba/namba.go` `buildStructureDoc` emits a depth-limited tree with a short hard-coded skip list, which means `.gitignore`, build outputs, caches, IDE state, downloaded runtimes, vendored code, and logs are not treated as a first-class configurable scoping policy.
- `internal/namba/namba.go` `buildTechDoc`, `buildCodemaps`, `buildEntryPointsDoc`, and `buildDependenciesDoc` generate high-level summaries, but they do not extract the richer facts needed for planning: app boundaries, route surfaces, background jobs, module responsibilities, state stores, auth boundaries, runtime constraints, deployment structure, test coverage shape, or code-vs-doc drift.
- The current generated outputs in `.namba/project/structure.md`, `.namba/project/tech.md`, and `.namba/project/codemaps/*.md` confirm the result: they are helpful as a snapshot, but too thin to serve as reliable planning intelligence for the kind of improvement axes requested here.

As a result, `namba plan` cannot yet do the following at the required quality bar:

- separate signal from repository noise before analysis starts
- prioritize code/configuration over weaker documents
- infer application type, entry points, modules, state, auth, integrations, and operating constraints in a framework-agnostic way
- report code-vs-document mismatches explicitly instead of flattening them into a neutral summary
- fail or warn when the generated docs are obviously too thin for the repository size
- adapt the same planning quality bar across Go, Java/Spring, Node/Nest, Python/FastAPI, Rails, React/Vite, and mixed-system repositories

## Goal

Redesign the analysis pipeline that feeds `namba plan` so planning is grounded on evidence-based project intelligence rather than shallow tree summaries, with explicit input scoping, semantic extraction, mismatch reporting, and analysis quality gates.

## Target Reader

- Namba maintainers who need `namba project` output to be reliable enough to base a new SPEC on it.
- Codex operators who need to decide quickly what the repository actually is, where it starts, how it is deployed, and what is risky before `namba plan` or `namba run`.
- Maintainers of multi-app or mixed-runtime repositories who cannot use a single flat tree dump as planning context.

## Context

- Project: namba-ai
- Project type: existing
- Language: go
- Mode: tdd
- Work type: plan
- Planning surface under improvement: `$namba-plan`, `namba plan`, and the project-analysis context that `namba plan` depends on
- Primary implementation area: `internal/namba/namba.go`, new analyzer-focused internals under `internal/namba/`, `.agents/skills/namba-plan/SKILL.md`, `.agents/skills/namba-project/SKILL.md`, generated docs, and any new `.namba/config/sections/*` config needed for repository-specific analysis rules
- Existing artifacts to evolve rather than discard: `.namba/project/product.md`, `.namba/project/tech.md`, `.namba/project/structure.md`, `.namba/project/codemaps/*.md`

## Desired Outcome

- `namba project` uses a configurable analysis scope that excludes noise by default and elevates high-signal inputs such as `src/**`, tests, build config, deploy config, CI, infra, and current design docs.
- The analyzer detects app boundaries before summarization, so monorepos and mixed repositories are represented as separate systems instead of one flattened tree dump.
- Every generated conclusion can cite supporting files and carry an explicit confidence level, with weaker inferences labeled as such.
- Source priority is explicit and stable: executable code and authoritative config outrank tests, tests outrank build/deploy config, build/deploy config outranks maintained docs, and maintained docs outrank stale planning material.
- Code-vs-document conflicts are surfaced as first-class findings rather than silently averaged into prose.
- The generated project docs cover:
  - project purpose and primary user flow
  - runtime and deployment topology
  - module map and responsibility boundaries
  - entry points and externally visible interfaces
  - data and state model summary
  - authentication, authorization, and security mechanisms
  - test map, validation commands, and obvious coverage gaps
  - mismatch and drift report
  - structure listing as appendix material rather than the main document
- Analysis success and analysis quality are treated separately, so thin outputs produce warnings or failures even if files were technically generated.
- The architecture cleanly separates a framework-agnostic analyzer core from language/framework adapters.
- The v1 release is a `planning context foundation release`: it fixes scoping, evidence/confidence/conflict handling, mismatch reporting, document hierarchy, output compatibility, and quality-gate behavior before broad adapter rollout.
- The new core is capable of extracting, at minimum, application type, entry points, layer boundaries, state stores, auth model, external integrations, runtime/deploy constraints, and test/ops risk signals for the current repository class.
- Repository-specific config can tune include paths, exclude paths, source priority, key roles, important runtimes/services, and output templates without forcing every repository into one fixed shape.

## V1 Success Definition

- A maintainer can run `namba project` on a repository with more than one major system boundary and receive a primary landing document plus per-system summaries that are materially more actionable than the current `product.md` plus shallow codemap output.
- A maintainer can tell which claims are directly evidenced, which are inferred, and which are in conflict with existing docs.
- Thin analysis output is visible as a CLI quality signal instead of silently looking successful.
- The current `.namba/project/*` entry points remain understandable and compatible during the transition.

## Implementation Priority

1. Noise exclusion and input scoping
2. Semantic extraction for entry points, modules, state, auth, integrations, and operating structure
3. Code-vs-document mismatch reporting
4. Analysis quality gates and thin-output detection
5. Language/framework adapter separation and expansion

## Scope

- Replace the current shallow `namba project` analysis path with an evidence-backed analyzer pipeline.
- Introduce a repository-analysis config surface under `.namba/config/sections/` for include/exclude rules, source priority, important roles, important runtimes/services, and output template selection.
- Add a framework-agnostic analyzer core responsible for:
  - file classification
  - repository and app-boundary detection
  - source-priority handling
  - evidence/confidence attachment
  - conflict detection
  - quality-gate evaluation
- Add adapter hooks for stack-specific signal extraction so framework knowledge lives outside the core analyzer.
- Ship a Go-first implementation and a generic adapter seam that proves how stack-specific extraction plugs into the core analyzer without forcing broad first-party adapter rollout into the same delivery unit.
- Expand `.namba/project/*` output so the generated document set includes richer planning artifacts instead of relying mainly on `product.md`, `tech.md`, `structure.md`, and a few codemap summaries.
- Preserve `structure.md` as a secondary appendix-style artifact rather than the main planning document.
- Update `$namba-plan` and `$namba-project` skill guidance so planning instructions consume evidence, honor source priority, preserve conflicts, and recognize multi-app boundaries.
- Add regression coverage for scoping, semantic extraction, mismatch reporting, adapter dispatch, and thin-output warnings/failures.

## Minimum Analysis Contract

- `evidence`
  - Every non-trivial generated claim must include at least one supporting file reference.
  - When a claim depends on multiple signals, the output can attach more than one evidence reference.
- `confidence`
  - `high`: directly supported by executable code or authoritative runtime/build/deploy configuration.
  - `medium`: supported by multiple weaker sources such as tests plus configuration, or consistent heuristics across code paths.
  - `low`: partial or heuristic inference that should not be phrased as established fact.
- `conflict`
  - A conflict record must identify the contested claim, the stronger source, the weaker or stale source, and a short reason why they disagree.
  - Conflicts must be reported explicitly, not flattened into a neutral summary sentence.

## Output Compatibility And Document Hierarchy

- Compatibility strategy
  - Extend the current `.namba/project/*` surface rather than replacing it in one step.
  - `product.md` remains the primary landing document for human readers in v1.
  - `tech.md` remains a canonical technical hub and links to deeper per-system artifacts.
  - `structure.md` remains generated but is explicitly appendix material.
  - Existing `codemaps/*.md` remain supporting artifacts until later consolidation.
- Reading hierarchy
  - First landing document: `.namba/project/product.md`
  - Technical hub: `.namba/project/tech.md`
  - System-by-system summaries follow a fixed repeatable order:
    - system purpose
    - entry points and interfaces
    - module boundaries
    - data and state
    - auth/integrations
    - deploy/runtime/test risks
  - mismatch reporting and evidence references must be surfaced from the primary docs, with detailed support artifacts linked rather than hidden entirely in appendices.

## Quality Gate Contract

- `namba project` still emits generated docs when analysis completes, even if quality warnings are present, so the operator can inspect what was thin.
- Exit `0` with warnings when the analyzer completes but detects non-critical thin output, low-confidence concentration, or partial coverage gaps.
- Exit non-zero when required foundation artifacts cannot be produced, the minimum evidence/confidence/conflict contract is violated, or critical thin-output thresholds are hit.
- Quality-gate output must tell the operator which systems or artifacts were too thin and why.

## Non-Goals

- Do not treat a successful file write as proof that the analysis is good enough.
- Do not keep the current flat output model as the only project-document surface.
- Do not attempt a perfect, exhaustive framework implementation for every ecosystem in the same first cut; deep first-party adapters for Java/Spring, Node/Nest, React/Vite, Python/FastAPI, and Rails belong to a follow-up phase after the foundation release lands.
- Do not fold conflicting code and document claims into a single ambiguous sentence.
- Do not make full Git diff-driven incremental refresh the critical path for this first implementation; capture the design seams so `sync` and `project` can adopt incrementality cleanly in follow-up work.

## Design Constraints

- Keep `.namba/` as the source of truth for generated planning context.
- Treat executable code and authoritative configuration as stronger evidence than prose documentation.
- Mark inferred conclusions explicitly when direct evidence is incomplete.
- Make multi-app and multi-runtime repositories readable by splitting summaries per system boundary.
- Keep the output useful for future SPEC planning, not just for one-off repository descriptions.
- Prefer renderer/source changes over hand-patching generated docs.
