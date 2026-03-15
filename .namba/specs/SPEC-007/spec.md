# SPEC-007

## Goal

Port the MoAI-style `init` experience into NambaAI with a Codex-native scaffold.

## Context

- `moai init .` provides a guided setup flow with explicit choices such as project name and TDD/DDD.
- Claude Code and Codex do not expose the same project-local primitives.
- NambaAI must preserve the workflow intent while generating assets that Codex can actually use.

## Requirements

- `namba init .` should support a guided wizard in interactive terminals.
- Non-interactive environments must still work with `--yes` and explicit flags.
- Generated assets must include repo-local Codex skills, Codex role cards, AGENTS, and `.namba` config.
- Secrets such as PATs must not be written to generated config.
- The generated scaffold must explain how Claude/MoAI concepts map onto Codex.
