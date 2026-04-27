# SPEC-040 Contract

## Invariants

- Evolve existing Namba-managed skills only; do not add new standalone skills.
- Treat `ComposioHQ/awesome-codex-skills` as an audit source for patterns, not as an import source.
- Keep Namba policy, local code, generated templates, and `.namba/` artifacts authoritative.
- Change generated skill surfaces through source templates plus `namba regen`.
- Keep command-entry skills compact and load detailed recipes only through progressive disclosure or deterministic helper candidates.
- Do not add third-party skill installation, Composio CLI/OAuth, Slack/Notion/app automation, or `$CODEX_HOME/skills` flows.

## Adopted Pattern Set

| Upstream pattern | Adapted Namba rule |
| --- | --- |
| Precise skill `description` metadata | Every affected Namba skill must state explicit command triggers and real task triggers without broad catch-all wording. |
| Progressive disclosure | Long review, CI, frontend, or MCP recipes should move into existing-skill references or Namba-owned helper candidates. |
| Deterministic scripts | Repetitive review-thread, CI-log, and frontend-validation steps should be candidates for small helper scripts with `--help` and fixture tests. |
| Thread-aware review handling | `$namba-review-resolve` must use review-thread state for resolution decisions, not flat comments alone. |
| CI log inspection | `namba pr` and review-resolution workflows should inspect failing GitHub Actions logs and summarize bounded snippets before handoff. |
| Webapp testing helper discipline | Frontend validation should manage servers, wait for rendered state, capture screenshots, inspect DOM facts, and report console errors. |
| MCP evaluation quality | Harness/MCP work should define workflow-first tools, concise outputs, actionable errors, and stable read-only evaluation scenarios. |

## Rejected Pattern Set

| Upstream pattern | Reason rejected |
| --- | --- |
| Skill installer or manual copy into `$CODEX_HOME/skills` | Conflicts with Namba repo-local skill ownership and duplicate-discovery avoidance. |
| Composio connect/connect-apps flows | Adds external CLI/OAuth dependencies outside this repository's Namba contract. |
| Slack, Notion, email, meeting, lead, invoice, or app automations | Outside requested scope and would introduce third-party workflow coupling. |
| Copying upstream skills wholesale | The request is to adapt useful ideas into current Namba skills, not vendor a skill collection. |

## Progressive Disclosure Rules

- Keep `SKILL.md` focused on when to trigger and the shortest reliable workflow.
- Move detailed procedure into a reference or helper candidate when a skill body would otherwise become a manual.
- Reference files, if added, must live under an existing Namba-owned skill or `.namba/codex/` path and must be generated or manifest-tracked consistently.
- Helper scripts, if added, must be callable without loading source into context for normal use.

## PR And Review Evidence Rules

- Resolve the PR from the current branch unless the user provides a PR.
- Fetch unresolved review threads with thread-level state.
- Assign each thread one of: `fixed-and-resolved`, `answered-open`, or `skipped-with-rationale`.
- Reply on the original thread with changed paths, commit or diff summary when available, validation commands, and CI/check evidence when relevant.
- Resolve only threads that were fixed or conclusively answered.
- Inspect PR checks before handoff; for failing GitHub Actions checks, capture run URL and bounded failure snippets.
- For external checks, report status and URL without attempting provider-specific automation.
- Ensure the configured Codex review marker exists exactly once.

## Deterministic Helper Candidate Criteria

A helper candidate is implementation-ready only if it:

- has a clear command name, `--help`, and documented input/output shape
- can run in read-only mode by default
- supports fixture-based tests or local dummy-server tests
- emits bounded text or JSON suitable for summarization
- does not silently mutate GitHub, repo files, or external systems
- makes network and authentication assumptions explicit
- avoids Composio, Slack, Notion, and other third-party app dependencies

## Harness And MCP Quality Criteria

- Prefer workflow tools over raw endpoint wrappers.
- Return high-signal concise output by default and offer detailed mode only when useful.
- Include pagination, truncation, or filtering expectations for large outputs.
- Make errors actionable by suggesting the next recoverable step.
- Design evaluations as independent, read-only, realistic, verifiable, and stable scenarios.
- Treat long-running MCP or frontend servers as managed processes; do not leave blocking server processes in the main validation path.

## Batch Migration Rules

- Inventory all affected generated and source surfaces before editing.
- Separate mechanical wording changes from behavioral changes.
- Update source templates first, regenerate, and review generated diffs.
- Add or update tests before relying on generated output inspection.
- Split the work if helper plumbing, trigger metadata, and documentation changes become hard to review together.
