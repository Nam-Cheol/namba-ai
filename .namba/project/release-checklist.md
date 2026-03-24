# Release Checklist

- [ ] `namba regen` rerun if template-generated Codex assets changed
- [ ] `namba sync` artifacts refreshed
- [ ] `namba pr` used for the GitHub review handoff when the branch is ready
- [ ] README and `.namba/codex/README.md` reflect update, release, and parallel workflow behavior
- [ ] Working tree is clean and the current branch is `main`
- [ ] Validation commands passed
- [ ] `namba release --version vX.Y.Z` or `namba release --bump patch` executed
- [ ] If `--push` was not used, `main` and the release tag were pushed manually
- [ ] GitHub Release workflow completed and published assets plus `checksums.txt`
