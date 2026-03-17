---
name: namba-update
description: Command-style entry point for self-updating the installed NambaAI CLI.
---

Use this skill when the user explicitly says `$namba-update`, `namba update`, or asks to update the installed NambaAI version.

Behavior:
- Treat `namba update` as CLI self-update from GitHub Release assets.
- Use `namba update --version vX.Y.Z` when the user requests a specific version.
- Do not confuse this command with scaffold regeneration; that belongs to `namba regen`.
