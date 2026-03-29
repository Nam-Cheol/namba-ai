# SPEC-015 Plan

1. Inspect the current `namba sync` flow, structure skip rules, and README replacement path to reproduce unstable outputs and localized bundle drift.
2. Run the relevant review passes under `.namba/specs/SPEC-015/reviews/` and refresh the readiness summary before implementation if the risk profile justifies it.
3. Design deterministic skip rules for structure and related synced docs so transient cache or temp paths do not enter tracked artifacts.
4. Update README bundle sync so every configured language and its generated guide docs refresh from the current renderer content in one pass.
5. Define the release/build-time version injection strategy so `namba --version` can report a trustworthy value for tagged releases and a sensible value for local development builds.
6. Add a top-level `namba --version` flow and wire it to the chosen version source without requiring repository context.
7. Verify and tighten the default `namba update` behavior so no-flag execution targets the latest GitHub Release while `--version` remains the explicit pinning path.
8. Improve `namba update` terminal UX so users can easily understand current versus target version, what asset/platform is being used, whether a restart is needed, and what to do on failure.
9. Add uninstall guidance to the generated README bundles and related generated install/update documentation across English, Korean, Japanese, and Chinese.
10. Add regression tests for transient artifact exclusion, multilingual README synchronization, version reporting, update defaults, and stable update UX copy.
11. Run validation commands and a targeted `namba sync` plus CLI lifecycle verification on the affected flows.
12. Refresh resulting artifacts with `namba sync` and review the final diff for noise-free output.
