# Acceptance

- [x] `install.sh` downloads or reads the matching `checksums.txt` before
      extraction.
- [x] `install.sh` verifies the archive SHA-256 checksum before extraction.
- [x] `install.ps1` downloads or reads the matching `checksums.txt` before
      extraction.
- [x] `install.ps1` verifies the archive SHA-256 checksum before extraction.
- [x] Missing checksum lines fail closed.
- [x] Checksum mismatches fail closed.
- [x] No extraction occurs before checksum verification succeeds.
- [x] No installation occurs before extraction succeeds.
- [x] Unix installation copies only the expected `namba` binary.
- [x] Windows installation copies only the expected `namba.exe` binary.
- [x] Checksum matching is exact by asset basename, with no partial asset-name
      matches.
- [x] Supported checksum formats include `<sha256>  <asset-name>`,
      `<sha256> *<asset-name>`, and `<sha256>  ./<asset-name>`.
- [x] The Unix installer fails clearly when no SHA-256 tool is available.
- [x] The Unix installer fails when the archive is missing `namba`.
- [x] The Windows installer fails when the archive is missing `namba.exe`.
- [x] Installer tests cover success, mismatch, missing checksum,
      no-extract-on-failure, missing expected binary, and exact asset matching.
- [x] Installer tests do not use live GitHub network calls.
- [x] Installer tests do not touch the real user install directory, `PATH`,
      shell profile, PowerShell profile, or home directory.
- [x] Test-only installer environment variables, if added, are documented as
      test-only and do not weaken normal behavior.
- [x] CI runs installer tests through `python3 -m unittest discover -s tests`.
- [x] README install documentation states that installers verify release
      checksums before installing and fail closed on verification errors.
- [x] Users can manually verify checksums if desired, without turning the
      quick-start install section into a long security lecture.
- [x] Release documentation states that `checksums.txt` is required, if release
      documentation exists.
- [x] Release archives are documented or confirmed as listed in
      `checksums.txt` by exact asset name.
- [x] No insecure checksum bypass is introduced.
- [x] `python3 -m unittest discover -s tests` passes.
- [x] `go test ./...` passes.
- [x] `go vet ./...` passes.
- [x] Existing formatting check passes:
      `gofmt -l "cmd" "internal" "namba_test.go"`.
- [x] No runtime logs, temporary archives, extracted files, or generated test
      artifacts are committed.
