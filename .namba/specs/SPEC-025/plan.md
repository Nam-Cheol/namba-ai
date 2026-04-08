# SPEC-025 Plan

1. Refresh project context with `namba project` and confirm the skill/agent scaffold roots, regen behavior, MCP presets, and agent-thread limit from repo config and authoritative code.
2. Lock the product and architecture decisions in this SPEC:
   - phase 1 is `$namba-create` skill-first
   - generation is preview-first and confirmation-gated
   - raise the repo-managed same-workspace `max_threads` default from `3` to `5` for `agent_mode: multi`
   - "5 agents" means five independent role outputs with direct same-workspace support under repo defaults
   - user-authored outputs must be distinguished from Namba-managed built-ins
3. Extend the generated command-entry surface so `$namba-create` is discoverable in:
   - `internal/namba/templates.go`
   - `internal/namba/codex.go`
   - generated AGENTS/help/readme guidance that explains when to use create vs plan vs harness
4. Update generated repo-local Codex defaults so `.codex/config.toml` emits `[agents] max_threads = 5` for `agent_mode: multi`, while leaving worktree `max_parallel_workers` as a separate control surface.
5. Implement the clarification engine and routing contract:
   - unresolved -> narrowed -> confirmed state progression
   - remaining-unknown summary each turn
   - explicit `skill` / `agent` / `both` override precedence
   - preview/confirm summary before any writes
6. Implement safe artifact generation for skills and agents:
   - allowlisted paths
   - slug normalization
   - overwrite confirmation
   - `.toml` plus `.md` mirror generation for agents
   - rejection of stale Claude-only primitives or repo-policy violations in durable instructions
7. Implement ownership and regen-preservation behavior so user-authored `namba-create` outputs survive `namba regen` while built-in managed scaffolds remain fully regenerable from source.
8. Implement session-refresh signaling and validation hooks for instruction-surface mutations, including clear user-facing messaging when a fresh Codex session is required.
9. Add regression coverage for:
   - clarification-state progression
   - no writes before confirmation
   - explicit-user-intent override
   - skill/agent/both branch selection
   - generated `.codex/config.toml` default changes from `3` to `5` for `agent_mode: multi`
   - overwrite and path-safety handling
   - regen preservation for user-authored outputs
   - session-refresh behavior
   - five-role same-workspace planning and verification record requirements
10. Run validation commands.
11. Refresh `.namba/specs/SPEC-025/reviews/*.md` and `readiness.md` if implementation decisions materially change the current review conclusions.
12. Sync artifacts with `namba sync`.
