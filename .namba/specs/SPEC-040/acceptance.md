# Acceptance

- [x] The SPEC includes an explicit audit baseline for the useful `ComposioHQ/awesome-codex-skills` patterns and rejected upstream flows.
- [x] No new standalone Namba skill is added; all planned changes adapt existing Namba-managed skills, templates, docs, or helper candidates.
- [x] The plan excludes third-party skill installation, Composio CLI/OAuth dependencies, Slack/Notion/app automation, and any `$CODEX_HOME/skills` install flow.
- [x] A harness adaptation map identifies which upstream patterns influence `$namba-review-resolve`, `namba pr`, `namba harness`, frontend validation, MCP/evaluation criteria, and batch migration guidance.
- [x] Skill trigger metadata criteria cover explicit command triggers, natural-language equivalents where useful, and read/write boundaries.
- [x] Progressive-disclosure criteria keep command-entry `SKILL.md` files lean and require references or helper candidates for long procedural recipes.
- [x] `$namba-review-resolve` guidance requires thread-aware review state, per-thread outcomes, original-thread replies, validation evidence, CI/check evidence when relevant, and resolution only after the item is addressed.
- [x] `namba pr` guidance requires `namba sync`, configured validation, PR check inspection, bounded GitHub Actions failure snippets when checks fail, external-check URL reporting, and exactly one configured Codex review marker.
- [x] Deterministic helper-script candidates are evaluated against clear criteria: `--help`, fixture/local testability, bounded output, no destructive actions, explicit network assumptions, and no third-party app/OAuth coupling.
- [x] Frontend validation guidance covers server lifecycle, rendered-state inspection, screenshots, console-log evidence, and Playwright-style checks without creating a separate frontend workflow.
- [x] Harness/MCP quality criteria cover workflow-first design, context-budgeted outputs, actionable errors, and independent/read-only/realistic/verifiable/stable evaluation scenarios.
- [x] Batch-migration guidance requires inventory, mechanical-versus-behavioral classification, template-first edits, generated diff review, and validation.
- [x] Implementation updates source templates before generated skill outputs and uses `namba regen` for managed skill changes.
- [x] Regression coverage locks the new trigger, PR/review evidence, helper-candidate, and harness/evaluation guidance where wording or behavior is contractual.
- [x] Validation commands pass: `gofmt -l "cmd" "internal" "namba_test.go"`, `GOCACHE=/tmp/namba-go-cache go test ./...`, `GOCACHE=/tmp/namba-go-cache go vet ./...`, and `git diff --check`.
