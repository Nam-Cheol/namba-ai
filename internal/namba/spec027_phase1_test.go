package namba

import (
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func TestNewExecutionRequestResolvesModeSpecificSessionModes(t *testing.T) {
	tests := []struct {
		name        string
		mode        executionMode
		sessionMode string
		want        string
	}{
		{name: "default keeps stateful", mode: executionModeDefault, sessionMode: "", want: "stateful"},
		{name: "solo rewrites stateful", mode: executionModeSolo, sessionMode: "", want: "solo"},
		{name: "team rewrites stateful", mode: executionModeTeam, sessionMode: "", want: "team"},
		{name: "parallel rewrites stateful", mode: executionModeParallel, sessionMode: "", want: "parallel-worker"},
		{name: "ephemeral stays explicit", mode: executionModeTeam, sessionMode: "ephemeral", want: "ephemeral"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := (&App{}).newExecutionRequest(
				"SPEC-027",
				t.TempDir(),
				"ship it",
				tt.mode,
				delegationPlan{},
				systemConfig{Runner: "", ApprovalPolicy: "", SandboxMode: ""},
				codexConfig{SessionMode: tt.sessionMode},
			)
			if req.Mode != normalizeExecutionMode(tt.mode) {
				t.Fatalf("expected normalized mode %q, got %q", normalizeExecutionMode(tt.mode), req.Mode)
			}
			if req.SessionMode != tt.want {
				t.Fatalf("expected session mode %q, got %q", tt.want, req.SessionMode)
			}
			if req.Runner != "codex" || req.ApprovalPolicy != "on-request" || req.SandboxMode != "workspace-write" {
				t.Fatalf("expected normalized runtime defaults, got %+v", req)
			}
		})
	}
}

func TestDetectLanguageFrameworkWithScanMatchesLegacyJavaFallback(t *testing.T) {
	root := t.TempDir()
	writeSpec027TestFile(t, filepath.Join(root, "src", "Main.java"), "class Main {}\n")

	scan := scanInitRepository(root)
	gotLanguage, gotFramework := detectLanguageFrameworkWithScan(root, scan)
	wantLanguage, wantFramework := legacyDetectLanguageFramework(root)
	if gotLanguage != wantLanguage || gotFramework != wantFramework {
		t.Fatalf("detectLanguageFrameworkWithScan() = (%q, %q), want (%q, %q)", gotLanguage, gotFramework, wantLanguage, wantFramework)
	}
}

func TestInitSharedScanMatchesLegacyMethodologyAndGoTargets(t *testing.T) {
	root := t.TempDir()
	writeSpec027TestFile(t, filepath.Join(root, "main.go"), "package main\n\nfunc main() {}\n")
	writeSpec027TestFile(t, filepath.Join(root, "cmd", "app", "main.go"), "package main\n\nfunc main() {}\n")
	writeSpec027TestFile(t, filepath.Join(root, "internal", "core", "service.go"), "package core\n")
	writeSpec027TestFile(t, filepath.Join(root, "pkg", "core", "service_test.go"), "package core\n")
	writeSpec027TestFile(t, filepath.Join(root, ".namba", "generated.go"), "package generated\n")

	scan := scanInitRepository(root)
	if got, want := detectMethodologyWithScan(scan), legacyDetectMethodology(root); got != want {
		t.Fatalf("detectMethodologyWithScan() = %q, want %q", got, want)
	}
	if got, want := defaultGoFormatCommandWithScan(scan), legacyDefaultGoFormatCommand(root); got != want {
		t.Fatalf("defaultGoFormatCommandWithScan() = %q, want %q", got, want)
	}
}

func BenchmarkSpec027LegacyInitDetection(b *testing.B) {
	root := buildSpec027InitFixture(b)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		language, framework := legacyDetectLanguageFramework(root)
		_, _, _ = legacyDefaultQualityCommands(root, language, framework)
		_ = legacyDetectMethodology(root)
	}
}

