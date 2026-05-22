---
name: namba-project
description: Command-style entry point for refreshing project docs and codemaps.
---

State effect: mutating workflow entry point. Use help/probe paths read-only, and otherwise expect repository state or GitHub state to change.

Generated instruction contract for this command skill:
- Purpose: keep the role or command scope explicit, bounded, and testable.
- Boundary: honor read-only versus mutating state effects, configured sandbox mode, and assigned file or workflow ownership.
- Required output: report concrete actions, changed paths or artifacts, validation evidence, and pass/fail status or blockers.
- Pass/fail criteria: claim success only when acceptance criteria and configured validation are satisfied; otherwise name the exact blocker and impact.
- Evidence expectations: cite source artifacts such as SPEC files, `.namba/` configs, diffs, test output, PR/check links, or generated manifests instead of relying on unsupported assertions.
- Security responsibilities: never expose or commit secrets; treat auth, privacy, destructive commands, permission changes, and external network or credential use as security-sensitive.
- Destructive command and escalation policy: do not run destructive commands unless explicitly requested, and request approval for privileged, networked, or sandbox-blocked actions.
- Fallback implementer boundary: if a specialist path is unavailable and the main/default implementer takes over, stay within the assigned scope and preserve the same evidence and validation duties.
- Portability: keep durable guidance non-project-specific unless the current repository config or SPEC explicitly provides the project detail.

Use this skill when the user explicitly says `$namba-project`, `namba project`, or asks to analyze the current repository before implementation.

Behavior:
- Prefer the installed `namba project` CLI when available.
- Refresh `.namba/project/*` docs and codemaps before planning or execution.
- Treat `product.md` as the landing document, `tech.md` as the technical hub, and `structure.md` as appendix material.
- Surface system boundaries, evidence/confidence, mismatch reporting, and quality warnings instead of flattening the repository into a shallow tree dump.
- Summarize entry points, per-system artifacts, and any drift or thin-output warnings after the refresh.
