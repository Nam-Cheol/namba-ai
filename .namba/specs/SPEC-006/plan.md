# SPEC-006 Plan

1. Add Codex-native scaffolding for `AGENTS.md`, `.agents/skills`, `.codex/skills`, and `.namba/codex` artifacts.
2. Add a `namba` repo skill that maps Namba commands to Codex-native workflow behavior.
3. Update `namba init` and `namba doctor` so initialized projects advertise Codex-native readiness.
4. Update README and generated docs to explain the `install -> namba init . -> open Codex` flow.
5. Validate with tests, refresh project docs, and sync artifacts.
6. Apply a Namba-oriented Codex status line in user config if supported by the local Codex installation.