# SPEC-038

## Problem

NambaAI has first-class command-entry skills such as `$namba-plan`, `$namba-run`, and `$namba-help`, but it does not have a dedicated coaching surface for users who arrive with vague intent or with the wrong Namba command in mind. That gap makes the general `$namba` router and `$namba-help` absorb two different jobs:

- `$namba-help` explains NambaAI usage and command semantics.
- A missing `$namba-coach` should clarify the user's current goal, correct command choice when needed, and hand off to the right workflow invocation.

## Goal

Add `namba-coach` as a first-class, managed, generated repo skill exposed as `$namba-coach`. It should sit alongside `$namba-plan`, `$namba-run`, and `$namba-help` in Codex-facing Namba surfaces while staying read-only and advisory.

## Scope

- Add `renderCoachCommandSkill()` in `internal/namba/templates.go`.
- Expose `$namba-coach` in `renderNambaSkill`, AGENTS text, and Codex usage copy as an official command-entry or guidance skill.
- Add `namba-coach` to the managed skill registry and `codexSkillTemplates()` in `internal/namba/codex.go`.
- Update README, workflow guide, and Codex integration generated docs in `internal/namba/readme.go` and template docs so `$namba-coach` has one consistent role.
- Ensure `namba init` and `namba regen` generate and preserve `.agents/skills/namba-coach/SKILL.md`; `namba sync` should refresh downstream docs, project artifacts, readiness, and manifest state after regen, but must not be treated as the repo-skill backfill path.
- Add focused tests for template content, generated skill registration, docs exposure, and scaffold stability.

## Skill Contract

`$namba-coach` is a read-only advisory skill. It must not create SPEC packages, edit files, generate artifacts, run implementation, or update review readiness directly.

Behavior:

- Restate the user's goal briefly.
- Follow this response order: brief restatement, up to three essential clarification questions if required, one primary executable handoff, optional single alternative when there is a real tradeoff, and a one- or two-sentence reason.
- Ask only 1-3 essential clarification questions when required.
- Treat "essential" as information needed to choose the correct Namba workflow or make the handoff command usable, not information that would fully specify implementation.
- Once the request is concrete enough, recommend exactly one primary command and at most one alternative.
- Present recommendations as directly executable invocations.
- Explain why the selected flow fits in one or two sentences.
- Correct a wrong command choice first instead of running it as-is when another Namba command or skill is clearly better.

Boundary with `$namba-help`:

- `$namba-help` explains how NambaAI works, what commands mean, and where docs live.
- `$namba-coach` uses the user's current idea or question to choose the next workflow handoff.

Routing rules:

- New feature or product change: `namba plan "<description>"`
- Reusable skill, agent, workflow, or orchestration SPEC: `namba harness "<description>"`
- Direct repo-local skill or custom agent creation: `$namba-create`
- Bug repair: `namba fix "<issue>"`
- Reviewable bugfix SPEC: `namba fix --command plan "<issue>"`
- Existing SPEC execution: `namba run SPEC-XXX`
- Usage or onboarding explanation: `$namba-help`
- Implementation finished and artifacts need refresh: `namba sync`
- Review handoff is ready: `namba pr "<Korean title>"`
- Approved PR is ready to merge: `namba land`

## Non-Goals

- Do not add a public `namba coach` Go CLI command in this slice.
- Do not create a separate SPEC artifact model, persistent coach output, readiness track, or review artifact type.
- Do not change the existing responsibilities of `$namba-help`, `$namba-create`, `namba plan`, `namba harness`, `namba fix`, or `namba run`.
- Do not let `$namba-coach` write `.namba/specs`, source files, skill files, custom-agent files, or review readiness files.

## Generated Output Expectations

After `namba regen` and `namba sync`, generated surfaces should consistently include `$namba-coach`:

- `AGENTS.md`
- `.agents/skills/namba/SKILL.md`
- `.agents/skills/namba-coach/SKILL.md`
- `.namba/codex/README.md`
- `README*.md`
- `docs/workflow-guide*.md`
- `.namba/manifest.json`

`namba regen` is the command that creates and preserves `.agents/skills/namba-coach/SKILL.md`. `namba sync` refreshes docs, project summaries, readiness, and PR-ready artifacts after implementation; it should not silently create missing managed skill files.

## Example Behavior

For the ambiguous request "todo 리스트를 만들고 싶은데 뭘 해야돼?", `$namba-coach` should not immediately create a SPEC. It should first ask only essential questions, such as target environment, UI surface, and persistence expectations. After the answers are concrete, it should hand off with an invocation like:

```text
namba plan "Build a todo list feature for <environment> with <UI surface> and <storage approach>."
```

If the user asks for `$namba-plan` but the intent is reusable skill, agent, workflow, or orchestration work, `$namba-coach` should recommend `namba harness "<description>"` and explain the routing correction.

If the user asks to directly create a repo-local skill or custom agent, `$namba-coach` should route to `$namba-create`. If the user asks to plan a reusable skill, agent, workflow, or orchestration change before implementation, it should route to `namba harness "<description>"` or `namba plan "<description>"` for Namba core managed-surface work.
