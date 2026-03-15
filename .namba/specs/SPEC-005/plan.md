# SPEC-005 Plan

1. Add a deterministic release asset naming scheme and cover it with tests.
2. Add GitHub Actions release packaging for Windows, Linux, and macOS archives.
3. Add `install.ps1` and `install.sh` that download the latest or requested release asset.
4. Ensure the installers place the binary in a user-local bin directory and register PATH for `namba`.
5. Rewrite `README.md` so the default installation path is release-based and the default command form is `namba`.
6. Run validation commands and sync `.namba` project artifacts.