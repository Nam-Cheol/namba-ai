# SPEC-032

## Goal

Establish a typed harness classification and adaptation contract so NambaAI can distinguish:

1. core harness changes
2. domain or user-requested harness changes
3. direct artifact creation

without relying primarily on scaffold prose or command-name heuristics.

## Command Choice Rationale

- This request changes platform routing and evidence contracts across `namba plan`, `namba harness`, and `$namba-create`.
- It is therefore a feature slice on Namba's planning and contract surface, not a direct domain-harness design request and not a direct artifact-generation request.

## Verified Local Context

- `internal/namba/namba.go` exposes separate `plan`, `harness`, and `fix` planning entrypoints, but harness-specific meaning is still embedded in `buildHarnessSpecDoc(...)`, `buildHarnessSpecPlanDoc(...)`, and `buildHarnessAcceptanceDoc(...)`.
- `internal/namba/create_engine.go` already has a typed direct-generation contract through `createRequest`, `createPreview`, and `createTarget`.
- `internal/namba/spec_review.go` already has structured review tracks and optional evidence anchors for `contract.md` and `baseline.md`, but it does not yet define a harness-specific evidence pack or eval contract.
- `internal/namba/planning_start.go` already defines the adjacent planning and worktree-start contract that this slice must preserve rather than relitigate.
- `.agents/skills/namba-plan/SKILL.md`, `.agents/skills/namba-harness/SKILL.md`, `.agents/skills/namba-create/SKILL.md`, `README.md`, and `docs/workflow-guide*.md` repeat route-selection guidance largely in prose.

## Problem

NambaAI currently overloads "harness" across three different layers:

1. NambaAI-owned core harness work
2. user-requested or domain-specific harness additions
3. direct creation of skill, agent, or workflow artifacts

That overload is manageable only while the system is small. The current repository already shows a structural imbalance:

- direct artifact creation has typed request and preview models in `internal/namba/create_engine.go`
- planning and worktree start already have typed helpers in `internal/namba/planning_start.go`
- review scaffolds already have typed templates plus evidence slots in `internal/namba/spec_review.go`
- harness intent itself is still described primarily by scaffold prose, skill wording, and README guidance

The result is growing platform risk:

- semantic drift: repeated docs and skills can redefine harness meaning differently
- routing ambiguity: `namba plan`, `namba harness`, and `$namba-create` do not share one typed decision contract
- validator gap: there is no canonical metadata model a classifier or validator can inspect
- evidence gap: harness work does not yet require a standard evidence pack strong enough for real review or eval
- review ambiguity: current readiness is advisory, but harness-specific acceptance does not yet bind to concrete classifier or evidence expectations

## Desired Outcome

- Harness-related requests are classified consistently into `core_harness_change`, `domain_harness_change`, or `direct_artifact_creation`.
- `namba plan`, `namba harness`, and `$namba-create` have non-overlapping role boundaries derived from one shared typed contract rather than from command-name folklore.
- Harness planning moves from prose-first scaffolds to metadata plus evidence.
- Future harness validators, eval packs, and dedicated reviewer roles can attach to one stable contract without redesigning `namba run` semantics.
- Operators can determine the correct entrypoint from a compact decision contract plus canonical examples, not only from reviewer interpretation.

## Scope

- Inventory the current command, skill, runtime, and review surfaces that participate in harness interpretation.
- Separate what is already structural from what still exists only as prose.
- Define a minimal typed harness request model and an adaptation contract that differentiates:
  - existing core harness modification
  - new domain harness addition or adaptation
  - direct artifact creation
- Redefine role boundaries for `namba plan`, `namba harness`, and `$namba-create` around that model.
- Define the standard harness evidence artifact set:
  - `contract.md`
  - `baseline.md`
  - `eval-plan.md`
  - `harness-map.md` when the work adapts or composes a domain harness
- Define acceptance criteria that can drive validator, review, and eval implementation later.
- Preserve the existing `.namba/specs/<SPEC>` model, planning-start contract, review scaffold flow, and `namba run` execution-mode semantics.

