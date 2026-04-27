package namba

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const releaseNotesDir = ".namba/releases"

var releaseCommitPrefixPattern = regexp.MustCompile(`^(feat|fix|docs|chore|refactor|test|ci|build|perf|style|revert)(\([^)]+\))?:\s*`)
var releaseReferencePattern = regexp.MustCompile(`(?i)\bPR\s*#\d+\b|\bSPEC-\d+\b|#\d+\b`)
var releaseSpecIDPattern = regexp.MustCompile(`(?i)\bSPEC-\d+\b`)
var releaseBodyDetailPattern = regexp.MustCompile(`^\s*[-*]\s+(.+)$`)
var releaseAcceptanceItemPattern = regexp.MustCompile(`^\s*-\s+\[[ xX]\]\s+(.+)$`)

const (
	maxReleaseDetailsPerCommit  = 12
	maxReleaseDetailsPerSection = 3
)

type releaseNoteCategory string

const (
	releaseNoteCategoryUserVisible releaseNoteCategory = "user-visible"
	releaseNoteCategoryFixes       releaseNoteCategory = "fixes"
	releaseNoteCategoryDocs        releaseNoteCategory = "docs-workflow"
	releaseNoteCategoryInternal    releaseNoteCategory = "internal-maintenance"
)

type releaseCommit struct {
	Hash      string
	ShortHash string
	Subject   string
	Body      string
	Category  releaseNoteCategory
	Refs      []string
	Details   []string
}

type releaseTarget struct {
	GOOS      string
	GOARCH    string
	Archive   string
	AssetName string
}

func releaseTargets() []releaseTarget {
	return []releaseTarget{
		{GOOS: "windows", GOARCH: "386", Archive: "zip", AssetName: "namba_Windows_x86.zip"},
		{GOOS: "windows", GOARCH: "amd64", Archive: "zip", AssetName: "namba_Windows_x86_64.zip"},
		{GOOS: "windows", GOARCH: "arm64", Archive: "zip", AssetName: "namba_Windows_arm64.zip"},
		{GOOS: "linux", GOARCH: "amd64", Archive: "tar.gz", AssetName: "namba_Linux_x86_64.tar.gz"},
		{GOOS: "linux", GOARCH: "arm64", Archive: "tar.gz", AssetName: "namba_Linux_arm64.tar.gz"},
		{GOOS: "darwin", GOARCH: "amd64", Archive: "tar.gz", AssetName: "namba_macOS_x86_64.tar.gz"},
		{GOOS: "darwin", GOARCH: "arm64", Archive: "tar.gz", AssetName: "namba_macOS_arm64.tar.gz"},
	}
}

func releaseAssetName(goos, goarch string) (string, error) {
	for _, target := range releaseTargets() {
		if target.GOOS == goos && target.GOARCH == goarch {
			return target.AssetName, nil
		}
	}
	return "", fmt.Errorf("unsupported release target %s/%s", goos, goarch)
}

func releaseNotesPath(version string) string {
	return filepath.ToSlash(filepath.Join(releaseNotesDir, version+".md"))
}

func previousReleaseTag(tags []string, version string) (string, error) {
	target, err := parseSemver(version)
	if err != nil {
		return "", err
	}

	var previous *semver
	for _, tag := range tags {
		parsed, err := parseSemver(tag)
		if err != nil {
			continue
		}
		if compareSemver(parsed, target) >= 0 {
			continue
		}
		if previous == nil || compareSemver(parsed, *previous) > 0 {
			value := parsed
			previous = &value
		}
	}
	if previous == nil {
		return "", nil
	}
	return formatSemver(*previous), nil
}

func collectReleaseCommits(ctx context.Context, app *App, root, previousTag string) ([]releaseCommit, error) {
	args := []string{"log", "--no-merges", "--reverse", "--format=%H%x00%h%x00%s%x00%b%x1e"}
	if previousTag != "" {
		args = append(args, previousTag+"..HEAD")
	} else {
		args = append(args, "HEAD")
	}

	output, err := app.runBinary(ctx, "git", args, root)
	if err != nil {
		return nil, fmt.Errorf("collect release history: %w", err)
	}

	commits, err := parseReleaseCommits(output)
	if err != nil {
		return nil, err
	}
	return enrichReleaseCommitsWithSpecDetails(root, commits), nil
}

