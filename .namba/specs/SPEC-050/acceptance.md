# Acceptance

- [ ] `.codex/hooks.json` remains the Codex interactive guardrail and does not
      duplicate Namba guard execution when plugin hooks are default-enabled.
- [ ] Duplicate-hook prevention is verified at the generated
      `.codex/hooks.json`, launcher, or shared guard level rather than inferred
      from `.namba/hooks.toml` or Namba runner hooks.
- [ ] `.namba/hooks.toml` remains documented as the `namba run` evidence and
      validation boundary.
- [ ] Generated `.codex/config.toml` is deterministic across repeated
      generation.
- [ ] Generated `.codex/config.toml` is valid under Codex 0.131 strict config
      parsing expectations.
- [ ] Repo-managed Codex config does not add user-specific model, auth, apps,
      web-search, sandbox, approval mode, or service-tier choices.
- [ ] The `namba codex access` contract for `approval_policy` and
      `sandbox_mode` is either preserved as an explicitly repo-safe baseline or
      narrowed/removed consistently across CLI behavior, templates, tests, and
      docs.
- [ ] Legacy hook or config keys that Codex 0.131 no longer expects are removed
      or no longer generated.
- [ ] `SessionStart` hook tests prove valid JSON output for normal payloads,
      session-id payloads, and missing optional metadata.
- [ ] `UserPromptSubmit` hook tests prove valid JSON output and preserved
      prompt-refinement behavior for empty, quoted, and multiline prompts.
- [ ] `PreToolUse` hook tests prove valid JSON output, preserved dangerous
      command blocking, and safe `updatedInput` rewrite behavior.
- [ ] `PreToolUse` hook tests prove deny/block decisions take precedence over
      `updatedInput` rewrites.
- [ ] `PermissionRequest` hook tests prove valid JSON output and preserved
      approval-risk guidance.
- [ ] `PostToolUse` hook tests prove valid JSON output for successful, failed,
      and missing-output tool payloads.
- [ ] `Stop` hook tests prove valid JSON output and preserved final report
      guardrails.
- [ ] POSIX launcher tests cover missing Python, quoted paths, stdin JSON, and
      structured failure JSON.
- [ ] POSIX launcher failure output preserves the originating hook event when
      possible and does not expose shell traces as the primary user-visible
      response.
- [ ] Windows launcher tests cover missing Python, quoted paths, stdin JSON,
      structured failure JSON, no-profile behavior, and unsupported
      stop-parsing avoidance.
- [ ] Windows launcher failure output preserves the originating hook event when
      possible and does not expose PowerShell traces as the primary user-visible
      response.
- [ ] Windows deny-read parity, scoped write roots, and ineffective firewall
      policy handling are reflected in code or documentation as appropriate.
- [ ] Effective workspace roots and permission or approval mode display behavior
      are documented.
- [ ] Git helper behavior that ignores configured hooks is documented as a
      limitation Namba cannot rely on for validation.
- [ ] Updated docs let a maintainer distinguish `.codex/hooks.json`
      interactive guardrails from `.namba/hooks.toml` validation/evidence
      policy and identify which Codex runtime settings remain user/session
      owned.
- [ ] Updated docs discuss Codex 0.131 boundaries in this order where practical:
      what Codex displays, what repo-managed Namba assets generate, what Namba
      validates, and what Namba cannot rely on or control.
- [ ] `AGENTS.md` explains the Codex 0.131 compatibility boundary clearly.
- [ ] `README.md` explains the Codex 0.131 compatibility boundary clearly.
- [ ] `README.ko.md` explains the Codex 0.131 compatibility boundary clearly.
- [ ] Generated Codex and Namba docs are refreshed where Namba owns the content.
- [ ] Dangerous command blocking is not weakened.
- [ ] Approval risk notes are not weakened.
- [ ] Prompt-refinement behavior is not weakened.
- [ ] Final report guardrails are not weakened.
- [ ] Repo-safe config boundaries are not weakened.
- [ ] `go test ./...` passes.
- [ ] `go vet ./...` passes.
- [ ] `gofmt -l "cmd" "internal" "namba_test.go"` reports no touched Go files
      needing formatting.
