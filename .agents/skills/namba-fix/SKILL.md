---
name: namba-fix
description: Command-style entry point for creating the next bug-fix SPEC package.
---

Use this skill when the user explicitly says `$namba-fix`, `namba fix`, or asks to prepare a bug-fix SPEC package.

Behavior:
- Prefer the installed `namba fix` CLI when available.
- Create the next sequential `SPEC-XXX` fix package under `.namba/specs/`.
- Bias toward the smallest safe fix and explicit regression coverage.