func parseReleaseCommits(output string) ([]releaseCommit, error) {
	records := strings.Split(output, "\x1e")
	commits := make([]releaseCommit, 0, len(records))
	for _, record := range records {
		record = strings.TrimSpace(record)
		if record == "" {
			continue
		}

		fields := strings.SplitN(record, "\x00", 4)
		if len(fields) != 4 {
			return nil, fmt.Errorf("invalid release log record: %q", record)
		}

		commit := releaseCommit{
			Hash:      strings.TrimSpace(fields[0]),
			ShortHash: strings.TrimSpace(fields[1]),
			Subject:   strings.TrimSpace(fields[2]),
			Body:      strings.TrimSpace(fields[3]),
		}
		commit.Category = categorizeReleaseCommit(commit.Subject, commit.Body)
		commit.Refs = releaseCommitReferences(commit.Subject, commit.Body)
		if isReleaseNotesPrepCommit(commit) {
			continue
		}
		commits = append(commits, commit)
	}
	return commits, nil
}

func enrichReleaseCommitsWithSpecDetails(root string, commits []releaseCommit) []releaseCommit {
	enriched := make([]releaseCommit, len(commits))
	copy(enriched, commits)
	for i := range enriched {
		details := releaseCommitBodyDetails(enriched[i].Body)
		for _, specID := range releaseCommitSpecIDs(enriched[i]) {
			details = append(details, readSpecReleaseDetails(root, specID)...)
		}
		enriched[i].Details = limitReleaseDetails(dedupeReleaseDetails(details), maxReleaseDetailsPerCommit)
	}
	return enriched
}

func releaseCommitBodyDetails(body string) []string {
	var details []string
	for _, line := range strings.Split(body, "\n") {
		match := releaseBodyDetailPattern.FindStringSubmatch(line)
		if len(match) != 2 {
			continue
		}
		detail := normalizeReleaseDetailText(match[1])
		if shouldIncludeReleaseDetail("", detail) {
			details = append(details, detail)
		}
	}
	return details
}

func releaseCommitSpecIDs(commit releaseCommit) []string {
	text := strings.Join(append([]string{commit.Subject, commit.Body}, commit.Refs...), "\n")
	matches := releaseSpecIDPattern.FindAllString(text, -1)
	if len(matches) == 0 {
		return nil
	}

	seen := make(map[string]bool)
	specIDs := make([]string, 0, len(matches))
	for _, match := range matches {
		specID := strings.ToUpper(strings.TrimSpace(match))
		if specID == "" || seen[specID] {
			continue
		}
		seen[specID] = true
		specIDs = append(specIDs, specID)
	}
	sort.Strings(specIDs)
	return specIDs
}

func readSpecReleaseDetails(root, specID string) []string {
	acceptancePath := filepath.Join(root, ".namba", "specs", specID, "acceptance.md")
	if content, err := os.ReadFile(acceptancePath); err == nil {
		if details := parseAcceptanceReleaseDetails(string(content)); len(details) > 0 {
			return details
		}
	}

	specPath := filepath.Join(root, ".namba", "specs", specID, "spec.md")
	if content, err := os.ReadFile(specPath); err == nil {
		return parseSpecReleaseDetails(string(content))
	}
	return nil
}

func parseAcceptanceReleaseDetails(content string) []string {
	currentSection := ""
	sectionOrder := make([]string, 0)
	sectionItems := make(map[string][]string)
	var unsectioned []string

	for _, line := range strings.Split(content, "\n") {
		if heading, ok := releaseMarkdownHeading(line); ok {
			if strings.EqualFold(heading, "Acceptance") {
				currentSection = ""
			} else {
				currentSection = heading
				if _, exists := sectionItems[currentSection]; !exists {
					sectionOrder = append(sectionOrder, currentSection)
				}
			}
			continue
		}

		match := releaseAcceptanceItemPattern.FindStringSubmatch(line)
		if len(match) != 2 {
			continue
		}
		detail := normalizeReleaseDetailText(match[1])
		if !shouldIncludeReleaseDetail(currentSection, detail) {
			continue
		}
		if currentSection == "" {
			unsectioned = append(unsectioned, detail)
			continue
		}
		sectionItems[currentSection] = append(sectionItems[currentSection], detail)
	}

	if len(sectionOrder) == 0 {
		return limitReleaseDetails(unsectioned, maxReleaseDetailsPerCommit)
	}

	var details []string
	details = append(details, unsectioned...)
	for _, section := range sectionOrder {
		items := sectionItems[section]
		for i, item := range items {
			if i >= maxReleaseDetailsPerSection {
				break
			}
			details = append(details, fmt.Sprintf("%s: %s", section, item))
			if len(details) >= maxReleaseDetailsPerCommit {
				return details
			}
		}
	}
	return limitReleaseDetails(details, maxReleaseDetailsPerCommit)
}