func BenchmarkSpec027SharedInitScan(b *testing.B) {
	root := buildSpec027InitFixture(b)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scan := scanInitRepository(root)
		language, framework := detectLanguageFrameworkWithScan(root, scan)
		_, _, _ = defaultQualityCommandsWithScan(root, language, framework, scan)
		_ = detectMethodologyWithScan(scan)
	}
}

func BenchmarkSpec027RunPreflightSetup(b *testing.B) {
	root := buildSpec027CommandWorkspace(b)
	app := NewApp(io.Discard, io.Discard)
	app.lookPath = func(name string) (string, error) {
		if name == "codex" || name == "git" {
			return name, nil
		}
		return "", fs.ErrNotExist
	}
	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		switch {
		case name == "codex" && len(args) == 1 && args[0] == "--version":
			return "codex-cli benchmark", nil
		case name == "codex" && len(args) == 2 && args[0] == "exec" && args[1] == "--help":
			return "-c, --config\n-s, --sandbox\n-m, --model\n-p, --profile\n--add-dir", nil
		case name == "codex" && len(args) == 3 && args[0] == "exec" && args[1] == "resume" && args[2] == "--help":
			return "-c, --config\n-m, --model", nil
		case isShellCommand(name):
			return "validation ok", nil
		default:
			return "", fs.ErrNotExist
		}
	}
	app.detectCodexCapabilities = func(context.Context, string, executionRequest) (codexCapabilityMatrix, error) {
		return testCodexCapabilities(), nil
	}

	req := app.newExecutionRequest(
		"SPEC-027",
		root,
		"phase-1",
		executionModeTeam,
		delegationPlan{IntegratorRole: "namba-implementer"},
		systemConfig{Runner: "codex", ApprovalPolicy: "on-request", SandboxMode: "workspace-write"},
		codexConfig{SessionMode: "stateful", RepairAttempts: 1},
	)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, _, err := app.runPreflight(context.Background(), req); err != nil {
			b.Fatalf("runPreflight failed: %v", err)
		}
	}
}

func BenchmarkSpec027ProjectCommand(b *testing.B) {
	root := buildSpec027CommandWorkspace(b)
	app := buildSpec027CommandApp(b, root)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := app.Run(context.Background(), []string{"project"}); err != nil {
			b.Fatalf("project failed: %v", err)
		}
	}
}

func BenchmarkSpec027SyncCommand(b *testing.B) {
	root := buildSpec027CommandWorkspace(b)
	app := buildSpec027CommandApp(b, root)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := app.Run(context.Background(), []string{"sync"}); err != nil {
			b.Fatalf("sync failed: %v", err)
		}
	}
}

func buildSpec027InitFixture(tb testing.TB) string {
	tb.Helper()

	root := tb.TempDir()
	writeSpec027TestFileTB(tb, filepath.Join(root, "go.mod"), "module example.com/spec027\n\ngo 1.22\n")
	for i := 0; i < 24; i++ {
		dir := filepath.Join(root, "cmd", "svc"+strconv.Itoa(i))
		writeSpec027TestFileTB(tb, filepath.Join(dir, "main.go"), "package main\n\nfunc main() {}\n")
	}
	for i := 0; i < 24; i++ {
		dir := filepath.Join(root, "internal", "pkg"+strconv.Itoa(i))
		writeSpec027TestFileTB(tb, filepath.Join(dir, "service.go"), "package pkg\n")
		writeSpec027TestFileTB(tb, filepath.Join(dir, "service_test.go"), "package pkg\n")
	}
	writeSpec027TestFileTB(tb, filepath.Join(root, ".namba", "generated.go"), "package generated\n")
	return root
}

