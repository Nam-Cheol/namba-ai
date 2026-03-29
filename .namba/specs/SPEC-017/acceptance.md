# Acceptance

- [ ] The reported issue described below is resolved:
  Codex CLI capability mismatch causes team and parallel runs to fail late because exec and exec resume support different options
- [ ] `codex exec` requests can fall back to config overrides when the installed CLI does not expose a direct flag
- [ ] `codex exec resume` requests fail during preflight when a requested control such as `profile` cannot be represented safely
- [ ] `--parallel` fails before creating worker worktrees when the request contract cannot be represented for the local Codex CLI
- [ ] Validation commands pass
- [ ] Existing behavior around the affected area is preserved
- [ ] A regression test covering the fix is present
