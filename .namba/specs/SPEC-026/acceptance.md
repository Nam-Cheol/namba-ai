# Acceptance

- [x] `$namba-create` remains the primary user-facing surface, but a real repo-tracked generator engine now exists behind the confirmed preview flow.
- [x] The implementation exposes a narrow callable adapter for the existing `$namba-create` skill wrapper without adding a new documented `namba create` Go CLI command.
- [x] The engine exposes a normalized preview model with target, slug or name, exact output paths, overwrite impact, validation plan, and session-refresh guidance.
- [x] No writes happen before confirmation.
- [x] A confirmed `skill` request writes only `.agents/skills/<slug>/SKILL.md`.
- [x] A confirmed `agent` request writes `.codex/agents/<slug>.toml` and `.codex/agents/<slug>.md` together.
- [x] A confirmed `both` request writes all expected files or none of them.
- [x] Invalid slugs, path traversal attempts, silent overwrites, and incomplete agent mirror writes are rejected.
- [x] If manifest persistence or paired-write completion fails, the engine rolls back the write set instead of leaving partial outputs behind.
- [x] Successful create outputs are tracked as user-authored so `namba regen` preserves them.
- [x] Generated docs and skill text describe `$namba-create` as a real generator entrypoint instead of a contract-only placeholder.
- [x] Regression coverage proves preview gating, exact path reporting, safe write behavior, rollback behavior, and regen preservation.
- [x] Validation commands pass.