func buildSpec027CommandWorkspace(tb testing.TB) string {
	tb.Helper()

	root := tb.TempDir()
	writeSpec027TestFileTB(tb, filepath.Join(root, "go.mod"), "module example.com/spec027\n\ngo 1.22\n")
	writeSpec027TestFileTB(tb, filepath.Join(root, "cmd", "app", "main.go"), "package main\n\nfunc main() {}\n")
	writeSpec027TestFileTB(tb, filepath.Join(root, "internal", "service", "service.go"), "package service\n")
	writeSpec027TestFileTB(tb, filepath.Join(root, "README.md"), "# SPEC-027\n")
	return root
}

func buildSpec027CommandApp(tb testing.TB, root string) *App {
	tb.Helper()

	app := NewApp(io.Discard, io.Discard)
	if err := app.Run(context.Background(), []string{"init", root, "--yes"}); err != nil {
		tb.Fatalf("init failed: %v", err)
	}
	app.getwd = func() (string, error) { return root, nil }
	return app
}

func writeSpec027TestFile(t *testing.T, path, body string) {
	t.Helper()
	writeSpec027TestFileTB(t, path, body)
}

func writeSpec027TestFileTB(tb testing.TB, path, body string) {
	tb.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		tb.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		tb.Fatalf("write %s: %v", path, err)
	}
}

func legacyDetectLanguageFramework(root string) (string, string) {
	switch {
	case exists(filepath.Join(root, "go.mod")):
		return "go", "none"
	case exists(filepath.Join(root, "pom.xml")) ||
		exists(filepath.Join(root, "build.gradle")) ||
		exists(filepath.Join(root, "build.gradle.kts")) ||
		exists(filepath.Join(root, "gradlew")) ||
		exists(filepath.Join(root, "gradlew.bat")) ||
		legacyTreeContainsExtension(root, ".java"):
		return "java", detectJavaFramework(root)
	case exists(filepath.Join(root, "package.json")):
		return "typescript", detectNodeFramework(root)
	case exists(filepath.Join(root, "pyproject.toml")) || exists(filepath.Join(root, "requirements.txt")):
		return "python", "none"
	default:
		return "unknown", "none"
	}
}

func legacyDetectMethodology(root string) string {
	source := 0
	tests := 0
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.Contains(path, string(filepath.Separator)+".namba"+string(filepath.Separator)) {
			return nil
		}
		switch filepath.Ext(path) {
		case ".go", ".java", ".js", ".ts", ".tsx", ".py", ".rs":
			source++
			if strings.Contains(strings.ToLower(filepath.Base(path)), "test") {
				tests++
			}
		}
		return nil
	})
	if source == 0 {
		return "tdd"
	}
	if float64(tests)/float64(source) >= 0.10 {
		return "tdd"
	}
	return "ddd"
}

func legacyTreeContainsExtension(root, ext string) bool {
	found := false
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ext {
			found = true
		}
		return nil
	})
	return found
}

func legacyDefaultGoFormatCommand(root string) string {
	skipDirs := map[string]bool{
		".git":     true,
		".namba":   true,
		".codex":   true,
		"external": true,
		"vendor":   true,
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return "none"
	}

	var targets []string
	for _, entry := range entries {
		name := entry.Name()
		switch {
		case entry.IsDir() && skipDirs[name]:
			continue
		case entry.IsDir() && legacyDirectoryContainsGo(filepath.Join(root, name)):
			targets = append(targets, strconv.Quote(name))
		case !entry.IsDir() && filepath.Ext(name) == ".go":
			targets = append(targets, strconv.Quote(name))
		}
	}

	if len(targets) == 0 {
		return "none"
	}

	sort.Strings(targets)
	return "gofmt -l " + strings.Join(targets, " ")
}

func legacyDirectoryContainsGo(root string) bool {
	found := false
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".go" {
			found = true
		}
		return nil
	})
	return found
}

func legacyDefaultQualityCommands(root, language, framework string) (string, string, string) {
	switch language {
	case "go":
		return "go test ./...", legacyDefaultGoFormatCommand(root), "go vet ./..."
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
