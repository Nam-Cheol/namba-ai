# Product

Source: README.md

# NambaAI

NambaAI is a Codex-first agentic development kit focused on SPEC-driven execution,
explicit quality gates, and worktree-based parallelism.

## Commands

- `namba init [path]`
- `namba doctor`
- `namba status`
- `namba project`
- `namba plan "<description>"`
- `namba run SPEC-XXX [--parallel] [--dry-run]`
- `namba sync`
- `namba worktree <new|list|remove|clean>`

## Runtime model

NambaAI does not emulate Claude hooks. It uses explicit phase validators and
Codex execution orchestration instead:

1. `project` generates project docs and codemaps
2. `plan` creates a SPEC package under `.namba/specs/`
3. `run` builds a Codex execution request from the SPEC and optionally fans out
   into git worktrees
4. `sync` refreshes docs and emits PR-ready artifacts
