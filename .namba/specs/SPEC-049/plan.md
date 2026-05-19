# SPEC-049 Plan

1. Refresh implementation context.
   - Inspect `install.sh`, `install.ps1`, release workflow artifacts, README
     install sections, release documentation, and current CI test workflows.
   - Confirm existing platform, architecture, asset-name, and version-selection
     behavior before changing installer logic.
2. Run plan review.
   - Run product, engineering, and design review artifacts under
     `.namba/specs/SPEC-049/reviews/`.
   - Refresh `reviews/readiness.md` before implementation starts.
3. Add Unix checksum verification.
   - Derive archive and `checksums.txt` URLs from the same pinned/latest mode.
   - Download archive and checksum file into temporary files, with test-only
     local fixture overrides if needed.
   - Parse checksum lines by exact basename and support normal, binary-marker,
     and `./asset` formats.
   - Compute SHA-256 using `sha256sum`, `shasum -a 256`, or optionally
     `openssl dgst -sha256`.
   - Fail before extraction on download/read failure, missing checksum,
     missing SHA-256 tool, or digest mismatch.
   - Extract only after verification into a temporary directory and install
     only the expected `namba` binary.
4. Add Windows checksum verification.
   - Derive archive and `checksums.txt` URLs from the same pinned/latest mode.
   - Download archive and checksum file into temporary files, with test-only
     local fixture overrides if needed.
   - Parse checksum lines by exact basename.
   - Use `Get-FileHash -Algorithm SHA256`.
   - Fail before `Expand-Archive` on download/read failure, missing checksum,
     or digest mismatch.
   - Extract only after verification into a temporary directory and install
     only the expected `namba.exe` binary.
5. Add archive safety checks.
   - Ensure installation copies only the expected binary from the extraction
     directory.
   - Reject suspicious archive entries when practical, including absolute paths
     and `..` traversal.
   - Fail clearly when the expected binary is missing.
6. Add deterministic installer tests.
   - Add `tests/installers/test_install_sh.py` covering matching checksum,
     mismatch, missing checksum line, no extraction before verification,
     missing `namba`, and exact asset matching.
   - Add `tests/installers/test_install_ps1.py` covering matching checksum,
     mismatch, missing checksum line, no `Expand-Archive` before verification,
     missing `namba.exe`, and exact asset matching.
   - Use Python standard library fixture creation with `tempfile`,
     `subprocess`, `hashlib`, `tarfile`, `zipfile`, and `pathlib`.
   - Skip PowerShell tests explicitly and visibly when `pwsh` or PowerShell is
     unavailable.
7. Update CI.
   - Ensure `python3 -m unittest discover -s tests` runs in CI.
   - Keep shellcheck optional unless the workflow installs it or already has it.
8. Update documentation.
   - Update README install documentation and localized install docs if present.
   - Update release documentation if present to require `checksums.txt` and
     exact asset-name entries.
   - Confirm release workflow continues to publish or require `checksums.txt`.
9. Validate.
   - Run `python3 -m unittest discover -s tests`.
   - Run `go test ./...`.
   - Run `go vet ./...`.
   - Run existing formatting check:
     `gofmt -l "cmd" "internal" "namba_test.go"`.
   - Search for unchecked extraction paths and insecure checksum bypass names.
10. Sync artifacts.
    - Run `namba sync` after implementation and validation.
