# SPEC-059 Eval Plan

## Unit Fixture Matrix

Create temporary `.namba` fixtures that cover:

- empty `.namba`
- missing `.namba/logs`
- missing `.namba/logs/runs`
- corrupt execution evidence JSON
- corrupt validation JSON
- partially corrupt parallel progress JSONL
- completed run evidence
- execution failed evidence
- validation failed evidence
- hook failed evidence
- missing evidence refs
- queue none
- queue pending
- queue running
- queue waiting
- queue blocked
- queue done
- stale queue heartbeat
- ready review summary
- pending review summary
- missing review summary
- malformed readiness summary
- incomplete SPEC core files
- missing harness evidence files
- release checklist unavailable
- release checklist partially complete
- diagnostics unavailable
- diagnostics available with failed doctor

## CLI Tests

Cover:

- `namba report --help`
- `namba status --help`
- `namba report`
- `namba report --format markdown`
- `namba report --format text`
- `namba report --format json`
- `namba report --json`
- `namba status --json`
- unsupported report format
- unsupported `status` argument other than `--json`

## JSON Contract Tests

Assert:

- `schema_version` is `namba-report/v1`
- required top-level fields are present
- missing data is represented without command failure
- known fixture counts are stable
- issue severity and category fields are present
- output can be unmarshaled by Go tests without relying on field order

## Markdown/Text Tests

Assert output includes:

- health
- top issues
- blocked reason summary
- queue summary
- missing evidence summary
- validation failure summary
- review readiness summary
- release readiness summary
- diagnostics summary
- next recommended action

## Validation

Run the repository-wide Go test suite after implementation.
