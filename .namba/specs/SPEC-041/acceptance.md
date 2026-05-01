# Acceptance

- [ ] NambaAI has an implementation-ready compatibility summary for Codex CLI `0.124.0`, `0.125.0`, `0.126.0`, `0.127.0`, and `0.128.0`, with official Codex docs or `openai/codex` release/tag evidence, preserved in `.namba/specs/SPEC-041/upstream-analysis.md` or a generated operator-facing reference.
- [ ] Capability tests cover `codex --version`, `codex exec --help`, and `codex exec resume --help` behavior for the target range, including additive CLI help; `codex exec --json` additive-field tolerance is tested only if implementation identifies a Namba-owned JSON consumer, otherwise it is documented as a compatibility note.
- [ ] NambaAI continues to prefer capability probing over fixed Codex version gating.
- [ ] Same-workspace Codex subagent capacity remains documented and validated as 5 threads, with explicit evidence that repo-managed `[agents] max_threads = 5` remains compatible with the target Codex config schema or local `0.128.0` config surface.
- [ ] Namba worktree parallelism remains explicitly separate from Codex subagent capacity; if it is raised from 3 to 5, workflow config, templates, docs, and parallel-run tests are updated together.
- [ ] Guidance distinguishes upstream `codex update` from NambaAI `namba update`.
- [ ] Guidance avoids recommending deprecated `--full-auto` and instead uses explicit sandbox, approval, and profile language.
- [ ] Repo-managed `.codex/config.toml` does not take over user-specific permission profiles, models, auth, apps, web search, or platform sandbox choices.
- [ ] Hook/plugin/MCP validation is scoped to Namba-owned integration boundaries: stable hook event names and observation payload shape, plugin-bundled hook enablement guidance where Namba documents it, MCP approval persistence assumptions, repo-managed MCP preset/cache/path boundaries, and custom MCP metadata where NambaAI touches them.
- [ ] Persisted `/goal` workflows are documented as future-facing opportunity, not required runtime behavior.
- [ ] Validation passes: `go test ./...`, `gofmt -l "cmd" "internal" "namba_test.go"`, and `go vet ./...`.
