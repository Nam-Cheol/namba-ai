---
name: namba-fix
description: Command-style entry point for direct bug repair or bugfix SPEC planning.
---

Use this skill when the user explicitly says `$namba-fix`, `namba fix`, or asks to repair a bug through Namba.

Behavior:
- Prefer the installed `namba fix` CLI when available.
- Treat `namba fix "<issue description>"` as the default direct-repair path in the current workspace.
- Use `namba fix --command run "<issue description>"` when the user wants the explicit direct-repair form.
- Use `namba fix --command plan "<issue description>"` when the user wants a reviewable bugfix SPEC package under `.namba/specs/`.
- Keep CLI help and flag probing read-only; `namba <command> --help`, `namba <command> -h`, and `namba help <command>` must not mutate repository state.
- Keep direct repairs small, add targeted regression coverage, run validation, and finish with `namba sync`.
