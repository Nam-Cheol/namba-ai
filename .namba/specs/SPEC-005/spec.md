# SPEC-005

## Goal

Ship release-based installation so users can install NambaAI without Go and run it as a global `namba` command.

## Scope

- Add GitHub Release packaging for Windows, Linux, and macOS.
- Add installer scripts that download a release asset instead of requiring `go install`.
- Install the binary into a user-local bin directory and register that directory on PATH.
- Update the public README so `namba` is the default invocation style.
- Keep build-from-source instructions as a secondary developer path.

## Non-Goals

- Homebrew, winget, apt, or chocolatey packages.
- Automatic self-update.
- Signed binaries.

## Constraints

- The primary installation path must not require a local Go toolchain.
- Windows examples should use the global `namba` command, not `./namba.exe`.
- Installer scripts and docs must remain UTF-8.