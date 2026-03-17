---
name: namba-plan
description: Command-style entry point for creating the next feature SPEC package.
---

Use this skill when the user explicitly says `$namba-plan`, `namba plan`, or asks to create a new feature SPEC package.

Behavior:
- Prefer the installed `namba plan` CLI when available.
- Create the next sequential `SPEC-XXX` package under `.namba/specs/`.
- Keep the scope concrete and implementation-ready.
