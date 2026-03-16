# Release Notes Draft

Project: namba-ai
Project type: existing
Reference SPEC: SPEC-007
Generated: 2026-03-16T17:22:34+09:00

## Highlights

- Init wizard supports project type selection for new versus existing repositories.
- Init wizard includes Java as a primary language option.
- Interactive terminal selection supports arrow keys and Enter where the terminal allows it.
- `namba fix "<description>"` creates bugfix-oriented SPEC packages.
- `namba release` can create and optionally push a release tag.

## Release Command

```text
namba release --version vX.Y.Z
git push origin main
git push origin vX.Y.Z
```

## Expected Assets

- `namba_Windows_x86_64.zip`
- `namba_Windows_arm64.zip`
- `namba_Linux_x86_64.tar.gz`
- `namba_Linux_arm64.tar.gz`
- `namba_macOS_x86_64.tar.gz`
- `namba_macOS_arm64.tar.gz`
