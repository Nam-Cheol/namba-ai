# Engineering Review

- Status: approved
- Last Reviewed: 2026-05-14
- Reviewer: namba-planner
- Command Skill: `$namba-plan-eng-review`
- Recommended Role: `namba-planner`

## Focus

- Lock architecture, sequencing, failure modes, trust boundaries, and validation strategy before execution starts.

## Findings

- The revised contract pins a repeatable per-asset schema strongly enough for implementation. `spec.md`, `acceptance.md`, and `frontend-brief.md` align on exact per-asset labels for both the manifest and `## Generated Asset Evidence`.
- The pre-run enforcement boundary is correct. The brief owns `Frontend implementation phase`, `Asset mode`, `Imagegen requirement`, and `Asset decision proof`; runtime enforcement keys off `Imagegen requirement: required`.
- The generation-plan versus run-evidence split is implementation-ready. `Generation plan status: ready` belongs in the pre-run brief, while generated files, saved paths, and rendered proof belong in run-result evidence.
- The deterministic test strategy covers first-frontend defaults, mixed mode, multi-asset parsing, small-asset evidence, missing per-asset proof, rejected substitutes, and avoids live imagegen/network dependency.
- Generated-surface ownership should stay disciplined: update template sources first, regenerate managed `.agents` and `.codex` artifacts, then sync project artifacts only when needed.

## Decisions

- Treat per-asset repeatability as a hard parser and validator contract, not prose guidance.
- Keep imagegen necessity decisions in the frontend brief validator. `execution.go` should enforce evidence against that pre-run decision.
- Treat `Generation plan status: ready` as pre-run readiness only. Completed generation belongs exclusively in `## Generated Asset Evidence`.
- Treat template renderers as the source of truth for managed skill and role-card text.

## Follow-ups

- During execution, verify the managed-surface path covers template content and scaffold wiring for `.agents/skills/namba-run/SKILL.md`, `.codex/agents/namba-designer.*`, and `.codex/agents/namba-frontend-implementer.*`.
- Preserve deterministic failure messages for missing saved asset paths and missing rendered usage evidence.
- Run `namba regen` after template updates, and run `namba sync` only if readiness or support docs need refresh.

## Recommendation

- Approved for implementation. The revised SPEC is specific enough on per-asset schema, pre-run imagegen enforcement, readiness-versus-evidence separation, and deterministic testability.
