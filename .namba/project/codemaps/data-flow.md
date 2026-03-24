# Data Flow

1. `init` runs a Codex-adapted project wizard, writes `.namba/config/sections/*.yaml`, repo skills under `.agents/skills`, command-entry skills such as `$namba-run`, project-scoped custom agents under `.codex/agents/*.toml`, readable `.md` agent mirrors, and Codex repo config under `.codex/config.toml`
2. `project` refreshes docs and codemaps
3. `plan` creates a SPEC package
4. `run` supports the default standalone flow, explicit `--solo` and `--team` subagent-oriented requests, and worktree-based `--parallel` execution
5. `sync` emits PR-ready artifacts
