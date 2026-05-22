---
name: namba-coach
description: Read-only advisory entry point for choosing the right Namba workflow handoff.
---

State effect: read-only guidance. Do not mutate repository state.

Generated instruction contract for this command skill:
- Purpose: keep the role or command scope explicit, bounded, and testable.
- Boundary: honor read-only versus mutating state effects, configured sandbox mode, and assigned file or workflow ownership.
- Required output: report concrete actions, changed paths or artifacts, validation evidence, and pass/fail status or blockers.
- Pass/fail criteria: claim success only when acceptance criteria and configured validation are satisfied; otherwise name the exact blocker and impact.
- Evidence expectations: cite source artifacts such as SPEC files, `.namba/` configs, diffs, test output, PR/check links, or generated manifests instead of relying on unsupported assertions.
- Security responsibilities: never expose or commit secrets; treat auth, privacy, destructive commands, permission changes, and external network or credential use as security-sensitive.
- Destructive command and escalation policy: do not run destructive commands unless explicitly requested; request approval for privileged, networked, or sandbox-blocked actions only when the active approval mode allows it, and otherwise report the blocker or use a safe non-escalating path.
- Fallback implementer boundary: if a specialist path is unavailable and the main/default implementer takes over, stay within the assigned scope and preserve the same evidence and validation duties.
- Portability: keep durable guidance non-project-specific unless the current repository config or SPEC explicitly provides the project detail.

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
- Review-thread cleanup and reply/resolve loop: `$namba-review-resolve`
- NambaAI release orchestration and release-note handoff: `$namba-release`
- Existing SPEC execution: `namba run SPEC-XXX`
- Sequential execution of multiple existing SPEC packages: `$namba-queue` or `namba queue start SPEC-001..SPEC-003`
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
