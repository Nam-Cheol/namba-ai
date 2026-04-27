package namba

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestReleaseAssetName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		goos    string
		goarch  string
		want    string
		wantErr bool
	}{
		{
			name:   "windows 386",
			goos:   "windows",
			goarch: "386",
			want:   "namba_Windows_x86.zip",
		},
		{
			name:   "windows amd64",
			goos:   "windows",
			goarch: "amd64",
			want:   "namba_Windows_x86_64.zip",
		},
		{
			name:   "windows arm64",
			goos:   "windows",
			goarch: "arm64",
			want:   "namba_Windows_arm64.zip",
		},
		{
			name:   "linux amd64",
			goos:   "linux",
			goarch: "amd64",
			want:   "namba_Linux_x86_64.tar.gz",
		},
		{
			name:   "mac arm64",
			goos:   "darwin",
			goarch: "arm64",
			want:   "namba_macOS_arm64.tar.gz",
		},
		{
			name:    "unsupported",
			goos:    "freebsd",
			goarch:  "amd64",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := releaseAssetName(tt.goos, tt.goarch)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %s/%s", tt.goos, tt.goarch)
				}
				return
			}
			if err != nil {
				t.Fatalf("releaseAssetName(%q, %q) returned error: %v", tt.goos, tt.goarch, err)
			}
			if got != tt.want {
				t.Fatalf("releaseAssetName(%q, %q) = %q, want %q", tt.goos, tt.goarch, got, tt.want)
			}
		})
	}
}

func TestReleaseNotesPath(t *testing.T) {
	t.Parallel()

	if got, want := releaseNotesPath("v1.2.3"), ".namba/releases/v1.2.3.md"; got != want {
		t.Fatalf("releaseNotesPath() = %q, want %q", got, want)
	}
}

func TestWriteReleaseNotesUsesProjectRoot(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path, err := writeReleaseNotes(root, "v1.2.3", "# notes")
	if err != nil {
		t.Fatalf("writeReleaseNotes returned error: %v", err)
	}
	if path != ".namba/releases/v1.2.3.md" {
		t.Fatalf("writeReleaseNotes path = %q", path)
	}

	content, err := os.ReadFile(filepath.Join(root, ".namba", "releases", "v1.2.3.md"))
	if err != nil {
		t.Fatalf("read written release notes: %v", err)
	}
	if string(content) != "# notes\n" {
		t.Fatalf("release notes content = %q", content)
	}
}

func TestReleaseWorkflowUsesNotesBodyPath(t *testing.T) {
	t.Parallel()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test file path")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	workflow, err := os.ReadFile(filepath.Join(repoRoot, ".github", "workflows", "release.yml"))
	if err != nil {
		t.Fatalf("read release workflow: %v", err)
	}

	if !strings.Contains(string(workflow), "body_path: .namba/releases/${{ github.ref_name }}.md") {
		t.Fatalf("expected release workflow to publish notes body path, got:\n%s", workflow)
	}
	if strings.Count(string(workflow), "actions/checkout@v6") < 2 {
		t.Fatalf("expected release workflow publish job to checkout notes artifact before body_path, got:\n%s", workflow)
	}
}
