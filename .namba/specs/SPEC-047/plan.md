# SPEC-047 Plan

1. Inspect generated-surface ownership for `.codex/hooks/namba_codex_guard.py`, `.codex/hooks/namba_codex_guard.sh`, `.codex/hooks/namba_codex_guard.ps1`, and `.codex/hooks.json`.
2. Locate and update any source templates before editing generated hook outputs directly.
3. Add black-box Python `unittest` regression tests for `.codex/hooks/namba_codex_guard.py`.
4. Pin current Codex-facing output schemas in tests for prompt context, shell denials, permission denials, non-deny risk notes, managed-surface reminders, trace lines, and Stop blocking.
5. Cover `UserPromptSubmit`, `PreToolUse`, `PermissionRequest`, `PostToolUse`, trace output via `NAMBA_HOOK_TRACE_PATH`, and minimal `Stop` behavior if needed.
6. Fix inverted runtime branch logic for prompt refinement, dangerous shell command denial, and permission request handling.
7. Regenerate hook outputs if templates drive the `.codex/hooks/` surface.
8. Add an explicit Python `unittest` hook-test step to CI.
9. Update documentation only where existing hook behavior descriptions become inaccurate.
10. Run `namba sync` after implementation changes.
11. Rerun the validation plan from `acceptance.md` after `namba sync`.
