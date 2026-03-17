# Data Flow

1. `init` runs a Codex-adapted project wizard, writes `.namba/config/sections/*.yaml`, repo skills under `.agents/skills`, Codex custom agents under `.codex/agents/*.toml`, readable `.md` agent mirrors, a compatibility mirror under `.codex/skills`, and Codex repo config under `.codex/config.toml`
2. `project` refreshes docs and codemaps
3. `plan` creates a SPEC package
4. `run` either builds a non-interactive Codex execution request or is interpreted as Codex-native in-session execution
5. `sync` emits PR-ready artifacts
