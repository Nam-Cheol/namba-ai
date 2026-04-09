# SPEC-026 Plan

1. Refresh project context and confirm the current phase-1 baseline from `SPEC-025`, the generated `$namba-create` skill surface, manifest ownership rules, and the absence of a real create engine in code.
2. Lock the follow-up boundary for this SPEC:
   - implement a Go generator engine plus the existing `$namba-create` wrapper
   - do not add a new `namba create` CLI
   - treat `SPEC-024` as closed historical context only
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
6. Integrate manifest ownership updates so successful create outputs remain user-authored and survive `namba regen`.
7. Update generated skill and doc surfaces so `$namba-create` reflects the real phase-2 generator behavior without expanding CLI scope.
8. Add regression coverage for:
   - no writes before confirmation
   - preview exactness
   - explicit target override precedence
   - valid `skill` / `agent` / `both` writes
   - invalid slug and path traversal rejection
   - overwrite refusal
   - no partial agent mirror writes
   - user-authored output preservation across `namba regen`
9. Run validation commands.
10. Refresh `.namba/specs/SPEC-026/reviews/*.md` and `readiness.md` if implementation decisions materially change review assumptions.
11. Sync artifacts with `namba sync`.
