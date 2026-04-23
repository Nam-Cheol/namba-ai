# Acceptance

- [ ] `namba init` interactive flow includes a guided Codex access step that makes the resulting `approval_policy` and `sandbox_mode` understandable at selection time.
- [ ] `namba init` and `namba codex access` share one user-facing access model: the same preset labels, consequence statements, and mapping to `approval_policy` / `sandbox_mode`.
- [ ] `namba codex access` without mutation flags prints the current effective access state in a terminal-friendly format, including the resolved preset label and effective `approval_policy` / `sandbox_mode`, and does not apply changes.
- [ ] A project-root command, `namba codex access`, can update the repo-owned Codex access defaults in an already initialized Namba repository.
- [ ] Flag-driven usage is supported for the post-init edit path with explicit `--approval-policy` and `--sandbox-mode` inputs, and help text documents the available behavior.
- [ ] Access-setting changes persist through `.namba/config/sections/system.yaml` and regenerate `.codex/config.toml` consistently.
- [ ] If the effective access pair is unchanged, the command avoids unnecessary write/manifest churn and returns a clear no-change result.
- [ ] The post-init access edit path uses the managed-output pipeline and does not overwrite unrelated non-managed files such as project docs or README bundles.
- [ ] When the edit changes instruction-surface files, the command emits a session refresh notice consistent with the repo's existing Codex asset regeneration flow; no-op runs suppress that warning.
- [ ] `namba init --help`, `namba codex access --help`, and generated getting-started guidance are updated so the access-setting flow is discoverable both at bootstrap time and after initialization.
- [ ] Invalid or unsupported flag combinations fail clearly with remediation-oriented errors, and `namba codex access` fails clearly outside a Namba-managed repository.
- [ ] Tests cover parser/help behavior, shared access normalization, existing-repo reconfiguration, generated config outputs, and the no-clobber safety guarantee.
- [ ] Validation commands pass.