func parseSpecReleaseDetails(content string) []string {
	capturing := false
	var details []string
	for _, line := range strings.Split(content, "\n") {
		if heading, ok := releaseMarkdownHeading(line); ok {
			switch strings.ToLower(heading) {
			case "goal", "desired outcome", "scope":
				capturing = true
			default:
				capturing = false
			}
			continue
		}
		if !capturing {
			continue
		}
		match := releaseBodyDetailPattern.FindStringSubmatch(line)
		if len(match) != 2 {
			continue
		}
		detail := normalizeReleaseDetailText(match[1])
		if shouldIncludeReleaseDetail("", detail) {
			details = append(details, detail)
		}
		if len(details) >= maxReleaseDetailsPerCommit {
			break
		}
	}
	return details
}

func releaseMarkdownHeading(line string) (string, bool) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "#") {
		return "", false
	}
	trimmed := strings.TrimLeft(line, "#")
	if len(trimmed) == len(line) || !strings.HasPrefix(trimmed, " ") {
		return "", false
	}
	heading := strings.TrimSpace(trimmed)
	return heading, heading != ""
}

func normalizeReleaseDetailText(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "- ")
	return strings.TrimSpace(value)
}

func shouldIncludeReleaseDetail(section, detail string) bool {
	if detail == "" {
		return false
	}
	normalizedSection := strings.ToLower(section)
	if strings.Contains(normalizedSection, "test") || strings.Contains(normalizedSection, "validation") {
		return false
	}
	normalizedDetail := strings.ToLower(detail)
	skipped := []string{
		"validation commands pass",
		"gofmt",
		"go test",
		"go vet",
		"git diff --check",
	}
	for _, needle := range skipped {
		if strings.Contains(normalizedDetail, needle) {
			return false
		}
	}
	return true
}

func dedupeReleaseDetails(details []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(details))
	for _, detail := range details {
		key := strings.ToLower(strings.TrimSpace(detail))
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, strings.TrimSpace(detail))
	}
	return result
}

func limitReleaseDetails(details []string, limit int) []string {
	if limit <= 0 || len(details) <= limit {
		return details
	}
	return details[:limit]
}

func categorizeReleaseCommit(subject, body string) releaseNoteCategory {
	normalized := strings.ToLower(strings.TrimSpace(subject))
	switch {
	case strings.HasPrefix(normalized, "fix"), strings.HasPrefix(normalized, "hotfix"), strings.HasPrefix(normalized, "bugfix"), strings.Contains(normalized, " regression "):
		return releaseNoteCategoryFixes
	case strings.HasPrefix(normalized, "docs"), strings.HasPrefix(normalized, "doc"), strings.Contains(normalized, "workflow"), strings.Contains(normalized, "readme"), strings.Contains(normalized, "guide"), strings.Contains(normalized, "sync"), strings.Contains(normalized, "release notes"), strings.Contains(normalized, "codex"), strings.Contains(normalized, "pr "):
		return releaseNoteCategoryDocs
	case strings.HasPrefix(normalized, "feat"), strings.HasPrefix(normalized, "feature"), strings.HasPrefix(normalized, "add"), strings.HasPrefix(normalized, "implement"), strings.Contains(normalized, "user-facing"), strings.Contains(normalized, "ui"), strings.Contains(normalized, "ux"):
		return releaseNoteCategoryUserVisible
	case strings.HasPrefix(normalized, "chore"), strings.HasPrefix(normalized, "refactor"), strings.HasPrefix(normalized, "test"), strings.HasPrefix(normalized, "ci"), strings.HasPrefix(normalized, "build"), strings.Contains(normalized, "dependency"), strings.Contains(normalized, "maintenance"), strings.Contains(normalized, "cleanup"):
		return releaseNoteCategoryInternal
	default:
		lowerBody := strings.ToLower(body)
		switch {
		case strings.Contains(lowerBody, "workflow"), strings.Contains(lowerBody, "readme"), strings.Contains(lowerBody, "release notes"), strings.Contains(lowerBody, "spec-"), strings.Contains(lowerBody, "codex"):
			return releaseNoteCategoryDocs
		case strings.Contains(normalized, "fix") || strings.Contains(lowerBody, "fix"):
			return releaseNoteCategoryFixes
		case strings.Contains(normalized, "release") || strings.Contains(normalized, "ship") || strings.Contains(normalized, "launch"):
			return releaseNoteCategoryUserVisible
		default:
			return releaseNoteCategoryInternal
		}
	}
}

