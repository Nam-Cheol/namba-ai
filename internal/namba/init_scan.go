package namba

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type initRepositoryScan struct {
	hasJava         bool
	sourceFiles     int
	testFiles       int
	goFormatTargets []string
}

func scanInitRepository(root string) initRepositoryScan {
	goTargets := make(map[string]bool)
	sourceFiles := 0
	testFiles := 0
	hasJava := false

	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}

		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return nil
		}
		rel = filepath.ToSlash(filepath.Clean(rel))
		ext := filepath.Ext(path)

		sourceFiles, testFiles, hasJava = scanFileForInitSignals(rel, ext, sourceFiles, testFiles, hasJava, goTargets)
		return nil
	})

	targets := make([]string, 0, len(goTargets))
	for target := range goTargets {
		targets = append(targets, target)
	}
	sort.Strings(targets)

	return initRepositoryScan{
		hasJava:         hasJava,
		sourceFiles:     sourceFiles,
		testFiles:       testFiles,
		goFormatTargets: targets,
	}
}

func scanFileForInitSignals(rel, ext string, sourceFiles, testFiles int, hasJava bool, goTargets map[string]bool) (int, int, bool) {
	if ext == ".java" {
		hasJava = true
	}
	if !pathWithinRepositorySubtree(rel, ".namba") {
		switch ext {
		case ".go", ".java", ".js", ".ts", ".tsx", ".py", ".rs":
			sourceFiles++
			if strings.Contains(strings.ToLower(filepath.Base(rel)), "test") {
				testFiles++
			}
		}
	}
	if ext != ".go" {
		return sourceFiles, testFiles, hasJava
	}

	topLevel := firstRepositoryPathSegment(rel)
	switch topLevel {
	case ".git", ".namba", ".codex", "external", "vendor":
		return sourceFiles, testFiles, hasJava
	}
	if strings.Contains(rel, "/") {
		goTargets[strconv.Quote(topLevel)] = true
		return sourceFiles, testFiles, hasJava
	}
	goTargets[strconv.Quote(filepath.Base(rel))] = true
	return sourceFiles, testFiles, hasJava
}

func firstRepositoryPathSegment(rel string) string {
	if rel == "" {
		return ""
	}
	parts := strings.Split(rel, "/")
	return parts[0]
}

func pathWithinRepositorySubtree(rel, subtree string) bool {
	return rel == subtree || strings.HasPrefix(rel, subtree+"/")
}

func detectLanguageFramework(root string) (string, string) {
	return detectLanguageFrameworkWithScan(root, scanInitRepository(root))
}

func detectLanguageFrameworkWithScan(root string, scan initRepositoryScan) (string, string) {
	switch {
	case exists(filepath.Join(root, "go.mod")):
		return "go", "none"
	case exists(filepath.Join(root, "pom.xml")) ||
		exists(filepath.Join(root, "build.gradle")) ||
		exists(filepath.Join(root, "build.gradle.kts")) ||
		exists(filepath.Join(root, "gradlew")) ||
		exists(filepath.Join(root, "gradlew.bat")) ||
		scan.hasJava:
		return "java", detectJavaFramework(root)
	case exists(filepath.Join(root, "package.json")):
		return "typescript", detectNodeFramework(root)
	case exists(filepath.Join(root, "pyproject.toml")) || exists(filepath.Join(root, "requirements.txt")):
		return "python", "none"
	default:
		return "unknown", "none"
	}
}

func detectJavaFramework(root string) string {
	if exists(filepath.Join(root, "pom.xml")) {
		data, err := os.ReadFile(filepath.Join(root, "pom.xml"))
		if err == nil && strings.Contains(strings.ToLower(string(data)), "spring-boot") {
			return "spring-boot"
		}
		return "maven"
	}

	for _, name := range []string{"build.gradle", "build.gradle.kts"} {
		if !exists(filepath.Join(root, name)) {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, name))
		if err == nil && strings.Contains(strings.ToLower(string(data)), "spring-boot") {
			return "spring-boot"
		}
		return "gradle"
	}

	if exists(filepath.Join(root, "gradlew")) || exists(filepath.Join(root, "gradlew.bat")) {
		return "gradle"
	}

	return "none"
}

func detectNodeFramework(root string) string {
	data, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err != nil {
		return "none"
	}
	text := string(data)
	switch {
	case strings.Contains(text, "\"next\""):
		return "nextjs"
	case strings.Contains(text, "\"react\""):
		return "react"
	case strings.Contains(text, "\"vue\""):
		return "vue"
	default:
		return "none"
	}
}

func detectMethodology(root string) string {
	return detectMethodologyWithScan(scanInitRepository(root))
}

func detectMethodologyWithScan(scan initRepositoryScan) string {
	if scan.sourceFiles == 0 {
		return "tdd"
	}
	if float64(scan.testFiles)/float64(scan.sourceFiles) >= 0.10 {
		return "tdd"
	}
	return "ddd"
}

func defaultQualityCommands(root, language, framework string) (string, string, string) {
	return defaultQualityCommandsWithScan(root, language, framework, scanInitRepository(root))
}

func defaultQualityCommandsWithScan(root, language, framework string, scan initRepositoryScan) (string, string, string) {
	switch language {
	case "go":
		return "go test ./...", defaultGoFormatCommandWithScan(scan), "go vet ./..."
	case "java":
		return defaultJavaQualityCommands(root, framework)
	case "typescript":
		return "npm test", "npm run lint", "npm run typecheck"
	case "python":
		return "pytest", "ruff check .", "none"
	default:
		return "none", "none", "none"
	}
}

func defaultJavaQualityCommands(root, framework string) (string, string, string) {
	switch normalizeFramework(framework) {
	case "spring-boot", "maven":
		return "mvn -q test", "mvn -q spotless:check", "mvn -q -DskipTests compile"
	case "gradle":
		gradle := defaultGradleCommand(root)
		return gradle + " test", gradle + " check", gradle + " compileJava"
	default:
		detected := detectJavaFramework(root)
		if detected != "none" {
			return defaultJavaQualityCommands(root, detected)
		}
		return "none", "none", "none"
	}
}

func defaultGradleCommand(root string) string {
	switch {
	case exists(filepath.Join(root, "gradlew")):
		return "./gradlew"
	case exists(filepath.Join(root, "gradlew.bat")):
		return ".\\gradlew.bat"
	default:
		return "gradle"
	}
}

func defaultGoFormatCommand(root string) string {
	return defaultGoFormatCommandWithScan(scanInitRepository(root))
}

func defaultGoFormatCommandWithScan(scan initRepositoryScan) string {
	if len(scan.goFormatTargets) == 0 {
		return "none"
	}
	return "gofmt -l " + strings.Join(scan.goFormatTargets, " ")
}
