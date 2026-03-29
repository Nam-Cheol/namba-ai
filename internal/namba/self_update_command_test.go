package namba

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseUpdateArgsDefaultsToLatestRelease(t *testing.T) {
	t.Parallel()

	opts, err := parseUpdateArgs(nil)
	if err != nil {
		t.Fatalf("parseUpdateArgs returned error: %v", err)
	}
	if opts.Version != "latest" {
		t.Fatalf("opts.Version = %q, want %q", opts.Version, "latest")
	}
}

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

	archiveData := makeTarGzArchive(t, "namba", []byte("new-binary"))
	stdout := &bytes.Buffer{}
	app := NewApp(stdout, &bytes.Buffer{})
	app.goos = "linux"
	app.goarch = "amd64"
	app.executablePath = func() (string, error) { return execPath, nil }
	app.downloadURL = func(_ context.Context, url string) ([]byte, error) {
		switch url {
		case releaseDownloadURL("latest", "namba_Linux_x86_64.tar.gz"):
			return archiveData, nil
		case releaseDownloadURL("latest", "checksums.txt"):
			return checksumManifest("namba_Linux_x86_64.tar.gz", archiveData), nil
		default:
			t.Fatalf("unexpected download url %q", url)
			return nil, nil
		}
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
	if !strings.Contains(stdout.String(), "Downloading latest release using namba_Linux_x86_64.tar.gz for linux/amd64") {
		t.Fatalf("unexpected stdout: %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Updated NambaAI from dev to latest release using namba_Linux_x86_64.tar.gz for linux/amd64") {
		t.Fatalf("unexpected stdout: %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "run 'namba --version' to confirm") {
		t.Fatalf("unexpected stdout: %q", stdout.String())
	}
}

func TestRunUpdateFailsWhenChecksumDoesNotMatch(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	execPath := filepath.Join(tmp, "namba")
	if err := os.WriteFile(execPath, []byte("old-binary"), 0o755); err != nil {
		t.Fatalf("write executable: %v", err)
	}

	archiveData := makeTarGzArchive(t, "namba", []byte("new-binary"))
	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	app.goos = "linux"
	app.goarch = "amd64"
	app.executablePath = func() (string, error) { return execPath, nil }
	app.downloadURL = func(_ context.Context, url string) ([]byte, error) {
		switch url {
		case releaseDownloadURL("latest", "namba_Linux_x86_64.tar.gz"):
			return archiveData, nil
		case releaseDownloadURL("latest", "checksums.txt"):
			return []byte("0000000000000000000000000000000000000000000000000000000000000000  namba_Linux_x86_64.tar.gz\n"), nil
		default:
			t.Fatalf("unexpected download url %q", url)
			return nil, nil
		}
	}

	err := app.Run(context.Background(), []string{"update"})
	if err == nil {
		t.Fatal("expected checksum mismatch error")
	}
	if !strings.Contains(err.Error(), "cannot verify namba_Linux_x86_64.tar.gz for linux/amd64 while updating to latest release") {
		t.Fatalf("error = %v, want verification context", err)
	}
	if !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("error = %v, want checksum mismatch", err)
	}
	if !strings.Contains(err.Error(), "Check the release assets and try again") {
		t.Fatalf("error = %v, want actionable guidance", err)
	}

	updated, readErr := os.ReadFile(execPath)
	if readErr != nil {
		t.Fatalf("read executable: %v", readErr)
	}
	if string(updated) != "old-binary" {
		t.Fatalf("updated executable = %q, want existing binary to remain", string(updated))
	}
}

func TestRunUpdateSchedulesReplacementOnWindows(t *testing.T) {
	t.Parallel()

	testRunUpdateSchedulesReplacementOnWindows(t, "amd64", "namba_Windows_x86_64.zip")
}

func TestRunUpdateSchedulesReplacementOnWindows386(t *testing.T) {
	t.Parallel()

	testRunUpdateSchedulesReplacementOnWindows(t, "386", "namba_Windows_x86.zip")
}

func testRunUpdateSchedulesReplacementOnWindows(t *testing.T, goarch, assetName string) {
	t.Helper()

	tmp := t.TempDir()
	execPath := filepath.Join(tmp, "namba.exe")
	if err := os.WriteFile(execPath, []byte("old-binary"), 0o755); err != nil {
		t.Fatalf("write executable: %v", err)
	}

	archiveData := makeZipArchive(t, "namba.exe", []byte("new-binary"))
	stdout := &bytes.Buffer{}
	app := NewApp(stdout, &bytes.Buffer{})
	app.goos = "windows"
	app.goarch = goarch
	app.executablePath = func() (string, error) { return execPath, nil }
	app.downloadURL = func(_ context.Context, url string) ([]byte, error) {
		switch url {
		case releaseDownloadURL("v1.2.3", assetName):
			return archiveData, nil
		case releaseDownloadURL("v1.2.3", "checksums.txt"):
			return checksumManifest(assetName, archiveData), nil
		default:
			t.Fatalf("unexpected download url %q", url)
			return nil, nil
		}
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
	if !strings.Contains(stdout.String(), "Scheduled NambaAI update from dev to v1.2.3 using "+assetName+" for windows/"+goarch) {
		t.Fatalf("unexpected stdout: %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "run 'namba --version'") {
		t.Fatalf("unexpected stdout: %q", stdout.String())
	}
}

func TestRunUpdateIncludesLatestReleaseFailureContext(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	execPath := filepath.Join(tmp, "namba")
	if err := os.WriteFile(execPath, []byte("old-binary"), 0o755); err != nil {
		t.Fatalf("write executable: %v", err)
	}

	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	app.goos = "linux"
	app.goarch = "amd64"
	app.executablePath = func() (string, error) { return execPath, nil }
	app.downloadURL = func(_ context.Context, url string) ([]byte, error) {
		return nil, errors.New("download " + url + " failed with status 404: Not Found")
	}

	err := app.Run(context.Background(), []string{"update"})
	if err == nil {
		t.Fatal("expected latest release download error")
	}
	if !strings.Contains(err.Error(), "failed to download namba_Linux_x86_64.tar.gz for linux/amd64 from latest release (404)") {
		t.Fatalf("error = %v, want latest release context", err)
	}
	if !strings.Contains(err.Error(), "namba update --version vX.Y.Z") {
		t.Fatalf("error = %v, want actionable retry guidance", err)
	}
}

func checksumManifest(assetName string, archiveData []byte) []byte {
	sum := sha256.Sum256(archiveData)
	return []byte(hex.EncodeToString(sum[:]) + "  " + assetName + "\n")
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
