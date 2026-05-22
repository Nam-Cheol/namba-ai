# SPEC-066 Contract

## Generated Surface Contract

Namba core generators must produce instruction surfaces for `.agents/skills/*`, `.codex/agents/*.md`, and `.codex/agents/*.toml` that state role or command purpose, read-only versus mutating boundaries, required output, pass/fail criteria, evidence expectations, security responsibilities, destructive command and escalation policy, fallback implementer boundaries, and non-project-specific portability.

## Source Of Truth

- Generator and renderer source under `internal/namba/templates.go`, `internal/namba/codex.go`, and `internal/namba/create_engine.go` owns the durable wording.
- Regenerated repo files and `.namba/manifest.json` are evidence, not manually patched source of truth.
- User-created skill and custom-agent outputs from `$namba-create` inherit the same safety and evidence contract without making them Namba-managed built-ins.

## Pass Criteria

- Fresh `init` output includes the contract in managed skills, role-card mirrors, and custom-agent TOML files.
- Existing `regen` output reproduces the same generated instruction contract without surface drift.
- Go tests prove the contract through deterministic renderer, init, regen, and create-flow assertions.
