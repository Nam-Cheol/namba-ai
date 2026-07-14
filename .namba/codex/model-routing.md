# Model Routing

Policy: `gpt-5.6-cost-balanced-v1`.

- `efficient` → `gpt-5.6-luna`
- `standard` → `gpt-5.6-terra`
- `deep` → `gpt-5.6-sol`

Namba does not provide per-turn model or reasoning overrides. `requested` is the deterministic policy choice; `effective` is recorded only when Codex reports it; `external_unobserved` means that observation was unavailable and was not inferred.

`namba run --dry-run` reports planned turns. A non-required Sol decision can fall back to Terra high with evidence. Required Sol failure is `blocked_model_unavailable`; install or authorize the exact model and rerun the command after the capability probe succeeds.

Sol is read-only for design, architecture, or high-risk checkpoints. Terra or Luna receives its checkpoint and performs implementation.
