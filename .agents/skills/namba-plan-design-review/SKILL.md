---
name: namba-plan-design-review
description: Command-style entry point for design review of a SPEC before implementation starts.
---

State effect: mutating review update. Writes SPEC review artifacts and readiness summaries.

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

Use this skill when the user explicitly says `$namba-plan-design-review` or asks for a design review on a SPEC package.

Behavior:
- Resolve the target SPEC from an explicit `SPEC-XXX`; otherwise use the latest SPEC under `.namba/specs/`.
- Read `.namba/specs/<SPEC>/spec.md`, `plan.md`, and `acceptance.md` before writing review notes.
- Update `.namba/specs/<SPEC>/reviews/design.md` with status, reviewer, findings, decisions, follow-ups, and recommendation.
- Prefer `namba-designer` as the review role when subagent routing is appropriate.
- Check art-direction clarity, palette temperature and undertone logic, restrained saturation, semantic component choice, anti-generic composition, meaningful motion, redesign of the most generic section when applicable, and anti-overcorrection safeguards.
- Treat card, border, and grid usage as valid when they are justified; the regression target is default reliance or meaningless fallback use, not a categorical ban.
- Reject both washed-out gray minimalism and novelty-for-novelty. Preserve accessibility, design-system fit, and implementation realism.
- Refresh `.namba/specs/<SPEC>/reviews/readiness.md` so the advisory summary reflects the latest review state.
- Keep the review advisory by default; surface missing depth or blockers clearly without silently turning the workflow into a hard gate.
