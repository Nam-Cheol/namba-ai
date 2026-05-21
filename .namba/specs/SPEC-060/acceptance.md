# Acceptance

- [x] `namba sync` regenerates `README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md`, `docs/getting-started*.md`, and `docs/workflow-guide*.md` from renderer/config output, not manual generated-file edits.
- [x] The root README first screen answers these user questions in order: what NambaAI is, why the generated docs are trustworthy, what to run first, which command fits the user's situation, and where to read next.
- [x] The root README contains a hero section using the existing configured hero image when available, with alt text and a text-first fallback when no hero image is configured.
- [x] The root README contains a badge row for release, CI, security, license, and documentation targets. Missing or inapplicable badge data degrades to safe text links or omission without broken links.
- [x] The root README contains a CTA row implemented with linked badges or plain GitHub Markdown links, not unsupported HTML buttons.
- [x] The root README contains a command-selection section covering `namba project`, `namba plan`, `namba harness`, `namba fix`, `namba queue`, `namba sync`, and `namba pr`, with each command tied to a user situation and next document.
- [x] Generated README and guide documents use card-style workflow or navigation summaries implemented as Markdown tables or short structured lists.
- [x] Generated README and guide documents include a compact step-by-step quick-start path.
- [x] Advanced or lengthy reference material is available through guide documents or `<details><summary>` sections, while essential new-user instructions remain visible outside collapsed content.
- [x] README, getting-started, workflow-guide, release, CI, and security references use consistent cross-document navigation grammar.
- [x] English, Korean, Japanese, and Chinese README and guide outputs keep structurally equivalent section order, navigation slots, CTA purposes, command-selection rows, and advanced-details positions.
- [x] Existing documentation coverage is preserved, including command skills, custom agents, run modes, queue flow, review readiness, release notes, CI, and security references.
- [x] Internal repository files and asset links use relative links wherever practical.
- [x] The renderer emits only GitHub-safe Markdown plus the minimal existing GitHub-compatible HTML needed for hero images and `<details><summary>` blocks. It does not emit custom CSS, JavaScript, iframe tags, style attributes, or unsupported HTML buttons.
- [x] Existing README and documentation contract tests are updated rather than removed.
- [x] Tests verify hero section, badge row, CTA row, command-selection section, quick-start section, card-style workflow/navigation section, collapsible advanced section, localized structural consistency, guide cross-document navigation, internal relative links, unsupported CSS/JavaScript/iframe absence, and deterministic generated output.
- [x] Focused README/sync contract tests pass.
- [x] `go test ./...` passes.
