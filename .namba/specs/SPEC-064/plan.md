# SPEC-064 Plan

1. Refresh context with `namba project` if implementation starts from stale project docs.
2. Add focused tests before implementation:
   - locale default selection for `NAMBA_LANG`, `LC_ALL`, `LANG`, and fallback `en`;
   - language-first step order and `b`/`back` behavior after the language step;
   - flag-free language labels and ASCII/plain rendering;
   - repo-state-first defaults for existing and empty repositories;
   - Codex access preview output remains semantically identical;
   - generated getting-started docs mention the new flow for `ko`, `en`, `ja`, and `zh` (Simplified Chinese).
3. Add a navigation test matrix before changing the state machine:
   - language step is first and has no back target;
   - project type and project details can go back to language after language selection;
   - manual Git skips provider and URL steps while preserving back to Git mode;
   - GitHub skips GitLab URL while preserving back to provider;
   - GitLab includes instance URL and preserves back through provider and Git mode.
4. Define the terminal capability and plain-output contract in helpers:
   - text markers for default, recommended, current, and next states;
   - no country flags in any mode;
   - no ANSI, color-only, or emoji-only meaning in plain mode;
   - line prompt fallback whenever raw-key selection or terminal capability is unreliable.
5. Refactor wizard copy and rendering helpers:
   - introduce localized wizard message tables for supported human languages;
   - introduce terminal visual helpers for emoji-safe and ASCII/plain modes;
   - keep raw-key select and line-select fallback behavior separate from visual styling.
6. Reorder interactive wizard flow so language selection is first while preserving default profile detection and later back navigation.
7. Add the onboarding surfaces:
   - multilingual language-first entry;
   - stage progress rail;
   - repo-intelligence header;
   - purpose-first setup paths with recommended badges;
   - Codex access explanation plus existing policy/sandbox preview;
   - Git setup guidance;
   - final ready handoff card.
8. Update help and generated docs:
   - `initUsageText`;
   - `internal/namba/readme.go` getting-started bootstrap section;
   - generated `docs/getting-started*.md` outputs if managed docs are refreshed.
9. Run validation:
   - `go test ./...`
   - `go vet ./...`
   - `scripts/quality.sh`
10. Run `namba sync` after implementation to refresh managed docs and readiness artifacts.
