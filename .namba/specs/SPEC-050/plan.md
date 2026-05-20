# SPEC-050 Plan

1. Refresh implementation context.
   - Inspect the target files listed in `spec.md`, with code and generated
     templates treated as stronger evidence than docs.
   - Compare current generated Codex assets with Codex 0.131 release behavior:
     default-enabled plugin hooks, strict config parsing, service-tier and
     status-line display, permission and approval mode display, effective
     workspace roots, Windows sandbox hardening, and Git helper hook behavior.
2. Run plan review.
   - Run product, engineering, and design review artifacts under
     `.namba/specs/SPEC-050/reviews/`.
   - Refresh `reviews/readiness.md` before implementation starts.
3. Update hook config and runtime behavior.
   - Audit `.codex/hooks.json`, hook templates, and hook runtime event routing.
   - Ensure `SessionStart`, `UserPromptSubmit`, `PreToolUse`,
     `PermissionRequest`, `PostToolUse`, and `Stop` payloads all produce valid
     JSON.
   - Add session-id tolerant handling and safe PreToolUse `updatedInput`
     rewriting without weakening dangerous command blocks.
   - Enforce `PreToolUse` precedence: malformed payloads still return valid
     JSON, deny/block decisions beat rewrites, and rewrites apply only after an
     input is already allowed.
   - Prevent plugin-bundled hooks from causing duplicate Namba guard execution
     at the generated `.codex/hooks.json`, launcher, and shared guard boundary.
4. Update launcher behavior.
   - Harden `.codex/hooks/namba_codex_guard.sh` for missing Python, quoted
     paths, stdin JSON, and structured failure JSON.
   - Harden `.codex/hooks/namba_codex_guard.ps1` for missing Python, quoted
     paths, stdin JSON, structured failure JSON, no-profile launch behavior, and
     unsupported stop-parsing forms.
   - Keep `.codex/hooks/namba_codex_guard.py` as the shared policy authority.
5. Update config generation.
   - Audit `.codex/config.toml` generation through templates and YAML config
     sections.
   - Decide and implement the `namba codex access` ownership rule for
     `approval_policy` and `sandbox_mode`: preserve the current repo-safe
     baseline with tests/docs, or narrow/remove it consistently across CLI,
     templates, docs, and tests.
   - Remove legacy config or hook keys that Codex 0.131 no longer expects.
   - Keep output deterministic and parseable under strict TOML parsing.
   - Avoid user-specific model, auth, apps, web-search, sandbox, approval, and
     service-tier choices.
6. Update Windows and workspace behavior.
   - Align project analysis and Windows console helpers with effective
     workspace roots, scoped write roots, deny-read parity, ineffective
     firewall policy handling, and PowerShell profile avoidance.
   - Update `install.ps1` only where needed for Codex-facing PowerShell launch
     safety and profile avoidance.
7. Update documentation.
   - Update `AGENTS.md`, `README.md`, `README.ko.md`, and generated Codex/Namba
     docs to explain the Codex 0.131 compatibility boundary.
   - Document `.codex/hooks.json` versus `.namba/hooks.toml`, workspace root and
     permission behavior, Windows limitations, and Git helper hook limitations.
   - Use the reader order required by `spec.md`: Codex display, repo-managed
     Namba assets, Namba validation, and Namba limitations.
8. Add and update tests.
   - Extend `internal/namba/hook_runtime_test.go`,
     `internal/namba/hook_guard_test.go`, template tests, project-analysis
     tests, and Windows helper tests as needed.
   - Add fixtures for all six hook events and launcher failure paths.
   - Add deterministic config generation assertions.
9. Validate.
   - Run `go test ./...`.
   - Run `go vet ./...`.
   - Run `gofmt -l "cmd" "internal" "namba_test.go"` and format touched Go
     files if needed.
   - Regenerate or sync Namba artifacts with `namba sync` after implementation.
10. Final audit.
    - Confirm no dangerous-command blocking, approval risk notes,
      prompt-refinement behavior, final report guardrails, or repo-safe config
      boundaries were weakened.
    - Confirm no user-specific Codex runtime choices were added to repo-managed
      config.
    - Confirm a maintainer can distinguish Codex interactive guardrails from
      `namba run` evidence validation after reading the updated docs.
