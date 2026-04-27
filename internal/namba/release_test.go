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

func TestReleaseNotesEnrichSingleSpecCommitFromAcceptanceDetails(t *testing.T) {
	t.Parallel()

	root := canonicalTempDir(t)
	writeTestFile(t, filepath.Join(root, ".namba", "specs", "SPEC-041", "acceptance.md"), "# Acceptance\n\n## `$namba-release`\n\n- [x] Release notes expand single squash commits with SPEC acceptance details.\n- [ ] Release notes must not claim unfinished acceptance work as shipped.\n- [x] Release notes preserve SPEC IDs and short commit hashes.\n- [x] Validation commands pass: `go test ./...` and `go vet ./...`.\n\n## `namba pr`\n\n- [x] PR handoff inspects GitHub checks before review request.\n- [x] PR handoff reports bounded failure snippets when checks fail.\n- [x] PR handoff confirms the configured review marker once.\n\n## Tests And Validation\n\n- [x] `go test ./...` passes.\n")

	commits := []releaseCommit{
		{
			ShortHash: "abcdef0",
			Subject:   "SPEC-041 릴리스 노트 상세화",
			Category:  releaseNoteCategoryDocs,
			Refs:      []string{"SPEC-041"},
		},
	}

	notes := renderReleaseNotes("v0.5.6", "v0.5.5", enrichReleaseCommitsWithSpecDetails(root, commits))
	for _, want := range []string{
		"- SPEC-041 릴리스 노트 상세화 (SPEC-041, abcdef0)",
		"  - `$namba-release`: Release notes expand single squash commits with SPEC acceptance details.",
		"  - `namba pr`: PR handoff inspects GitHub checks before review request.",
		"  - `namba pr`: PR handoff confirms the configured review marker once.",
	} {
		if !strings.Contains(notes, want) {
			t.Fatalf("release notes missing %q:\n%s", want, notes)
		}
	}
	for _, unwanted := range []string{"Validation commands pass", "go test ./... passes", "unfinished acceptance work"} {
		if strings.Contains(notes, unwanted) {
			t.Fatalf("release notes should omit validation-only detail %q:\n%s", unwanted, notes)
		}
	}
}

func TestReleaseNotesEnrichCommitBodyBullets(t *testing.T) {
	t.Parallel()

	commits := []releaseCommit{
		{
			ShortHash: "1234567",
			Subject:   "fix: release body handoff",
			Body:      "- Preserve the generated GitHub Release body.\n- Avoid publishing a generic one-line summary.\nPR #42\n",
			Category:  releaseNoteCategoryFixes,
			Refs:      []string{"PR #42"},
		},
	}

	notes := renderReleaseNotes("v0.5.6", "v0.5.5", enrichReleaseCommitsWithSpecDetails("", commits))
	for _, want := range []string{
		"- release body handoff (PR #42, 1234567)",
		"  - Preserve the generated GitHub Release body.",
		"  - Avoid publishing a generic one-line summary.",
	} {
		if !strings.Contains(notes, want) {
			t.Fatalf("release notes missing %q:\n%s", want, notes)
		}
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
