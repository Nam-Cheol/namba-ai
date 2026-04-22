# SPEC-032 Baseline Evidence

## Purpose

Capture the current repository state before introducing a typed harness classification contract.

## Structured Today

| Surface | What is already structural | Evidence |
| --- | --- | --- |
| Planning start | shared planning-start resolver, worktree policy, SPEC allocation | `internal/namba/planning_start.go` |
| Direct artifact creation | typed create request, preview, overwrite, and validation contract | `internal/namba/create_engine.go` |
| Review scaffolds | typed review templates and readiness summary generation | `internal/namba/spec_review.go` |
| Public command surfaces | separate `plan`, `harness`, `fix`, `worktree`, and internal `__create` entrypoints | `internal/namba/namba.go`, `internal/namba/create_adapter.go` |

## Prose-Owned Today

| Surface | Harness meaning stored mainly in prose | Evidence |
| --- | --- | --- |
| Harness SPEC scaffold | harness problem frame, plan shape, and acceptance meaning | `internal/namba/namba.go` via `buildHarnessSpecDoc(...)`, `buildHarnessSpecPlanDoc(...)`, `buildHarnessAcceptanceDoc(...)` |
| Skill routing guidance | when to use `namba plan`, `namba harness`, `$namba-create` | `.agents/skills/namba-plan/SKILL.md`, `.agents/skills/namba-harness/SKILL.md`, `.agents/skills/namba-create/SKILL.md` |
| Generated docs | repeated command-choice explanations | `README.md`, `docs/workflow-guide.md`, `internal/namba/templates.go`, `internal/namba/readme.go` |

## Current Gaps

- There is no single typed model for harness request classification comparable to `createRequest`.
- There is no explicit field for "this request touches Namba core" versus "this is a domain harness extension."
- There is no standard harness evidence pack beyond the existing optional `contract.md` and `baseline.md` readiness anchors.
- There is no table-driven route contract shared by `namba plan`, `namba harness`, and `$namba-create`.
- The current readiness model is structurally useful, but harness-specific validator and eval expectations are not yet encoded.

## Planning Implication

The first slice should not start by inventing new runtime modes. It should first normalize classification, evidence, and route boundaries so later validator and eval work has a stable target.
