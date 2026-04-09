# Acceptance

- [ ] `$namba-create` remains the primary user-facing surface, but a real repo-tracked generator engine now exists behind the confirmed preview flow.
- [ ] The implementation does not add a new `namba create` Go CLI command.
- [ ] The engine exposes a normalized preview model with target, slug or name, exact output paths, overwrite impact, validation plan, and session-refresh guidance.
- [ ] No writes happen before confirmation.
- [ ] A confirmed `skill` request writes only `.agents/skills/<slug>/SKILL.md`.
- [ ] A confirmed `agent` request writes `.codex/agents/<slug>.toml` and `.codex/agents/<slug>.md` together.
- [ ] A confirmed `both` request writes all expected files or none of them.
- [ ] Invalid slugs, path traversal attempts, silent overwrites, and incomplete agent mirror writes are rejected.
- [ ] Successful create outputs are tracked as user-authored so `namba regen` preserves them.
- [ ] Generated docs and skill text describe `$namba-create` as a real generator entrypoint instead of a contract-only placeholder.
- [ ] Regression coverage proves preview gating, exact path reporting, safe write behavior, and regen preservation.
- [ ] Validation commands pass.
