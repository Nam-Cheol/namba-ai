package namba

import (
	"fmt"
	"strings"
)

func buildSpecDoc(kind, specID, description string, projectCfg projectConfig, qualityCfg qualityConfig) string {
	switch kind {
	case "fix":
		return buildFixSpecDoc(specID, description, projectCfg, qualityCfg)
	case "harness":
		return buildHarnessSpecDoc(specID, description, projectCfg, qualityCfg)
	default:
		return buildFeatureSpecDoc(specID, description, projectCfg, qualityCfg)
	}
}

func buildSpecPlanDoc(kind, specID string) string {
	switch kind {
	case "fix":
		return buildFixSpecPlanDoc(specID)
	case "harness":
		return buildHarnessSpecPlanDoc(specID)
	default:
		return buildFeatureSpecPlanDoc(specID)
	}
}

func buildFixSpecDoc(specID, description string, projectCfg projectConfig, qualityCfg qualityConfig) string {
	return fmt.Sprintf("# %s\n\n## Problem\n\n%s\n\n## Goal\n\nApply the smallest safe fix that resolves the reported issue.\n\n## Context\n\n- Project: %s\n- Project type: %s\n- Language: %s\n- Mode: %s\n- Work type: fix\n", specID, description, projectCfg.Name, projectCfg.ProjectType, projectCfg.Language, qualityCfg.DevelopmentMode)
}

func buildHarnessSpecDoc(specID, description string, projectCfg projectConfig, qualityCfg qualityConfig) string {
	return fmt.Sprintf("# %s\n\n## Problem\n\nThe current repository needs a dedicated harness-oriented planning flow for the following request:\n\n%s\n\n## Goal\n\nDesign a Codex-native harness change under the existing `SPEC-XXX` artifact flow without inventing a second planning model or importing Claude-only runtime primitives.\n\n## Context\n\n- Project: %s\n- Project type: %s\n- Language: %s\n- Mode: %s\n- Work type: plan\n- Planning surface: `namba harness \"<description>\"`\n\n## Desired Outcome\n\n- `namba harness \"<description>\"` acts as a top-level planning command while `namba plan` keeps its current feature-planning behavior.\n- The scaffold captures Codex-native execution topology, agent/skill boundaries, progressive-disclosure guidance, trigger strategy, and evaluation strategy for reusable skills or agents.\n- Help and accidental-write safety stay aligned with the shared command-parsing contract instead of creating command-specific drift.\n- The planned output remains under `.namba/specs/<SPEC>` with the normal review artifacts.\n\n## Non-Goals\n\n- Do not create a second artifact model outside `.namba/specs/`.\n- Do not emit `.claude/*`, `TeamCreate`, `SendMessage`, `TaskCreate`, or a mandatory `model: \"opus\"` requirement as part of the Codex-facing contract.\n- Do not change the default behavior of `namba plan`.\n", specID, description, projectCfg.Name, projectCfg.ProjectType, projectCfg.Language, qualityCfg.DevelopmentMode)
}

func buildFeatureSpecDoc(specID, description string, projectCfg projectConfig, qualityCfg qualityConfig) string {
	return fmt.Sprintf("# %s\n\n## Problem\n\n%s\n\n## Goal\n\nImplement the requested change under the normal feature-planning workflow.\n\n## Context\n\n- Project: %s\n- Project type: %s\n- Language: %s\n- Mode: %s\n- Work type: plan\n", specID, description, projectCfg.Name, projectCfg.ProjectType, projectCfg.Language, qualityCfg.DevelopmentMode)
}

func buildFixSpecPlanDoc(specID string) string {
	return fmt.Sprintf("# %s Plan\n\n1. Refresh project context with `namba project`\n2. Reproduce or inspect the reported issue\n3. Run the relevant review passes under `.namba/specs/%s/reviews/` and refresh the readiness summary\n4. Implement the smallest safe fix\n5. Run validation commands and targeted regression checks\n6. Sync artifacts with `namba sync`\n", specID, specID)
}

