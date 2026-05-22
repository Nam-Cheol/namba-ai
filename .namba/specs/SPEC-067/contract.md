# Refactor Contract

## Behavior-Preserving Contract

The implementation may move code, rename internal helpers, introduce same-package files, and introduce narrowly scoped subpackages only when dependency direction is proven. It must not change public behavior.

## Fixed Public Surface

- Public commands and help topics.
- Flag and argument semantics.
- stdout and stderr contract phrases.
- Exit success and failure behavior.
- Generated file layout under `.namba`, `.codex`, README, docs, logs, reports, eval artifacts, and releases.
- JSON schemas and evidence schemas.
- SPEC scaffold layout and SPEC semantics.
- Queue semantics.
- Run/evidence semantics.
- Guardrail and security blocking behavior.
- PR, land, and release semantics.
- Idempotency behavior.

## Flexible Internal Surface

- File ownership inside `internal/namba`.
- Helper function names.
- Same-package file boundaries.
- Private structs and pure helper signatures.
- Internal adapters, as long as existing tests or stronger characterization tests preserve public behavior.
- Minimal subpackages after dependency analysis proves no cycles and no public behavior drift.

## Forbidden Implementation Shortcuts

- Do not delete existing tests.
- Do not weaken assertions only to make refactor pass.
- Do not relax coverage threshold.
- Do not skip `scripts/quality.sh`.
- Do not add live external API tests.
- Do not hide failures or convert blocked unsafe behavior into success.
- Do not change generated layouts to simplify code movement.

## Dependency Direction

Preferred first step is same-package extraction:

```text
cmd/namba -> internal/namba
internal/namba files share package namba
```

Only later, if justified:

```text
internal/namba -> internal/namba/output
internal/namba -> internal/namba/specs
internal/namba -> internal/namba/project
internal/namba -> internal/namba/evidence
```

Subpackages must not import `internal/namba` back.
