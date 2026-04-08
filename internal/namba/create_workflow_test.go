package namba

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunRegenPreservesUserAuthoredCreateArtifacts(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	userSkillPath := filepath.Join(tmp, ".agents", "skills", "user-authored-skill", "SKILL.md")
	userAgentTomlPath := filepath.Join(tmp, ".codex", "agents", "user-authored-agent.toml")
	userAgentMarkdownPath := filepath.Join(tmp, ".codex", "agents", "user-authored-agent.md")

	if err := os.MkdirAll(filepath.Dir(userSkillPath), 0o755); err != nil {
		t.Fatalf("mkdir skill dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(userAgentTomlPath), 0o755); err != nil {
		t.Fatalf("mkdir agent dir: %v", err)
	}
	writeTestFile(t, userSkillPath, "user-authored skill remains owned by the user\n")
	writeTestFile(t, userAgentTomlPath, "name = \"user-authored-agent\"\n")
	writeTestFile(t, userAgentMarkdownPath, "# User Authored Agent\n")

	manifest, err := app.readManifest(tmp)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	now := app.now().Format(time.RFC3339)
	manifest = upsertManifest(manifest, ManifestEntry{
		Path:      ".agents/skills/user-authored-skill/SKILL.md",
		Kind:      manifestKind(".agents/skills/user-authored-skill/SKILL.md"),
		Checksum:  checksum("user-authored skill remains owned by the user\n"),
		UpdatedAt: now,
	})
	manifest = upsertManifest(manifest, ManifestEntry{
		Path:      ".codex/agents/user-authored-agent.toml",
		Kind:      manifestKind(".codex/agents/user-authored-agent.toml"),
		Checksum:  checksum("name = \"user-authored-agent\"\n"),
		UpdatedAt: now,
	})
	manifest = upsertManifest(manifest, ManifestEntry{
		Path:      ".codex/agents/user-authored-agent.md",
		Kind:      manifestKind(".codex/agents/user-authored-agent.md"),
		Checksum:  checksum("# User Authored Agent\n"),
		UpdatedAt: now,
	})
	if err := app.writeManifest(tmp, manifest); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	restore := chdirExecution(t, tmp)
	defer restore()

	if err := app.Run(context.Background(), []string{"regen"}); err != nil {
		t.Fatalf("regen failed: %v", err)
	}

	if got := mustReadFile(t, userSkillPath); got != "user-authored skill remains owned by the user\n" {
		t.Fatalf("user-authored skill changed after regen: %q", got)
	}
	if got := mustReadFile(t, userAgentTomlPath); got != "name = \"user-authored-agent\"\n" {
		t.Fatalf("user-authored agent toml changed after regen: %q", got)
	}
	if got := mustReadFile(t, userAgentMarkdownPath); got != "# User Authored Agent\n" {
		t.Fatalf("user-authored agent markdown changed after regen: %q", got)
	}

	manifest, err = app.readManifest(tmp)
	if err != nil {
		t.Fatalf("read manifest after regen: %v", err)
	}
	for _, rel := range []string{
		".agents/skills/user-authored-skill/SKILL.md",
		".codex/agents/user-authored-agent.toml",
		".codex/agents/user-authored-agent.md",
	} {
		entry, ok := findManifestEntry(manifest, rel)
		if !ok {
			t.Fatalf("expected manifest entry for %s to remain after regen", rel)
		}
		if entry.Owner != "" {
			t.Fatalf("expected user-authored manifest entry for %s to stay unowned, got %q", rel, entry.Owner)
		}
	}
}

func TestRunRegenRemovesStaleManagedCreateArtifacts(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	staleSkillPath := filepath.Join(tmp, ".agents", "skills", "namba-obsolete", "SKILL.md")
	staleAgentTomlPath := filepath.Join(tmp, ".codex", "agents", "namba-obsolete.toml")
	staleAgentMarkdownPath := filepath.Join(tmp, ".codex", "agents", "namba-obsolete.md")
	if err := os.MkdirAll(filepath.Dir(staleSkillPath), 0o755); err != nil {
		t.Fatalf("mkdir stale skill dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(staleAgentTomlPath), 0o755); err != nil {
		t.Fatalf("mkdir stale agent dir: %v", err)
	}
	writeTestFile(t, staleSkillPath, "stale managed skill\n")
	writeTestFile(t, staleAgentTomlPath, "name = \"namba-obsolete\"\n")
	writeTestFile(t, staleAgentMarkdownPath, "# Namba Obsolete\n")

	manifest, err := app.readManifest(tmp)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	now := app.now().Format(time.RFC3339)
	for _, entry := range []ManifestEntry{
		{
			Path:      ".agents/skills/namba-obsolete/SKILL.md",
			Kind:      manifestKind(".agents/skills/namba-obsolete/SKILL.md"),
			Owner:     manifestOwnerManaged,
			Checksum:  checksum("stale managed skill\n"),
			UpdatedAt: now,
		},
		{
			Path:      ".codex/agents/namba-obsolete.toml",
			Kind:      manifestKind(".codex/agents/namba-obsolete.toml"),
			Owner:     manifestOwnerManaged,
			Checksum:  checksum("name = \"namba-obsolete\"\n"),
			UpdatedAt: now,
		},
		{
			Path:      ".codex/agents/namba-obsolete.md",
			Kind:      manifestKind(".codex/agents/namba-obsolete.md"),
			Owner:     manifestOwnerManaged,
			Checksum:  checksum("# Namba Obsolete\n"),
			UpdatedAt: now,
		},
	} {
		manifest = upsertManifest(manifest, entry)
	}
	if err := app.writeManifest(tmp, manifest); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	restore := chdirExecution(t, tmp)
	defer restore()

	if err := app.Run(context.Background(), []string{"regen"}); err != nil {
		t.Fatalf("regen failed: %v", err)
	}

	for _, path := range []string{staleSkillPath, staleAgentTomlPath, staleAgentMarkdownPath} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("expected stale managed artifact %s to be removed, stat err=%v", path, err)
		}
	}

	manifest, err = app.readManifest(tmp)
	if err != nil {
		t.Fatalf("read manifest after regen: %v", err)
	}
	for _, rel := range []string{
		".agents/skills/namba-obsolete/SKILL.md",
		".codex/agents/namba-obsolete.toml",
		".codex/agents/namba-obsolete.md",
	} {
		if _, ok := findManifestEntry(manifest, rel); ok {
			t.Fatalf("expected stale managed manifest entry for %s to be removed", rel)
		}
	}
}

func TestCreateWorkflowSkillContractFixture(t *testing.T) {
	t.Parallel()

	contract := renderCreateCommandSkill()
	wantFragments := []string{
		"namba-create",
		"staged generator",
		"`unresolved` -> `narrowed` -> `confirmed`",
		"Do not write files until the target, slug, paths, and overwrite decisions are explicit",
		"Before any write, present a non-mutating preview",
		"skill`, `agent`, or `both`",
		"validation plan",
		"fresh Codex session will likely be required",
		"sequential-thinking",
		"context7",
		"playwright",
		"five independent role outputs",
		".agents/skills/<slug>/SKILL.md",
		".codex/agents/<slug>.toml",
		".codex/agents/<slug>.md",
	}

	for _, want := range wantFragments {
		if !strings.Contains(contract, want) {
			t.Fatalf("expected create workflow contract to contain %q, got %q", want, contract)
		}
	}
}
