---
name: namba-coach
description: Read-only advisory entry point for choosing the right Namba workflow handoff.
---

Use this skill when the user explicitly says `$namba-coach`, arrives with a vague current goal, asks what to do next, or appears to have selected the wrong Namba command.

Behavior:
- Stay read-only. Do not create SPEC packages, edit repository files, generate skill, agent, source, or review artifacts, run implementation, update `.namba/specs/<SPEC>/reviews/readiness.md`, or invoke a public `namba coach` CLI command.
- Restate the user's current goal briefly before choosing a path.
- Follow this response order: brief restatement, up to three essential clarification questions when required, one primary executable handoff, optional single alternative when there is a real tradeoff, and a one- or two-sentence reason.
- Treat essential clarification as information needed to choose the correct Namba workflow or make the handoff command usable, not information that would fully specify implementation.
- Ask only 1-3 essential clarification questions when the request is underspecified.
- Once the request is concrete enough, recommend exactly one primary executable invocation and at most one alternative.
- Correct a clearly wrong command choice first instead of running it as-is.

Boundary with `$namba-help`:
- `$namba-help` explains how NambaAI works, what commands mean, and where docs live.
- `$namba-coach` uses the user's current idea or question to choose the next workflow handoff.

Routing rules:
- New feature or product change: `namba plan "<description>"`
- Reusable skill, agent, workflow, or orchestration SPEC: `namba harness "<description>"`
- Direct repo-local skill or custom-agent creation: `$namba-create`
- Bug repair: `namba fix "<issue>"`
- Reviewable bugfix SPEC: `namba fix --command plan "<issue>"`
- Existing SPEC execution: `namba run SPEC-XXX`
- Usage or onboarding explanation: `$namba-help`
- Implementation finished and artifacts need refresh: `namba sync`
- Review handoff is ready: `namba pr "<Korean title>"`
- Approved PR is ready to merge: `namba land`

Wrong-command correction:
- If the user asks for `$namba-plan` but the intent is reusable skill, agent, workflow, or orchestration planning, recommend `namba harness "<description>"` unless the request touches Namba core managed surfaces, where `namba plan "<description>"` remains appropriate.
- If the user asks to directly create a repo-local skill or custom agent artifact, recommend `$namba-create` instead of `namba harness`.
- If the user asks how Namba commands work or where docs live, recommend `$namba-help` instead of turning the answer into a planning handoff.

Example:
- For `todo 리스트를 만들고 싶은데 뭘 해야돼?`, ask only essential questions first: target environment, UI surface, and whether tasks should be local-only or persisted somewhere.
- After those answers are concrete, hand off with `namba plan "Build a todo list feature for <environment> with <UI surface> and <storage approach>."`
