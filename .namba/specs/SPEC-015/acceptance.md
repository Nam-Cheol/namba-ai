# Acceptance

- [ ] In a clean repository where only transient runtime or cache paths such as `.gocache/` or `.tmp/` are present, running `namba sync` does not create tracked changes in generated artifacts, including `.namba/project/*` and `.namba/manifest.json`.
- [ ] Running `namba sync` twice in a row without meaningful source changes leaves the working tree clean after the second run.
- [ ] When README management is enabled, one `namba sync` run refreshes `README.md`, `README.ko.md`, `README.ja.md`, and `README.zh.md` from the current renderer content.
- [ ] Matching generated guide docs for the configured languages, including install/update lifecycle guidance, stay in sync with the current renderer contract during `namba sync`.
- [ ] `namba --version` works without repository context and prints a single concise line in the form `namba <version>`.
- [ ] Tagged release builds inject the release tag as the CLI version label, and local development builds fall back to the literal non-release label `dev`.
- [ ] `namba update` without `--version` resolves the latest GitHub Release asset for the current OS and architecture, and pinned updates via `--version vX.Y.Z` continue to work.
- [ ] On a successful `namba update`, terminal output includes the target version label, enough asset or platform context for the user to understand what was updated, and the correct next-step guidance, including restart guidance when replacement is deferred on Windows.
- [ ] On an update failure, terminal output or error text includes the requested version context plus actionable guidance about missing releases/assets, checksum failure, or download/network failure.
- [ ] Generated README bundles document install, update, and uninstall flows consistently for supported platforms in English, Korean, Japanese, and Chinese.
- [ ] Existing English README and workflow-guide behavior is preserved while adding multilingual parity guarantees.
- [ ] Validation commands pass
- [ ] Tests covering transient artifact exclusion, multilingual README sync behavior, version reporting, and update behavior are present
