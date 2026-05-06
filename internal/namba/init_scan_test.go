package namba

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestDefaultGoFormatCommandUsesSingleInventoryTargets(t *testing.T) {
	t.Parallel()

	root := canonicalTempDir(t)
	writeScanFixture(t, filepath.Join(root, "main.go"), "package main\n")
	writeScanFixture(t, filepath.Join(root, "cmd", "app", "main.go"), "package main\n")
	writeScanFixture(t, filepath.Join(root, "internal", "core", "value.go"), "package core\n")
	writeScanFixture(t, filepath.Join(root, ".namba", "generated.go"), "package ignored\n")
	writeScanFixture(t, filepath.Join(root, "vendor", "dep", "dep.go"), "package ignored\n")

	command := defaultGoFormatCommand(root)
	if command != `gofmt -l "cmd" "internal" "main.go"` {
		t.Fatalf("defaultGoFormatCommand = %q", command)
	}
}

func TestScanInitRepositorySkipsGeneratedAndDependencyTrees(t *testing.T) {
	t.Parallel()

	root := canonicalTempDir(t)
	writeScanFixture(t, filepath.Join(root, "src", "app.ts"), "export const app = true\n")
	writeScanFixture(t, filepath.Join(root, "src", "app.test.ts"), "test('app', () => {})\n")
	writeScanFixture(t, filepath.Join(root, "node_modules", ".pnpm", "dep", "index.ts"), "export const dep = true\n")
	writeScanFixture(t, filepath.Join(root, "vendor", "dep", "dep.go"), "package dep\n")
	writeScanFixture(t, filepath.Join(root, ".namba", "generated.go"), "package ignored\n")
	writeScanFixture(t, filepath.Join(root, ".codex", "agents", "generated.ts"), "export const ignored = true\n")
	writeScanFixture(t, filepath.Join(root, ".agents", "skills", "generated.ts"), "export const ignored = true\n")

	scan := scanInitRepository(root)
	if scan.sourceFiles != 2 {
		t.Fatalf("expected only source files outside generated/dependency trees, got %+v", scan)
	}
	if scan.testFiles != 1 {
		t.Fatalf("expected only real test files outside generated/dependency trees, got %+v", scan)
	}
	if len(scan.goFormatTargets) != 0 {
		t.Fatalf("expected dependency Go files to be ignored for formatting targets, got %+v", scan.goFormatTargets)
	}
}

func BenchmarkScanInitRepositoryLargeWorkspace(b *testing.B) {
	root := b.TempDir()
	for i := 0; i < 250; i++ {
		writeScanFixtureBenchmark(b, filepath.Join(root, "services", "svc"+itoa(i), "internal", "core", "file.go"), "package core\n")
		writeScanFixtureBenchmark(b, filepath.Join(root, "services", "svc"+itoa(i), "internal", "core", "file_test.go"), "package core\n")
	}
	for i := 0; i < 120; i++ {
		writeScanFixtureBenchmark(b, filepath.Join(root, "java", "pkg"+itoa(i), "Demo.java"), "class Demo {}\n")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = scanInitRepository(root)
	}
}

func writeScanFixture(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func writeScanFixtureBenchmark(b *testing.B, path, body string) {
	b.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		b.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		b.Fatalf("write %s: %v", path, err)
	}
}

func itoa(value int) string {
	return strconv.Itoa(value)
}
