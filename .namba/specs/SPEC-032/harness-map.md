# SPEC-032 Harness Map

## Purpose

Show the structural boundary between core harness work, domain harness work, and direct artifact creation.

## Layer Map

| Layer | Owns what | Typical surfaces |
| --- | --- | --- |
| Core harness | Namba-owned routing, planning, review, validator, runtime, and built-in skill or agent contracts | `internal/namba/*`, built-in `namba-*` skills, built-in Namba custom agents, generated workflow docs |
| Domain harness | Reusable user or domain-specific harness contracts built on top of the core harness | future domain-specific SPECs, user-authored harness docs, user-authored reusable skill or agent systems |
| Direct artifacts | Concrete generated outputs that realize an already-resolved request | `.agents/skills/<slug>/SKILL.md`, `.codex/agents/<slug>.toml`, `.codex/agents/<slug>.md` |

## Adaptation Rules

- Domain harness work may depend on a core harness contract, but it must name that dependency through `base_contract_ref` instead of implying it in prose.
- If a domain harness change needs to edit Namba-owned core surfaces, it is no longer only a domain harness change and must escalate to `core_harness_change`.
- Direct artifact creation may implement a resolved domain harness decision later, but it must not replace the planning or review step when the contract itself is still moving.
- `harness-map.md` is mandatory only when:
  - a domain harness adapts or composes an existing harness
  - or the route boundary with direct artifact creation needs explicit documentation

It is not mandatory for every brand-new domain harness request in v1.

## Split Guidance

- Use one SPEC when the main task is defining or changing the contract.
- Use `$namba-create` after the contract is stable and the target artifact is clear.
- Split the work when one request mixes:
  - core harness contract change
  - domain harness addition
  - direct artifact generation

The split keeps review, validation, and ownership boundaries clear.
