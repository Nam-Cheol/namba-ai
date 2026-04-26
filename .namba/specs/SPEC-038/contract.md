# SPEC-038 Contract

## Surface

`$namba-coach` is a first-class, managed, generated repo skill under `.agents/skills/namba-coach/SKILL.md`.

It is exposed alongside existing Namba command-entry skills in:

- `AGENTS.md`
- `.agents/skills/namba/SKILL.md`
- `.namba/codex/README.md`
- `README*.md`
- `docs/workflow-guide*.md`

## Responsibilities

`$namba-coach` turns a user's current idea, vague request, or mistaken command choice into a concrete next Namba workflow.

It must:

- Restate the user's goal briefly.
- Follow this answer order: brief restatement, up to three essential clarification questions if required, one primary executable command, optional single alternative when there is a real tradeoff, and a short reason.
- Treat essential clarification as information required to choose the correct workflow or make the command handoff usable, not information that fully specifies implementation.
- Ask only 1-3 essential clarification questions when the request is underspecified.
- Recommend exactly one primary executable invocation once the request is concrete enough.
- Offer at most one alternative invocation when there is a meaningful tradeoff.
- Explain the recommendation in one or two sentences.
- Correct clearly wrong command choices first instead of executing them unchanged.

## Read-Only Boundary

`$namba-coach` must not:

- Create SPEC packages.
- Edit repository files.
- Generate skill, agent, source, or review artifacts.
- Run implementation.
- Update `.namba/specs/<SPEC>/reviews/readiness.md`.
- Add a public `namba coach` CLI command.

## Boundary With Existing Skills

- `$namba-help` explains NambaAI usage, command semantics, and documentation locations.
- `$namba-coach` selects the next workflow handoff for the current user goal.
- `$namba-create` creates repo-local skill or custom-agent artifacts directly after preview and confirmation.
- `namba plan` creates feature-oriented SPEC packages.
- `namba harness` creates reusable skill, agent, workflow, or orchestration SPEC packages when they do not require direct Namba core changes.
- Namba core changes, including managed generated skill registry changes, remain explicit in SPEC planning and validation.

## Routing Table

- Feature or product change: `namba plan "<description>"`
- Reusable skill, agent, workflow, or orchestration SPEC: `namba harness "<description>"`
- Direct repo-local skill or custom-agent artifact: `$namba-create`
- Direct bug repair: `namba fix "<issue>"`
- Reviewable bugfix SPEC: `namba fix --command plan "<issue>"`
- Existing SPEC execution: `namba run SPEC-XXX`
- Usage or onboarding explanation: `$namba-help`
- Implementation finished and generated artifacts need refresh: `namba sync`
- Review handoff is ready: `namba pr "<Korean title>"`
- Approved PR is ready to merge: `namba land`

## Wrong-Command Examples

- If a user invokes `$namba-plan` for reusable skill, agent, workflow, or orchestration planning, recommend `namba harness "<description>"` unless the requested change touches Namba core managed surfaces, in which case recommend `namba plan "<description>"`.
- If a user asks to directly create a repo-local skill or custom agent artifact, recommend `$namba-create` instead of `namba harness`.
- If a user asks how Namba commands work or where docs live, recommend `$namba-help` instead of turning the answer into a planning handoff.

## Acceptance Example

For "todo 리스트를 만들고 싶은데 뭘 해야돼?", coach should ask essential questions first:

- Which environment should this target?
- What UI surface should users interact with?
- Should tasks be local-only or persisted somewhere?

After those answers, coach should hand off with:

```text
namba plan "Build a todo list feature for <environment> with <UI surface> and <storage approach>."
```
