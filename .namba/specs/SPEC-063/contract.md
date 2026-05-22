# SPEC-063 Contract

## Runtime Contract

- `namba sync` is the durable path for the new verification guide and the README/workflow-guide links.
- Generated docs must carry the existing generated-file warning convention.
- The durable source must be a renderer, template, sync config, or equivalent source-of-truth input.
- Manual edits to generated README or workflow-guide files are allowed only as regenerated outputs.

## Verification Guide Contract

The guide must explain:

- `namba eval`: when to run it, what it measures, how regressions are reported, and what to do next.
- `namba report`: how reports are produced or consumed and what fields or summaries matter.
- `scripts/quality.sh`: local parity with CI, main checks, default artifact paths, and failure behavior.
- `.namba` evidence: project, run, queue, hook, review, and report evidence locations at a practical operator level.
- CI consumption: how CI mirrors or consumes local verification and artifacts.
- Failure interpretation: map command failures, schema/report failures, eval regressions, coverage failures, missing tools, and drift to concrete next actions.

## Final Report Contract

- The final hook may require `# NAMBA-AI 작업 결과 보고` and the configured sections.
- The hook must instruct Codex to preserve detailed content from the prior answer inside that frame.
- Required preserved detail classes: changed files, commands run, validation results, artifact paths, unresolved blockers, risks, and the next concrete command or handoff.
- The hook must not encourage replacing a detailed answer with a short generic frame.

## Compatibility

- Preserve public CLI semantics for `namba eval`, `namba report`, `scripts/quality.sh`, CI, and evidence generation.
- Existing generated docs should remain stable except for intentional verification-guide links and related text.
- Existing output-contract validation should continue accepting concise reports while gaining coverage for detail preservation.
