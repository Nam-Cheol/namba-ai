---
name: namba-sync
description: Command-style entry point for refreshing Namba project artifacts after implementation.
---

Use this skill when the user explicitly says `$namba-sync`, `namba sync`, or asks to refresh PR-ready Namba artifacts after changes.

Behavior:
- Refresh `.namba/project/*` docs, release notes/checklists, codemaps, and any README bundles enabled by `.namba/config/sections/docs.yaml` after implementation.
- Use `namba regen` separately when template-generated scaffold assets changed.
- Run validation first when code changed and the quality config requires it.
