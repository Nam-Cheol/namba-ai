---
name: namba-project
description: Command-style entry point for refreshing project docs and codemaps.
---

Use this skill when the user explicitly says `$namba-project`, `namba project`, or asks to analyze the current repository before implementation.

Behavior:
- Prefer the installed `namba project` CLI when available.
- Refresh `.namba/project/*` docs and codemaps before planning or execution.
- Treat `product.md` as the landing document, `tech.md` as the technical hub, and `structure.md` as appendix material.
- Surface system boundaries, evidence/confidence, mismatch reporting, and quality warnings instead of flattening the repository into a shallow tree dump.
- Summarize entry points, per-system artifacts, and any drift or thin-output warnings after the refresh.
