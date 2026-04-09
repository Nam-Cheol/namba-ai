# SPEC-026 Plan

1. Refresh project context and confirm the current phase-1 baseline from `SPEC-025`, the generated `$namba-create` skill surface, manifest ownership rules, and the absence of a real create engine in code.
2. Lock the follow-up boundary for this SPEC:
   - implement a Go generator engine plus the existing `$namba-create` wrapper
   - do not add a new documented `namba create` CLI
   - treat `SPEC-024` as closed historical context only
   - define the narrow internal adapter the skill wrapper will use to invoke the engine
3. Define internal request and preview models for:
   - target: `skill` / `agent` / `both`
   - normalized slug or name
   - exact output paths
   - overwrite policy and impact
   - validation plan and refresh guidance
4. Implement normalization and safety checks for invalid slugs, path traversal, overwrite confirmation, and durable write eligibility.
5. Implement repo-tracked writes for:
   - `.agents/skills/<slug>/SKILL.md`
   - `.codex/agents/<slug>.toml`
   - `.codex/agents/<slug>.md`
   while keeping agent writes atomic and `both` writes all-or-nothing.
6. Integrate manifest ownership updates so successful create outputs remain user-authored and survive `namba regen`, and roll back the write set when manifest persistence fails.
7. Update generated skill and doc surfaces so `$namba-create` reflects the real phase-2 generator behavior without expanding the documented CLI scope.
8. Add regression coverage for:
   - no writes before confirmation
   - preview exactness
   - explicit target override precedence
   - valid `skill` / `agent` / `both` writes
   - invalid slug and path traversal rejection
   - overwrite refusal
   - no partial agent mirror writes
   - rollback on manifest-update or paired-write failure
   - user-authored output preservation across `namba regen`
9. Regenerate managed skill and agent surfaces with `namba regen` when the templates or generated contract text change.
10. Run validation commands.
11. Refresh `.namba/specs/SPEC-026/reviews/*.md` and `readiness.md` if implementation decisions materially change review assumptions.
12. Sync artifacts with `namba sync`.
