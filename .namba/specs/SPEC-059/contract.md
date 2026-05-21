# SPEC-059 Report Contract

## Command Contract

`namba report` is a read-only local observability command.

Supported v1 forms:

```sh
namba report
namba report --format markdown
namba report --format text
namba report --format json
namba report --json
namba report --spec SPEC-059
namba report --since 7d
namba report --fail-on blocked
namba status --json
```

`namba status` without flags must keep its current text output.

## JSON Contract

The full report JSON schema is `namba-report/v1`.

Required top-level fields:

- `schema_version`
- `generated_at`
- `project`
- `summary`
- `runs`
- `queue`
- `specs`
- `release`
- `diagnostics`
- `issues`
- `warnings`

Schema compatibility rule:

- fields present in v1 must keep their names and meanings;
- future additions must be optional;
- missing local data must be represented as `missing`, `none`, `unknown`, or warning entries instead of omitted failure paths.

## Health Contract

Health states:

- `ok`: no current blockers and enough evidence exists to summarize the workspace.
- `attention`: issues exist but no active blocker is detected.
- `blocked`: active queue, validation, review, or evidence state requires human recovery.
- `unknown`: there is not enough evidence to determine health.

## Data-State Contract

Parser state values:

- `present`
- `missing`
- `not_applicable`
- `corrupt`
- `unknown`

## Safety Contract

Report generation must:

- read only local workspace files;
- avoid network calls;
- avoid GitHub, Codex, browser, and external observability vendor calls;
- avoid mutating `.namba` or source files;
- treat path escapes outside `.namba` as warnings.

## Human Output Contract

The report must lead with:

1. health
2. top counts
3. top issues
4. queue state
5. latest run or evidence status
6. review readiness
7. stale candidates
8. diagnostics
9. next recommended action
