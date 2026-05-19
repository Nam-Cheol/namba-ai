# SPEC-049

## Title

Harden NambaAI installers with checksum verification before extraction

## Problem

NambaAI installers download release archives and install a local executable that
interacts with project and Codex workflow surfaces. Archive integrity is part of
the harness trust boundary, so installers must fail closed unless the downloaded
archive matches the release checksum for the exact expected asset.

## Goal

Verify downloaded release archives with SHA-256 checksums from the matching
release `checksums.txt` before any extraction or installation happens.

## Context

- Project: namba-ai
- Project type: existing
- Language: go
- Mode: tdd
- Work type: plan
- Target surfaces: `install.sh`, `install.ps1`, installer tests, installer CI
  workflow entries, README install documentation, localized install
  documentation if present, release documentation if present, and
  `.github/workflows/release.yml` only if needed to preserve `checksums.txt` as
  a required release artifact.

## Scope

- Add checksum verification to the Unix installer before archive extraction.
- Add checksum verification to the Windows installer before `Expand-Archive`.
- Download or read `checksums.txt` from the same release channel as the archive.
- Match checksum entries by exact asset basename.
- Extract archives only into temporary directories.
- Install only the expected binary: `namba` on Unix and `namba.exe` on Windows.
- Add deterministic installer tests using local fixtures, no live GitHub calls.
- Add CI coverage for the installer tests.
- Update install and release documentation where those docs exist.

## Out Of Scope

- Redesigning the release workflow.
- Adding a package manager.
- Requiring Go, GitHub authentication, binary signing, Sigstore/cosign, SLSA
  provenance, or SBOM generation for installation.
- Changing the installed binary name.
- Changing normal install locations except where needed for safe tests.
- Adding any insecure fallback or bypass, including `NAMBA_SKIP_CHECKSUM`,
  `-SkipChecksum`, or silent unchecked installation.

## Checksum Source Behavior

- Pinned versions use archive URL
  `https://github.com/Nam-Cheol/namba-ai/releases/download/$VERSION/$ASSET_NAME`
  and checksums URL
  `https://github.com/Nam-Cheol/namba-ai/releases/download/$VERSION/checksums.txt`.
- Latest installs use archive URL
  `https://github.com/Nam-Cheol/namba-ai/releases/latest/download/$ASSET_NAME`
  and checksums URL
  `https://github.com/Nam-Cheol/namba-ai/releases/latest/download/checksums.txt`.
- The archive and checksum file must be derived from the same pinned/latest
  mode so `latest` cannot mix with a pinned release.

## Verification Requirements

1. Preserve existing OS, architecture, asset-name, and version-selection
   behavior.
2. Download the release archive to a temporary file.
3. Download or read the matching `checksums.txt`.
4. Find the checksum line for the exact expected asset basename.
5. Compute the SHA-256 digest of the downloaded archive.
6. Compare expected and actual digests case-insensitively.
7. Fail before extraction if `checksums.txt` cannot be downloaded or read, the
   asset line is missing, no SHA-256 tool is available, or the digest differs.
8. Extract only after checksum verification succeeds.
9. Install only after extraction succeeds.
10. Clean up temporary archives, checksums, and extraction directories.

## Checksum Line Parsing

Support common checksum formats:

```text
<sha256>  <asset-name>
<sha256> *<asset-name>
<sha256>  ./<asset-name>
```

Matching must be exact by basename. A checksum for
`namba-ai-darwin-arm64.tar.gz` must not satisfy
`namba-ai-darwin-amd64.tar.gz`, and the installer must not select the first
checksum line blindly.

## Unix Installer Requirements

- Keep `install.sh` POSIX `/bin/sh` compatible unless the script already
  intentionally requires Bash.
- Verify the archive before extraction.
- Support `sha256sum`, `shasum -a 256`, and optionally
  `openssl dgst -sha256`.
- Fail clearly when no checksum tool exists.
- Extract to a temporary directory.
- Install only the expected `namba` binary.
- Fail when the archive does not contain the expected binary.
- Avoid writing runtime artifacts to the repository.

## Windows Installer Requirements

- Verify the archive before `Expand-Archive`.
- Use `Get-FileHash -Algorithm SHA256`.
- Download or read `checksums.txt` from the same release channel as the archive.
- Match the exact asset basename.
- Fail when the checksum line is missing or mismatched.
- Extract to a temporary directory.
- Install only the expected `namba.exe` binary.
- Fail when the archive does not contain the expected binary.

## Archive Safety

- Extract only into temporary directories.
- Install only `namba` or `namba.exe`; do not install arbitrary archive files.
- Reject suspicious archive entries when practical, including absolute paths
  and `..` traversal.
- At minimum, tests must prove installation fails when the expected binary is
  missing.

## Test Strategy

- Use Python standard library `unittest` where practical.
- Suggested locations:
  - `tests/installers/test_install_sh.py`
  - `tests/installers/test_install_ps1.py`
- Use local fixture archives and fixture checksums.
- Use temporary directories for archives, checksum files, extraction, and
  installation.
- Do not touch the real install directory, `PATH`, shell profile, PowerShell
  profile, or home directory.
- Do not use live GitHub network calls.
- Allowed test-only environment variables, if needed:
  - `NAMBA_INSTALL_TEST_ASSET_PATH`
  - `NAMBA_INSTALL_TEST_CHECKSUMS_PATH`
  - `NAMBA_INSTALL_TEST_DIR`
  - `NAMBA_INSTALL_TEST_VERSION`
  - `NAMBA_INSTALL_TEST_OS`
  - `NAMBA_INSTALL_TEST_ARCH`
- Any test-only environment variables must be documented as test-only and must
  not weaken normal installer behavior.

## CI And Documentation

- Add CI coverage for installer tests.
- Do not make `shellcheck install.sh` a hard CI requirement unless the workflow
  installs shellcheck or already has it available.
- Update install documentation to state that installers verify release
  checksums before installing, checksum verification fails closed, and users can
  manually verify checksums if desired.
- Keep quick-start install commands concise.
- If release documentation exists, state that `checksums.txt` is a required
  release artifact, release archives must be listed in it, and installer
  verification depends on exact asset names.

## Risks

- Checksum parsing might accept partial asset-name matches; mitigate with exact
  basename tests.
- `latest` archive and `checksums.txt` might be derived from different release
  channels; mitigate by deriving both URLs from the same mode.
- POSIX compatibility might break; mitigate by keeping `install.sh` portable and
  testing with `sh`.
- PowerShell tests may be unavailable on some CI runners; skip explicitly and
  visibly when PowerShell is absent, while keeping Unix tests active.
- Tests might mutate the developer environment; mitigate with temporary install
  directories and test-only overrides.
- Incomplete releases will now fail installation; this is intended and should
  be paired with clear error messages.

## Follow-Up Backlog

- Add Sigstore/cosign signing for release artifacts.
- Add SLSA provenance for release artifacts.
- Add SBOM generation.
- Add advanced installer verification for attestations.
- Add package-manager distribution if adoption justifies it.
- Add a documented flow that downloads the installer first so users can inspect
  it before running it.
