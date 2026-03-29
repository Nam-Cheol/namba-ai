package namba_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Nam-Cheol/namba-ai/internal/namba"
)

func TestSyncNoOpKeepsArtifactsStable(t *testing.T) {
	tmp := t.TempDir()
	app := namba.NewApp(&bytes.Buffer{}, &bytes.Buffer{})

	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tmp, ".gocache"), 0o755); err != nil {
		t.Fatalf("mkdir .gocache: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, ".gocache", "cache.txt"), []byte("cache"), 0o644); err != nil {
		t.Fatalf("write .gocache fixture: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tmp, ".tmp"), 0o755); err != nil {
		t.Fatalf("mkdir .tmp: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, ".tmp", "runtime.log"), []byte("temp"), 0o644); err != nil {
		t.Fatalf("write .tmp fixture: %v", err)
	}

	restore := chdir(t, tmp)
	defer restore()

	if err := app.Run(context.Background(), []string{"sync"}); err != nil {
		t.Fatalf("first sync failed: %v", err)
	}

	structurePath := filepath.Join(tmp, ".namba", "project", "structure.md")
	structure := mustRead(t, structurePath)
	for _, unwanted := range []string{".gocache", ".tmp"} {
		if strings.Contains(structure, unwanted) {
			t.Fatalf("expected structure doc to skip %q, got: %s", unwanted, structure)
		}
	}

	tracked := []string{
		filepath.Join(tmp, ".namba", "manifest.json"),
		filepath.Join(tmp, ".namba", "project", "change-summary.md"),
		filepath.Join(tmp, ".namba", "project", "release-notes.md"),
		structurePath,
		filepath.Join(tmp, "README.md"),
		filepath.Join(tmp, "README.ko.md"),
		filepath.Join(tmp, "README.ja.md"),
		filepath.Join(tmp, "README.zh.md"),
	}
	before := snapshotFiles(t, tracked)
	for _, path := range []string{
		filepath.Join(tmp, ".namba", "project", "change-summary.md"),
		filepath.Join(tmp, ".namba", "project", "release-notes.md"),
	} {
		if strings.Contains(before[path].Body, "Generated:") {
			t.Fatalf("expected %s to avoid generated timestamp churn, got: %s", path, before[path].Body)
		}
	}

	time.Sleep(1100 * time.Millisecond)

	if err := app.Run(context.Background(), []string{"sync"}); err != nil {
		t.Fatalf("second sync failed: %v", err)
	}

	after := snapshotFiles(t, tracked)
	for _, path := range tracked {
		if before[path].Body != after[path].Body {
			t.Fatalf("expected no-op sync to keep %s unchanged", path)
		}
		if !before[path].ModTime.Equal(after[path].ModTime) {
			t.Fatalf("expected no-op sync to avoid rewriting %s", path)
		}
	}
}

type fileSnapshot struct {
	Body    string
	ModTime time.Time
}

func snapshotFiles(t *testing.T, paths []string) map[string]fileSnapshot {
	t.Helper()

	snapshots := make(map[string]fileSnapshot, len(paths))
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat %s: %v", path, err)
		}
		snapshots[path] = fileSnapshot{
			Body:    mustRead(t, path),
			ModTime: info.ModTime(),
		}
	}
	return snapshots
}
