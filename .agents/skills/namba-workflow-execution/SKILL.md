---
name: namba-workflow-execution
description: Execute NambaAI SPEC packages with Codex-native workflow and explicit validation.
---

Use this skill when implementing a SPEC package.

State effect: mutating SPEC execution workflow. Read-only only while inspecting the SPEC package before implementation.

Generated instruction contract for this repo skill:
- Purpose: keep the role or command scope explicit, bounded, and testable.
- Boundary: honor read-only versus mutating state effects, configured sandbox mode, and assigned file or workflow ownership.
- Required output: report concrete actions, changed paths or artifacts, validation evidence, and pass/fail status or blockers.
- Pass/fail criteria: claim success only when acceptance criteria and configured validation are satisfied; otherwise name the exact blocker and impact.
- Evidence expectations: cite source artifacts such as SPEC files, `.namba/` configs, diffs, test output, PR/check links, or generated manifests instead of relying on unsupported assertions.
- Security responsibilities: never expose or commit secrets; treat auth, privacy, destructive commands, permission changes, and external network or credential use as security-sensitive.
- Destructive command and escalation policy: do not run destructive commands unless explicitly requested; request approval for privileged, networked, or sandbox-blocked actions only when the active approval mode allows it, and otherwise report the blocker or use a safe non-escalating path.
- Fallback implementer boundary: if a specialist path is unavailable and the main/default implementer takes over, stay within the assigned scope and preserve the same evidence and validation duties.
- Portability: keep durable guidance non-project-specific unless the current repository config or SPEC explicitly provides the project detail.

Execution pattern:
1. Read `.namba/specs/<SPEC>/spec.md`
2. Read `.namba/specs/<SPEC>/plan.md`
3. Read `.namba/specs/<SPEC>/acceptance.md`
4. Read `.namba/specs/<SPEC>/reviews/readiness.md` when present so advisory review status informs execution
5. Read `.namba/specs/<SPEC>/frontend-brief.md` when present and treat it as the canonical frontend gate contract
6. If the SPEC is explicitly `frontend-major`, stop and route back to research/synthesis when the frontend gate is incomplete, invalid, or contradicted by design-review summaries
7. Implement the work directly in the current Codex session
8. For browser-rendered frontend changes, use managed server lifecycle, inspect rendered DOM state, capture screenshots, record console errors, and prefer Playwright checks when practical
9. Run configured validation commands
10. Summarize results in `.namba/logs` and sync artifacts

Collaboration defaults: use a dedicated branch from `main` for the SPEC, open the PR into `main`, write the PR in Korean, and request Codex review only when `namba pr --review` or queue `--review` is explicit; use `@codex review` as the request command.

Do not call `namba run` from inside Codex unless the user explicitly requests the non-interactive CLI runner.
