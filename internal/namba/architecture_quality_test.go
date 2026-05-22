package namba

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInternalPackageDoesNotImportCommandGlue(t *testing.T) {
	t.Parallel()

	root := repoRootForHookTest(t)
	internalRoot := filepath.Join(root, "internal", "namba")
	err := filepath.WalkDir(internalRoot, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imported := range file.Imports {
			pathValue := strings.Trim(imported.Path.Value, `"`)
			if strings.Contains(pathValue, "/cmd/namba") {
				t.Fatalf("%s imports command glue package %q", path, pathValue)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk internal package: %v", err)
	}
}

func TestQualityScriptAndCIExposeHarnessGateArtifacts(t *testing.T) {
	t.Parallel()

	root := repoRootForHookTest(t)
	scriptData, err := os.ReadFile(filepath.Join(root, "scripts", "quality.sh"))
	if err != nil {
		t.Fatalf("read quality script: %v", err)
	}
	ciData, err := os.ReadFile(filepath.Join(root, ".github", "workflows", "ci.yml"))
	if err != nil {
		t.Fatalf("read CI workflow: %v", err)
	}
	script := string(scriptData)
	ci := string(ciData)
	for _, want := range []string{
		"--scorecard-out",
		"--summary-out",
		"Report JSON artifact",
		"Evidence schema validation",
		"coverage.out",
		"coverage.txt",
		"eval-scorecard.json",
		"eval-summary.md",
		"report.json",
		"schema-validation.txt",
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("quality script missing %q", want)
		}
	}
	for _, want := range []string{
		"Upload harness quality artifacts",
		"eval-scorecard.json",
		"eval-summary.md",
		"report.json",
		"schema-validation.txt",
		"coverage.out",
		"coverage.txt",
		"govulncheck",
		"staticcheck",
		"go test -race ./...",
	} {
		if !strings.Contains(ci, want) {
			t.Fatalf("CI workflow missing %q", want)
		}
	}
}
