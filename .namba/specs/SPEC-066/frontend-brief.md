# Generated Instruction Contract Brief

Task Classification: frontend-minor
Classification Rationale: Validator-readable non-visual classification: this SPEC does not implement a browser, app, or visual product surface. It hardens Namba core managed generator/source-of-truth behavior for init/regen generated instruction surfaces.
Frontend Gate Status: not-applicable
Problem Gate: not-applicable
Reference Gate: not-applicable
Critique Gate: not-applicable
Decision Gate: not-applicable
Prototype Gate: not-applicable
Prototype Evidence: n/a
Frontend implementation phase: not-applicable
Asset mode: not-applicable
Imagegen requirement: not-applicable
Asset decision proof: Generated bitmap assets are not relevant because this SPEC changes CLI-managed text templates and agent/skill contracts.

## Generated Contract Brief

- Primary surfaces: `.agents/skills/*`, `.codex/agents/*.md`, and `.codex/agents/*.toml`.
- Source-of-truth paths: `internal/namba/codex.go`, `internal/namba/templates.go`, `internal/namba/create_engine.go`, `internal/namba/agent_runtime.go`, and related `internal/namba/*_test.go`.
- Managed boundary: `.namba/manifest.json` ownership must remain correct.
- Contract risk: generated guidance can become generic, project-specific, unclear about authority boundaries, or difficult to validate.
- Direction: concise role purpose, explicit read-only versus mutating boundaries, required outputs, pass/fail criteria, evidence expectations, security responsibilities, fallback implementer boundaries, destructive command/escalation policy, and implementation-ready validation notes.
- Accessibility and resilience: plain Markdown, scannable headings, deterministic language, and no dependency on images, colors, external services, or LLM-judge evaluation.
- Evidence target: Go golden/snapshot/unit tests and temporary-repo CLI integration tests prove generated output quality, init/regen parity, and idempotence.

## Review Interpretation

- Product review: future `namba init` users receive clearer skill/agent guidance; non-goals and downstream backfill exclusions are explicit.
- Engineering review: generator source changes, init/regen parity, manifest ownership, golden/snapshot/unit/integration tests, and validation commands are sufficient.
- Design review: generated instruction contract clarity, role boundary clarity, and output-contract readability only.
- Security advisory: read-only versus mutating boundary, secrets prohibition, destructive command/escalation policy, and non-project-specific guidance.