func releaseCommitReferences(subject, body string) []string {
	text := subject + "\n" + body
	matches := releaseReferencePattern.FindAllString(text, -1)
	if len(matches) == 0 {
		return nil
	}

	seen := make(map[string]bool)
	refs := make([]string, 0, len(matches))
	for _, match := range matches {
		ref := strings.TrimSpace(match)
		if ref == "" || seen[ref] {
			continue
		}
		seen[ref] = true
		refs = append(refs, ref)
	}
	sort.Strings(refs)
	return refs
}

func isReleaseNotesPrepCommit(commit releaseCommit) bool {
	normalized := strings.ToLower(commit.Subject + "\n" + commit.Body)
	return strings.Contains(normalized, "[namba-release-notes]") || strings.Contains(normalized, "skip-release-notes")
}

func normalizeReleaseSubject(subject string) string {
	subject = strings.TrimSpace(subject)
	if subject == "" {
		return "변경 내용"
	}

	if match := releaseCommitPrefixPattern.FindStringIndex(subject); match != nil {
		subject = strings.TrimSpace(subject[match[1]:])
	}

	replacements := []string{
		"Merge pull request",
		"Merge branch",
	}
	for _, replacement := range replacements {
		if strings.HasPrefix(subject, replacement) {
			return subject
		}
	}

	return subject
}

func renderReleaseNotes(version, previousTag string, commits []releaseCommit) string {
	lines := []string{fmt.Sprintf("# %s 릴리즈 노트", version), ""}
	if previousTag != "" {
		lines = append(lines, fmt.Sprintf("%s 이후 변경 사항입니다.", previousTag), "")
	}

	grouped := map[releaseNoteCategory][]releaseCommit{
		releaseNoteCategoryUserVisible: nil,
		releaseNoteCategoryFixes:       nil,
		releaseNoteCategoryDocs:        nil,
		releaseNoteCategoryInternal:    nil,
	}
	for _, commit := range commits {
		grouped[commit.Category] = append(grouped[commit.Category], commit)
	}

	sections := []struct {
		title    string
		category releaseNoteCategory
	}{
		{title: "사용자에게 보이는 변경", category: releaseNoteCategoryUserVisible},
		{title: "수정", category: releaseNoteCategoryFixes},
		{title: "문서 및 워크플로", category: releaseNoteCategoryDocs},
		{title: "내부 정비", category: releaseNoteCategoryInternal},
	}

	for _, section := range sections {
		items := grouped[section.category]
		if len(items) == 0 {
			continue
		}

		lines = append(lines, fmt.Sprintf("## %s", section.title), "")
		for _, commit := range items {
			lines = append(lines, fmt.Sprintf("- %s%s", normalizeReleaseSubject(commit.Subject), renderReleaseCommitSuffix(commit)))
			for _, detail := range commit.Details {
				lines = append(lines, fmt.Sprintf("  - %s", detail))
			}
		}
		lines = append(lines, "")
	}

	if len(commits) == 0 {
		lines = append(lines, "- 이전 릴리스 이후 커밋이 없어 자동 요약을 만들지 못했습니다.", "")
	}

	return strings.Join(lines, "\n")
}

func renderReleaseCommitSuffix(commit releaseCommit) string {
	var parts []string
	for _, ref := range commit.Refs {
		parts = append(parts, ref)
	}
	if commit.ShortHash != "" {
		parts = append(parts, commit.ShortHash)
	}
	if len(parts) == 0 {
		return ""
	}
	return fmt.Sprintf(" (%s)", strings.Join(parts, ", "))
}

func writeReleaseNotes(root, version, content string) (string, error) {
	path := releaseNotesPath(version)
	absPath := filepath.Join(root, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return "", fmt.Errorf("create release notes directory: %w", err)
	}
	if err := os.WriteFile(absPath, []byte(content+"\n"), 0o644); err != nil {
		return "", fmt.Errorf("write release notes: %w", err)
	}
	return path, nil
}
