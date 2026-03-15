# Data Flow

1. `init` creates AGENTS, skills, and `.namba`
2. `project` refreshes docs and codemaps
3. `plan` creates a SPEC package
4. `run` builds a Codex execution request and validates the result
5. `sync` emits PR-ready artifacts
