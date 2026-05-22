# Frontend Brief

Task Classification: frontend-minor
Classification Rationale: This is a CLI terminal onboarding UX redesign, not a browser or app frontend. The design review should focus on terminal information hierarchy, localization, and plain-output resilience.
Frontend Gate Status: not-applicable
Problem Gate: not-applicable
Reference Gate: not-applicable
Critique Gate: not-applicable
Decision Gate: not-applicable
Prototype Gate: not-applicable
Prototype Evidence: n/a
Frontend implementation phase: not-applicable
Asset mode: not-applicable
Imagegen requirement: not-applicable
Asset decision proof: Terminal CLI output does not need generated bitmap assets; visual work is text, symbols, ANSI styling, and ASCII fallback.

## CLI UX Brief

- Primary surface: `namba init` interactive terminal wizard.
- Design risk: users currently see Korean explanatory copy before selecting language, and flag emoji language labels can render as broken glyph pairs in some terminals.
- Direction: compact language-first entry, then localized step copy with a restrained visual vocabulary that works with and without emoji.
- Visual grammar: no country flags, no mandatory full-screen layout, no nested cards, no decorative bloat. Use progress, labels, defaults, and recommendations as functional cues.
- Accessibility and resilience: line prompts, numeric choices, default echoes, ASCII/plain mode, and no reliance on color alone.
- Evidence target: tests should assert the generated terminal text contains the new structural cues and excludes flag emoji.
