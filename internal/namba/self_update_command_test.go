package namba

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseUpdateArgs(t *testing.T) {
	t.Parallel()

	opts, err := parseUpdateArgs([]string{"--version", "v1.2.3"})
	if err != nil {
		t.Fatalf("parseUpdateArgs returned error: %v", err)
	}
	if opts.Version != "v1.2.3" {
		t.Fatalf("opts.Version = %q, want %q", opts.Version, "v1.2.3")
	}
}

func TestRunUpdateReplacesExecutableOnUnix(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	execPath := filepath.Join(tmp, "namba")
	if err := os.WriteFile(execPath, []byte("old-binary"), 0o755); err != nil {
		t.Fatalf("write executable: %v", err)
	}

	stdout := &bytes.Buffer{}
	app := NewApp(stdout, &bytes.Buffer{})
	app.goos = "linux"
	app.goarch = "amd64"
	app.executablePath = func() (string, error) { return execPath, nil }
	app.downloadURL = func(_ context.Context, url string) ([]byte, error) {
		want := releaseDownloadURL("latest", "namba_Linux_x86_64.tar.gz")
		if url != want {
			t.Fatalf("download url = %q, want %q", url, want)
		}
		return makeTarGzArchive(t, "namba", []byte("new-binary")), nil
	}

	if err := app.Run(context.Background(), []string{"update"}); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	updated, err := os.ReadFile(execPath)
	if err != nil {
		t.Fatalf("read executable: %v", err)
	}
	if string(updated) != "new-binary" {
		t.Fatalf("updated executable = %q, want %q", string(updated), "new-binary")
	}
	if !strings.Contains(stdout.String(), "Updated NambaAI to latest") {
		t.Fatalf("unexpected stdout: %q", stdout.String())
	}
}

func TestRunUpdateSchedulesReplacementOnWindows(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	execPath := filepath.Join(tmp, "namba.exe")
	if err := os.WriteFile(execPath, []byte("old-binary"), 0o755); err != nil {
		t.Fatalf("write executable: %v", err)
	}

	stdout := &bytes.Buffer{}
	app := NewApp(stdout, &bytes.Buffer{})
	app.goos = "windows"
	app.goarch = "amd64"
	app.executablePath = func() (string, error) { return execPath, nil }
	app.downloadURL = func(_ context.Context, url string) ([]byte, error) {
		want := releaseDownloadURL("v1.2.3", "namba_Windows_x86_64.zip")
		if url != want {
			t.Fatalf("download url = %q, want %q", url, want)
		}
		return makeZipArchive(t, "namba.exe", []byte("new-binary")), nil
	}

	var startedName string
	var startedArgs []string
	app.startCmd = func(name string, args []string, dir string) error {
		startedName = name
		startedArgs = append([]string(nil), args...)
		return nil
	}

	if err := app.Run(context.Background(), []string{"update", "--version", "v1.2.3"}); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	if startedName != "powershell" {
		t.Fatalf("started command = %q, want %q", startedName, "powershell")
	}
	scriptPath := startedArgs[len(startedArgs)-1]
	script := mustReadFile(t, scriptPath)
	if !strings.Contains(script, execPath) || !strings.Contains(script, "Copy-Item") {
		t.Fatalf("unexpected helper script: %s", script)
	}

	existing, err := os.ReadFile(execPath)
	if err != nil {
		t.Fatalf("read executable: %v", err)
	}
	if string(existing) != "old-binary" {
		t.Fatalf("windows update should schedule replacement, got %q", string(existing))
	}
	if !strings.Contains(stdout.String(), "Scheduled NambaAI update to v1.2.3") {
		t.Fatalf("unexpected stdout: %q", stdout.String())
	}
}

func makeTarGzArchive(t *testing.T, name string, body []byte) []byte {
	t.Helper()

	var archive bytes.Buffer
	gz := gzip.NewWriter(&archive)
	tw := tar.NewWriter(gz)
	if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Size: int64(len(body))}); err != nil {
		t.Fatalf("write tar header: %v", err)
	}
	if _, err := tw.Write(body); err != nil {
		t.Fatalf("write tar body: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("close tar writer: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("close gzip writer: %v", err)
	}
	return archive.Bytes()
}

func makeZipArchive(t *testing.T, name string, body []byte) []byte {
	t.Helper()

	var archive bytes.Buffer
	zw := zip.NewWriter(&archive)
	writer, err := zw.Create(name)
	if err != nil {
		t.Fatalf("create zip entry: %v", err)
	}
	if _, err := writer.Write(body); err != nil {
		t.Fatalf("write zip entry: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}
	return archive.Bytes()
}
