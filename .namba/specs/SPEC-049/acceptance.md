# Acceptance

- [ ] `install.sh` downloads or reads the matching `checksums.txt` before
      extraction.
- [ ] `install.sh` verifies the archive SHA-256 checksum before extraction.
- [ ] `install.ps1` downloads or reads the matching `checksums.txt` before
      extraction.
- [ ] `install.ps1` verifies the archive SHA-256 checksum before extraction.
- [ ] Missing checksum lines fail closed.
- [ ] Checksum mismatches fail closed.
- [ ] No extraction occurs before checksum verification succeeds.
- [ ] No installation occurs before extraction succeeds.
- [ ] Unix installation copies only the expected `namba` binary.
- [ ] Windows installation copies only the expected `namba.exe` binary.
- [ ] Checksum matching is exact by asset basename, with no partial asset-name
      matches.
- [ ] Supported checksum formats include `<sha256>  <asset-name>`,
      `<sha256> *<asset-name>`, and `<sha256>  ./<asset-name>`.
- [ ] The Unix installer fails clearly when no SHA-256 tool is available.
- [ ] The Unix installer fails when the archive is missing `namba`.
- [ ] The Windows installer fails when the archive is missing `namba.exe`.
- [ ] Installer tests cover success, mismatch, missing checksum,
      no-extract-on-failure, missing expected binary, and exact asset matching.
- [ ] Installer tests do not use live GitHub network calls.
- [ ] Installer tests do not touch the real user install directory, `PATH`,
      shell profile, PowerShell profile, or home directory.
- [ ] Test-only installer environment variables, if added, are documented as
      test-only and do not weaken normal behavior.
- [ ] CI runs installer tests through `python3 -m unittest discover -s tests`.
- [ ] README install documentation states that installers verify release
      checksums before installing and fail closed on verification errors.
- [ ] Users can manually verify checksums if desired, without turning the
      quick-start install section into a long security lecture.
- [ ] Release documentation states that `checksums.txt` is required, if release
      documentation exists.
- [ ] Release archives are documented or confirmed as listed in
      `checksums.txt` by exact asset name.
- [ ] No insecure checksum bypass is introduced.
- [ ] `python3 -m unittest discover -s tests` passes.
- [ ] `go test ./...` passes.
- [ ] `go vet ./...` passes.
- [ ] Existing formatting check passes:
      `gofmt -l "cmd" "internal" "namba_test.go"`.
- [ ] No runtime logs, temporary archives, extracted files, or generated test
      artifacts are committed.
