---
name: namba-project
description: Command-style entry point for refreshing project docs and codemaps.
---

Use this skill when the user explicitly says `$namba-project`, `namba project`, or asks to analyze the current repository before implementation.

Behavior:
- Prefer the installed `namba project` CLI when available.
- Refresh `.namba/project/*` docs and codemaps before planning or execution.
- Summarize entry points, structure, and generated artifacts after the refresh.