## Proposed Contract

### 1. Canonical Request Classes

- `core_harness_change`
  - Affects Namba-owned planning, routing, review, validator, runtime, or built-in skill/agent surfaces.
  - Default entrypoint: `namba plan`.
- `domain_harness_change`
  - Adds or adapts a reusable user or domain harness on top of the existing Namba core contract without changing Namba-owned core semantics.
  - Default entrypoint: `namba harness`.
- `direct_artifact_creation`
  - The user's goal is the artifact itself under `.agents/skills/*`, `.codex/agents/*`, or equivalent direct output, and a reviewable harness SPEC is not the primary deliverable.
  - Default entrypoint: `$namba-create`.

### 2. Minimal Typed Metadata

Add one canonical typed metadata object for harness-classified work, but split the transport explicitly:

- `delivery_mode=spec`
  - persist the typed object at `.namba/specs/<SPEC>/harness-request.json`
- `delivery_mode=direct`
  - carry the same object transiently through the existing `namba __create preview|apply` JSON adapter and any associated preview/apply artifacts
  - do not invent a fake SPEC path for direct creation in v1

Use JSON in v1, not YAML. That keeps the first implementation aligned with the repo's existing JSON-based create adapter and avoids assuming nested YAML support where the current config parser is still flat `key: value`.

Minimum required fields:

- `request_kind`
  - `core_harness_change | domain_harness_change | direct_artifact_creation`
- `delivery_mode`
  - `spec | direct`
- `adaptation_mode`
  - `modify_core | extend_domain | compose_domain | generate_artifact`
- `base_contract_ref`
  - reference to the core or domain harness contract being adapted; empty only when genuinely new
- `touches_namba_core`
  - boolean guardrail for routing and review escalation
- `artifact_targets`
  - normalized list such as `skill`, `agent`, `workflow`, `validator`, `eval-pack`, or `docs`
- `required_evidence`
  - normalized artifact list; at minimum `contract`, `baseline`, `eval-plan`, plus `harness-map` when a domain harness adapts or composes an existing harness, or when the route boundary with direct artifact creation needs explicit documentation
- `required_reviews`
  - normalized review profile for v1: `product`, `engineering`, `design`

### 2A. Operator Cheat Sheet

- Changing Namba itself or its built-in harness contract -> `namba plan`
- Adding or adapting a reusable domain harness without changing Namba core semantics -> `namba harness`
- Creating the skill or agent artifact directly, with no unresolved harness-contract change -> `$namba-create`

### 3. Decision Contract

- Route to `namba plan` when the request changes Namba-owned core harness behavior, planning or routing semantics, built-in command-entry skills, generated docs, validator rules, or execution and review contracts.
- Route to `namba harness` when the request introduces or adapts a reusable domain harness but preserves Namba's core harness contract.
- Route to `$namba-create` when the primary outcome is direct artifact creation and the target is clear enough to preview and apply without a harness SPEC.
- If a request mixes direct artifact creation with unresolved core-contract change, force `namba plan` first or split the work into separate packages; do not let `$namba-create` silently redefine the harness contract.
- If a request starts as `domain_harness_change` but `touches_namba_core=true` becomes true during planning, escalate it to `core_harness_change` instead of keeping it under domain scope.

### 3A. Canonical Route Examples

1. "Redefine how built-in `namba-harness` chooses its template and readiness requirements."
   - class: `core_harness_change`
   - route: `namba plan`
2. "Add a reusable fintech-analysis harness that teams can plan against, without editing built-in Namba routing."
   - class: `domain_harness_change`
   - route: `namba harness`
3. "Create a repo-local release-triage skill and paired custom agent for my project."
   - class: `direct_artifact_creation`
   - route: `$namba-create`
4. "Create a new skill, and also change Namba's built-in command-selection guidance so this new pattern becomes part of core routing."
   - class: `core_harness_change`
   - route: `namba plan`
   - note: split direct generation into a later step if needed
