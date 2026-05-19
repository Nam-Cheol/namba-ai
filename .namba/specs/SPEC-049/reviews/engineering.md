# Engineering Review

- Status: clear
- Last Reviewed: 2026-05-19
- Reviewer: namba-planner
- Command Skill: `$namba-plan-eng-review`
- Recommended Role: `namba-planner`

## Focus

- Lock architecture, sequencing, failure modes, trust boundaries, and validation strategy before execution starts.

## Findings

- The trust boundary and ordering are explicit enough: archive download,
  checksum download/read, exact checksum lookup, digest computation, and digest
  comparison must all complete before extraction.
- Exact basename matching is correctly called out as the central parsing risk.
  Tests must include near-miss assets such as `darwin-arm64` versus
  `darwin-amd64`.
- The Unix SHA-256 tool fallback order is realistic and remains portable when
  implemented in POSIX `sh`.
- The Windows requirement to use `Get-FileHash -Algorithm SHA256` is direct and
  testable.
- The proposed test hooks are narrow and safe: local asset path, local checksum
  path, temporary install dir, and test OS/arch/version overrides.
- CI risk is bounded by making PowerShell test skips explicit when PowerShell is
  unavailable while keeping Unix installer tests mandatory.
- The archive-safety requirement is intentionally modest for this SPEC:
  installation must copy only the expected binary, and traversal rejection is
  required when practical.

## Decisions

- Implement checksum parsing as an exact basename lookup, not a substring or
  first-line selection.
- Derive archive and checksum URLs from one version/latest decision point.
- Keep all extraction in temporary directories and copy only `namba` or
  `namba.exe` into the install directory.
- Prefer deterministic Python `unittest` coverage over live network or release
  fixture dependency.
- Keep shellcheck optional unless CI has an explicit shellcheck setup step.

## Follow-ups

- [non-blocking] During implementation, consider small helper functions for
  checksum-line parsing and SHA-256 computation so tests can exercise behavior
  without brittle whole-script assertions.
- [non-blocking] Confirm whether the existing release workflow already emits
  `checksums.txt`; update only the minimum release artifact contract if needed.
- [post-implementation] Add stronger archive traversal rejection if the first
  implementation lands with binary-only installation checks but not full entry
  validation.

## Recommendation

- Clear to proceed. The plan has enough sequencing, test seams, and validation
  coverage for implementation.
