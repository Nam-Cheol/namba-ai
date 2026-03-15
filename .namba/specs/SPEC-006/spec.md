# SPEC-006

## Goal

Convert Namba from a Codex-aware wrapper into a Codex-native repository workflow surface.

## Scope

- Make `namba init .` create Codex-native repository assets.
- Add repo-local skills under `.agents/skills/` and keep `.codex/skills/` as a compatibility mirror.
- Add a repo skill named `namba` that works as the in-Codex Namba command surface.
- Update `AGENTS.md` so `namba run SPEC-XXX` is interpreted as Codex-native in-session execution.
- Keep the standalone `namba run` CLI for non-interactive runner execution.
- Add project docs that explain how Codex uses Namba after init.
- Align README with the Codex-native model and document status line customization.

## Non-Goals

- Building an officially supported custom `/namba` slash command if Codex does not expose one.
- Removing the existing non-interactive `namba run` runner path.
- Replacing global Codex configuration with per-project config when the feature is not officially supported.

## Constraints

- Follow Codex official primitives that are currently documented: `AGENTS.md`, repo skills, and user config.
- Preserve compatibility with environments that still read `.codex/skills/`.
- `namba init .` should be enough to make a project Codex-ready.