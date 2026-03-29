# SPEC-013

## Goal

Update NambaAI's Codex upstream baseline so generated docs, config, and workflow guidance match the latest official Codex documentation instead of the older GitHub-only baseline.

## Context

- Project: namba-ai
- Project type: existing
- Language: go
- Mode: tdd
- Work type: plan
- Verified external context:
  - The official Code generation guide says Codex works best with the latest GPT-5 family such as `gpt-5.4`, and starting with `gpt-5.4` OpenAI recommends the general-purpose model for most code generation tasks.
  - The live Codex subagents docs say current Codex releases enable subagent workflows by default.
  - The current repo still uses `https://github.com/openai/codex` as the primary upstream reference in `docs/codex-upstream-reference.md`.
  - The current generated repo config is intentionally minimal and does not reflect the broader current `config.toml` surface now documented by OpenAI, including newer approval, permissions, tool, feature, and profile controls.
  - Current generated Namba text still says Codex does not expose a documented stop-hook surface, while the latest config reference now documents `features.codex_hooks`, so that claim must be re-audited.

## Problem

NambaAI's generated Codex-facing assets are drifting behind current official Codex docs. The project still points maintainers at a GitHub repo as the primary upstream baseline, preserves outdated wording around multi-agent and hook capabilities, and under-documents which parts of the modern Codex config surface Namba intentionally owns.

## Desired Outcome

- NambaAI treats the live Codex developer docs as the primary source of truth for Codex behavior, with the GitHub repo as a supplemental implementation reference.
- Generated guidance reflects current Codex semantics around `AGENTS.md`, repo skills in `.agents/skills`, project-scoped `.codex/config.toml`, and built-in subagent workflows.
- Namba no longer makes stale claims about experimental or undocumented Codex capabilities when the official docs have advanced.
- Repo-level Codex config generation either adopts the minimum current fields Namba should own or explicitly documents which advanced fields remain user-managed.
- Codex setup guidance documents the current Windows recommendation to prefer WSL for the best CLI experience.

## Scope

- Audit the current official Codex docs relevant to Namba scaffolding: CLI setup, subagents, customization/skills, sandboxing, and config reference.
- Update generated text in `AGENTS.md`, `.namba/codex/*`, workflow docs, and `docs/codex-upstream-reference.md`.
- Revisit `.codex/config.toml` generation and related docs so the generated baseline matches current Codex semantics without over-owning personal preferences.
- Update tests that assert generated Codex docs and config output.
- Refresh project docs and codemaps after the change.

## Non-Goals

- Do not redesign `namba run --solo|--team|--parallel` semantics that were already introduced for Codex orchestration.
- Do not implement a full Codex app server or Responses API integration in this SPEC.
- Do not attempt one-to-one hook parity unless the official Codex hook surface is sufficient, stable, and materially better than Namba's current explicit validator flow.

## Design Constraints

- Treat current `developers.openai.com/codex` pages as the product-level source of truth; use the GitHub repo only as a secondary implementation reference.
- Prefer a narrow, intentional Namba-owned config surface instead of copying the full upstream `config.toml` schema.
- Keep generated assets backward-compatible with the existing Namba workflow and branch/PR conventions.
