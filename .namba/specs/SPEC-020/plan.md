# SPEC-020 Plan

1. Inspect the current planning-command parsing path, scaffold generator, and generated guidance surfaces to identify where a dedicated `namba harness` surface fits without changing default `namba plan` behavior.
2. Add explicit `namba harness "<description>"` parsing with safe validation and clear help text, reusing the read-only help behavior that `SPEC-019` establishes for `plan` and related planning commands.
3. Define harness-template scaffold content for `spec.md`, `plan.md`, and `acceptance.md` so the generated SPEC captures execution topology, agent/skill boundaries, progressive-disclosure layout, trigger guidance, and evaluation strategy.
4. Add shared reference content or template helpers that explain how portable Harness concepts map onto Codex-native Namba primitives such as repo skills, custom agents, explicit subagents, and worktree fan-out/fan-in.
5. Keep the default `namba plan "<description>"` flow unchanged for ordinary feature work.
6. Update `.agents/skills/namba-plan/SKILL.md`, `.agents/skills/namba/SKILL.md`, `internal/namba/templates.go`, `internal/namba/readme.go`, and generated README/workflow docs so users can discover when to use `namba plan` versus `namba harness`.
7. Add regression tests for the new command parsing, harness scaffold generation, no-Claude-primitive output guarantees, and generated documentation/help text.
8. Run validation commands and regenerate any derived docs or instruction surfaces required to keep the repository contract aligned.
