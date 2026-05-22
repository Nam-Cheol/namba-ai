# SPEC-066 Baseline

## Previous State

Generated Namba instruction surfaces already existed, but contract details were distributed unevenly across command skills and custom-agent role cards. Some surfaces named state effects or responsibilities, while others did not consistently state pass/fail criteria, evidence expectations, destructive command approval rules, secrets handling, fallback implementer boundaries, or portability requirements.

## Regression Risk

- Manual edits to checked-in generated files could appear to fix this repository while leaving future `namba init` users unchanged.
- Updating only command-entry skills would leave lower-priority generated role cards and TOML agents with weaker guidance.
- Adding project-specific wording to core templates would leak this repository's assumptions into future repositories.

## Baseline Evidence

The implementation compares generated output through Go renderer tests, fresh temporary-repo `init`, current-repo `regen`, and temporary existing-repo `regen` checks.
