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
	return commits, nil
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
