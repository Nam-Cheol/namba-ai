package namba

import (
	"path/filepath"
	"strings"
)

type syncSupportContext struct {
	LatestSpec               string
	LatestReadinessPath      string
	LatestReadinessSummary   string
	HasLatestReadiness       bool
	LatestExecutionProof     executionEvidenceManifest
	LatestExecutionProofPath string
	HasLatestExecutionProof  bool
}

func (a *App) buildSyncSupportContext(root, latestSpec string, readinessAdvisories map[string]string) syncSupportContext {
	support := syncSupportContext{
		LatestSpec: normalizedLatestSpec(latestSpec),
	}
	if support.LatestSpec != "none" {
		support.LatestReadinessPath = specReviewReadinessPath(support.LatestSpec)
		if readinessAdvisories != nil {
			if advisory, ok := readinessAdvisories[support.LatestSpec]; ok {
				support.HasLatestReadiness = true
				support.LatestReadinessSummary = advisory
			}
		} else if specReviewReadinessExists(root, support.LatestSpec) {
			support.HasLatestReadiness = true
			states := loadSpecReviewStatesWithReadFile(a.readFile, filepath.Join(root, specsDir, support.LatestSpec))
			support.LatestReadinessSummary = specReviewAdvisorySummaryFromStates(states)
		}
	}
	if manifest, relPath, ok := a.latestExecutionEvidence(root); ok {
		support.HasLatestExecutionProof = true
		support.LatestExecutionProof = manifest
		support.LatestExecutionProofPath = relPath
	}
	return support
}

func buildChangeSummaryDocWithSupport(projectCfg projectConfig, profile initProfile, support syncSupportContext) string {
	lines := changeSummaryHeaderLines(projectCfg, support.LatestSpec)
	lines = append(lines, "")
	lines = append(lines, changeSummaryWorkflowDocsSection(profile)...)
	lines = append(lines, "")
	lines = append(lines, changeSummaryRefreshCommandsSection()...)
	if readinessLines := changeSummaryLatestReviewReadinessSectionFromSupport(support); len(readinessLines) > 0 {
		lines = append(lines, "")
		lines = append(lines, readinessLines...)
	}
	if proofLines := changeSummaryLatestExecutionProofSectionFromSupport(support); len(proofLines) > 0 {
		lines = append(lines, "")
		lines = append(lines, proofLines...)
	}
	return stringsJoinLines(lines)
}

func buildPRChecklistDocWithSupport(profile initProfile, support syncSupportContext) string {
	lines := prChecklistHeaderLines()
	lines = append(lines, prChecklistCoreItems(profile)...)
	lines = append(lines, prChecklistLatestReviewReadinessItemFromSupport(support)...)
	lines = append(lines, prChecklistLatestExecutionProofItemFromSupport(support)...)
	return stringsJoinLines(lines)
}

func changeSummaryLatestReviewReadinessSectionFromSupport(support syncSupportContext) []string {
	if !support.HasLatestReadiness {
		return nil
	}
	return []string{
		"## Latest Review Readiness",
		"",
		"- Latest readiness artifact: `" + support.LatestReadinessPath + "`",
		"- Advisory summary: " + support.LatestReadinessSummary,
	}
}

func changeSummaryLatestExecutionProofSectionFromSupport(support syncSupportContext) []string {
	if !support.HasLatestExecutionProof {
		return nil
	}
	manifest := support.LatestExecutionProof
	return []string{
		"## Latest Execution Proof",
		"",
		"- Latest execution proof artifact: `" + support.LatestExecutionProofPath + "`",
		"- Proof target: `" + fallbackOrValue(manifest.SpecID, "unknown") + "`",
		"- Execution proof status: `" + manifest.Status + "`",
		"- Execution mode: `" + fallbackOrValue(manifest.ExecutionMode, "unknown") + "`",
		"- Base artifacts: `request=" + string(manifest.Request.State) + "`, `preflight=" + string(manifest.Preflight.State) + "`, `execution=" + string(manifest.Execution.State) + "`, `validation=" + string(manifest.Validation.State) + "`, `progress=" + string(manifest.Progress.State) + "`",
		"- Browser evidence: `" + string(manifest.Extensions.Browser.State) + "`",
		"- Runtime evidence: `" + string(manifest.Extensions.Runtime.State) + "`",
	}
}

func prChecklistLatestReviewReadinessItemFromSupport(support syncSupportContext) []string {
	if !support.HasLatestReadiness {
		return nil
	}
	return []string{"- [ ] Latest SPEC review readiness checked: `" + support.LatestReadinessPath + "`"}
}

func prChecklistLatestExecutionProofItemFromSupport(support syncSupportContext) []string {
	if !support.HasLatestExecutionProof {
		return nil
	}
	manifest := support.LatestExecutionProof
	return []string{"- [ ] Latest execution proof checked: `" + support.LatestExecutionProofPath + "` (`" + manifest.Status + "`, target `" + fallbackOrValue(manifest.SpecID, "unknown") + "`)"}
}

func isSyncProjectSupportManagedPath(rel string) bool {
	switch rel {
	case filepath.ToSlash(projectDir + "/change-summary.md"),
		filepath.ToSlash(projectDir + "/pr-checklist.md"),
		filepath.ToSlash(projectDir + "/release-notes.md"),
		filepath.ToSlash(projectDir + "/release-checklist.md"):
		return true
	default:
		return false
	}
}

func stringsJoinLines(lines []string) string {
	return strings.Join(lines, "\n") + "\n"
}