5. "Adapt an existing domain harness so it composes two reusable domain workflows."
   - class: `domain_harness_change`
   - route: `namba harness`
6. "I only know I want a skill and agent for X, but I am not sure whether I need a new reusable harness contract first."
   - class: unresolved
   - route: prefer `namba plan` until the contract/artifact boundary is explicit

### 4. Evidence Contract

- `contract.md`
  - canonical classification, invariants, entrypoint decision, and field definitions
- `baseline.md`
  - current-state inventory, with explicit separation between structured state and prose-owned behavior
- `eval-plan.md`
  - classifier fixtures, validator expectations, should-route and should-not-route cases, and review or eval hooks
- `harness-map.md`
  - required when a domain harness adapts or composes an existing harness, or when the route boundary with direct artifact creation needs explicit documentation; maps the relationship between core harness, domain harness, and direct artifact outputs

### 4A. Compatibility And Enforcement

- Existing SPECs without `harness-request.json` remain on the legacy readiness path. This slice must not retroactively require new evidence for unrelated historical SPECs.
- `extraction-map.md` remains a valid legacy evidence artifact where earlier SPECs already use it.
- New harness evidence requirements apply only to harness-classified slices that opt into the typed contract.
- V1 validator ownership is explicit:
  - a narrow harness-evidence validator helper is invoked during review-readiness refresh and `namba sync` for harness-classified SPECs
  - it stays advisory in readiness output in v1 rather than redefining `namba run` execution validation
  - regression tests exercise the same helper directly for evidence completeness and route-boundary guarantees

### 5. Acceptance Strategy

This SPEC should not stop at guidance text. It must define an implementation path where:

- route selection can be table-driven and regression-tested
- route-choice guidance is concrete enough that operators can choose an entrypoint from examples and the decision contract
- missing required evidence can be validator-detectable
- review readiness can consume harness evidence without breaking the current generic review scaffold
- future harness-specific eval packs can bind to stable metadata instead of scraping prose

## Migration Approach

1. Baseline the current surfaces and record which parts are structural versus prose-only.
2. Introduce the typed harness request model inside the existing `.namba/specs/<SPEC>` artifact model.
3. Bind `namba plan`, `namba harness`, and `$namba-create` to one shared decision contract while keeping their user-facing purpose distinct.
4. Require the harness evidence pack for harness-classified work, but preserve backward compatibility with current review and readiness flows by keeping `contract.md` and `baseline.md` as the first compatibility anchors.
5. Add validator and eval fixtures plus dedicated role follow-ups in later slices, without reopening `namba run` mode semantics.

## Non-Goals

- Do not implement every future harness-specific role, validator, or runtime mode in this slice.
- Do not reduce the change to wording cleanup in README or skill prose only.
- Do not redesign `namba run`, `--solo`, `--team`, or `--parallel` semantics beyond referencing them as existing downstream contracts.
- Do not create a second planning artifact model outside `.namba/specs/<SPEC>/`.
- Do not let direct artifact generation become a backdoor for changing Namba's core harness contract without a reviewable SPEC.
- Do not generalize the review runtime beyond `product`, `engineering`, and `design` in v1 just to make `required_reviews` look future-proof.

## Design Constraints

- Keep `.namba/` as source of truth and preserve the existing review scaffold under `.namba/specs/<SPEC>/reviews/`.
- Preserve the safe-by-default planning and worktree contract defined by the current planning-start surface.
- Prefer repo code and authoritative config over generated prose when source and docs diverge.
- Keep the first typed model intentionally small; add fields only when they change routing, evidence, or review behavior.
- Keep evidence artifacts human-readable first, but structured enough that a future validator can check presence, completeness, and compatibility.
- Preserve backward-compatible readiness behavior: new harness evidence should layer onto current review artifacts rather than replacing them outright.
- Keep transport unified around JSON in v1 so SPEC routes and direct-create routes can share one Go model without adding a second parser surface first.
