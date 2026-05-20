# Acceptance

- [ ] Only `.github/workflows/release.yml` is changed for the implementation.
- [ ] The release workflow no longer uses `windows-latest`; the Windows runner is pinned to `windows-2022`.
- [ ] Node.js 20 deprecation warnings from release workflow action/runtime usage are removed.
- [ ] Existing release behavior is preserved.
- [ ] YAML/action static validation passes for the changed workflow.
- [ ] GitHub Actions execution is confirmed through a `workflow_dispatch` or PR check path that does not publish a real GitHub Release.
- [ ] Tag-push based validation is not used for this SPEC.