func buildHarnessSpecPlanDoc(specID string) string {
	return fmt.Sprintf("# %s Plan\n\n1. Refresh project context with `namba project`\n2. Define the top-level `namba harness` command contract while keeping `namba plan` unchanged\n3. Capture Codex-native execution topology, agent/skill boundaries, progressive-disclosure layout, trigger guidance, and evaluation strategy in the scaffold\n4. Run the relevant review passes under `.namba/specs/%s/reviews/` and refresh the readiness summary\n5. Implement the requested command and scaffold changes\n6. Run validation commands\n7. Sync artifacts with `namba sync`\n", specID, specID)
}

func buildFeatureSpecPlanDoc(specID string) string {
	return fmt.Sprintf("# %s Plan\n\n1. Refresh project context with `namba project`\n2. Run the relevant review passes under `.namba/specs/%s/reviews/` and refresh the readiness summary\n3. Implement the requested change\n4. Run validation commands\n5. Sync artifacts with `namba sync`\n", specID, specID)
}

func buildSpecAcceptanceDoc(kind, description, mode string) string {
	if kind == "fix" {
		return buildFixAcceptanceDoc(description, mode)
	}
	if kind == "harness" {
		return buildHarnessAcceptanceDoc(description, mode)
	}
	return buildFeatureAcceptanceDoc(description, mode)
}

func buildFeatureAcceptanceDoc(description, mode string) string {
	bullets := featureAcceptanceCoreLines(description)
	bullets = append(bullets, featureAcceptanceModeLine(mode))
	return strings.Join(bullets, "\n")
}

func buildHarnessAcceptanceDoc(description, mode string) string {
	bullets := harnessAcceptanceCoreLines(description)
	bullets = append(bullets, harnessAcceptanceModeLine(mode))
	return strings.Join(bullets, "\n")
}

func featureAcceptanceCoreLines(description string) []string {
	return []string{
		"# Acceptance",
		"",
		"- [ ] The requested behavior described below is implemented:",
		"  " + description,
		"- [ ] Validation commands pass",
	}
}

func featureAcceptanceModeLine(mode string) string {
	if mode == "tdd" {
		return "- [ ] Tests covering the new behavior are present"
	}
	return "- [ ] Existing behavior is preserved while improving the target area"
}

func harnessAcceptanceCoreLines(description string) []string {
	return []string{
		"# Acceptance",
		"",
		"- [ ] `namba harness \"<description>\"` creates the next sequential `SPEC-XXX` package with a harness-oriented scaffold.",
		"- [ ] `namba plan \"<description>\"` keeps its current default feature-planning behavior.",
		"- [ ] `namba harness --help` is read-only and does not create or mutate `.namba/specs/<SPEC>`.",
		"- [ ] The generated scaffold captures Codex-native execution topology, agent/skill boundaries, progressive-disclosure guidance, trigger strategy, and evaluation strategy.",
		"- [ ] The generated scaffold stays on the existing `.namba/specs/<SPEC>` artifact model and does not invent a second planning package type.",
		"- [ ] The generated scaffold excludes Claude-only primitives such as `.claude/*`, `TeamCreate`, `SendMessage`, `TaskCreate`, and a mandatory `model: \"opus\"` requirement.",
		"- [ ] The requested harness-oriented behavior described below is represented in the scaffold:",
		"  " + description,
		"- [ ] Validation commands pass",
	}
}

func harnessAcceptanceModeLine(mode string) string {
	if mode == "tdd" {
		return "- [ ] Tests covering the new command/scaffold behavior are present"
	}
	return "- [ ] Existing planning behavior is preserved while adding the harness surface"
}

func buildFixAcceptanceDoc(description, mode string) string {
	bullets := fixAcceptanceCoreLines(description)
	bullets = append(bullets, fixAcceptanceModeLine(mode))
	return strings.Join(bullets, "\n")
}

func fixAcceptanceCoreLines(description string) []string {
	return []string{
		"# Acceptance",
		"",
		"- [ ] The reported issue described below is resolved:",
		"  " + description,
		"- [ ] Validation commands pass",
		"- [ ] Existing behavior around the affected area is preserved",
	}
}

func fixAcceptanceModeLine(mode string) string {
	if mode == "tdd" {
		return "- [ ] A regression test covering the fix is present"
	}
	return "- [ ] A targeted reproduction or verification step is documented"
}
