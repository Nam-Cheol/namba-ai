package namba

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
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
		"namba __create preview",
		"namba __create apply",
		"`unresolved` -> `narrowed` -> `confirmed`",
		"Do not write files until the target, slug, paths, and overwrite decisions are explicit",
		"Before any write, present a non-mutating preview",
		"public `namba create` Go CLI command",
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

func TestInternalCreateAdapterIsHiddenFromPublicHelp(t *testing.T) {
	t.Parallel()

	if strings.Contains(usageText(), internalCreateCommandName) {
		t.Fatalf("expected public usage to hide %s, got %q", internalCreateCommandName, usageText())
	}
	if _, ok := commandUsageText(internalCreateCommandName); ok {
		t.Fatalf("expected public command usage to hide %s", internalCreateCommandName)
	}

	stdout := &bytes.Buffer{}
	app := NewApp(stdout, &bytes.Buffer{})
	err := app.Run(context.Background(), []string{"help", internalCreateCommandName})
	if err == nil || !strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("expected unknown command for hidden help topic, got %v", err)
	}
}

func TestInternalCreateAdapterPreviewJSONRoundTrip(t *testing.T) {
	t.Parallel()

	tmp, app := prepareCreateProject(t)
	restore := chdirExecution(t, tmp)
	defer restore()

	stdout := &bytes.Buffer{}
	app.stdout = stdout
	app.stdin = strings.NewReader(mustMarshalJSON(t, createRequest{
		Target:       createTargetBoth,
		Name:         "Insight Builder",
		Description:  "Create repo-local creation helpers.",
		Instructions: "Use this artifact to keep creation flows explicit and safe.",
	}))

	if err := app.Run(context.Background(), []string{internalCreateCommandName, "preview"}); err != nil {
		t.Fatalf("internal preview failed: %v", err)
	}

	var preview createPreview
	if err := json.Unmarshal(stdout.Bytes(), &preview); err != nil {
		t.Fatalf("unmarshal preview json: %v", err)
	}

	if preview.Target != createTargetBoth || preview.Slug != "insight-builder" {
		t.Fatalf("unexpected preview: %+v", preview)
	}
	if strings.TrimSpace(preview.PreviewDigest) == "" {
		t.Fatalf("expected preview digest, got %+v", preview)
	}
	for _, rel := range preview.ExactOutputPaths {
		if _, err := os.Stat(filepath.Join(tmp, filepath.FromSlash(rel))); !os.IsNotExist(err) {
			t.Fatalf("preview adapter should not create %s, stat err=%v", rel, err)
		}
	}
}

func TestInternalCreateAdapterApplyJSONRoundTrip(t *testing.T) {
	t.Parallel()

	tmp, app := prepareCreateProject(t)
	restore := chdirExecution(t, tmp)
	defer restore()

	previewStdout := &bytes.Buffer{}
	app.stdout = previewStdout
	app.stdin = strings.NewReader(mustMarshalJSON(t, createRequest{
		Target:       createTargetSkill,
		Name:         "Insight Builder",
		Description:  "Create repo-local creation helpers.",
		Instructions: "Use this artifact to keep creation flows explicit and safe.",
	}))

	if err := app.Run(context.Background(), []string{internalCreateCommandName, "preview"}); err != nil {
		t.Fatalf("internal preview failed: %v", err)
	}

	var preview createPreview
	if err := json.Unmarshal(previewStdout.Bytes(), &preview); err != nil {
		t.Fatalf("unmarshal preview json: %v", err)
	}

	applyStdout := &bytes.Buffer{}
	app.stdout = applyStdout
	app.stdin = strings.NewReader(mustMarshalJSON(t, createRequest{
		Target:        createTargetSkill,
		Name:          "Insight Builder",
		Description:   "Create repo-local creation helpers.",
		Instructions:  "Use this artifact to keep creation flows explicit and safe.",
		PreviewDigest: preview.PreviewDigest,
		Confirmed:     true,
	}))

	if err := app.Run(context.Background(), []string{internalCreateCommandName, "apply"}); err != nil {
		t.Fatalf("internal apply failed: %v", err)
	}

	var result createApplyResult
	if err := json.Unmarshal(applyStdout.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal apply json: %v", err)
	}

	if strings.Join(result.WrittenPaths, "\n") != ".agents/skills/insight-builder/SKILL.md" {
		t.Fatalf("unexpected written paths: %+v", result.WrittenPaths)
	}
	skillPath := filepath.Join(tmp, ".agents", "skills", "insight-builder", "SKILL.md")
	if got := mustReadFile(t, skillPath); !strings.Contains(got, "name: insight-builder") {
		t.Fatalf("expected adapter apply to write skill, got %q", got)
	}
}

func TestPreviewCreateNormalizesExactPathsAndDoesNotWrite(t *testing.T) {
	t.Parallel()

	tmp, app := prepareCreateProject(t)

	manifestBefore, err := app.readManifest(tmp)
	if err != nil {
		t.Fatalf("read manifest before preview: %v", err)
	}

	preview, err := app.previewCreate(tmp, createRequest{
		Target:       createTargetBoth,
		Name:         "Insight Builder",
		Description:  "Create repo-local creation helpers.",
		Instructions: "Use this artifact to keep creation flows explicit and safe.",
	})
	if err != nil {
		t.Fatalf("preview create: %v", err)
	}

	if preview.Target != createTargetBoth {
		t.Fatalf("expected both target, got %+v", preview)
	}
	if preview.Name != "Insight Builder" || preview.Slug != "insight-builder" {
		t.Fatalf("expected normalized preview identity, got %+v", preview)
	}
	wantPaths := []string{
		".agents/skills/insight-builder/SKILL.md",
		".codex/agents/insight-builder.toml",
		".codex/agents/insight-builder.md",
	}
	if strings.Join(preview.ExactOutputPaths, "\n") != strings.Join(wantPaths, "\n") {
		t.Fatalf("unexpected preview paths: got %v want %v", preview.ExactOutputPaths, wantPaths)
	}
	if preview.OverwriteImpact.Required {
		t.Fatalf("expected preview without overwrite requirement, got %+v", preview)
	}
	if strings.TrimSpace(preview.PreviewDigest) == "" {
		t.Fatalf("expected preview digest, got %+v", preview)
	}
	if !preview.SessionRefresh.Required || preview.SessionRefresh.Reason == "" {
		t.Fatalf("expected preview to require session refresh guidance, got %+v", preview)
	}
	for _, want := range []string{"go test ./...", `gofmt -l "cmd" "internal" "namba_test.go"`, "go vet ./..."} {
		if !strings.Contains(strings.Join(preview.ValidationPlan, "\n"), want) {
			t.Fatalf("expected validation plan to include %q, got %+v", want, preview.ValidationPlan)
		}
	}

	manifestAfter, err := app.readManifest(tmp)
	if err != nil {
		t.Fatalf("read manifest after preview: %v", err)
	}
	if len(manifestAfter.Entries) != len(manifestBefore.Entries) {
		t.Fatalf("preview should not mutate manifest: before=%d after=%d", len(manifestBefore.Entries), len(manifestAfter.Entries))
	}
	for _, rel := range wantPaths {
		if _, err := os.Stat(filepath.Join(tmp, filepath.FromSlash(rel))); !os.IsNotExist(err) {
			t.Fatalf("preview should not create %s, stat err=%v", rel, err)
		}
	}
}

func TestPreviewCreateReportsOverwriteImpact(t *testing.T) {
	t.Parallel()

	tmp, app := prepareCreateProject(t)

	existing := filepath.Join(tmp, ".agents", "skills", "insight-builder", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(existing), 0o755); err != nil {
		t.Fatalf("mkdir existing skill dir: %v", err)
	}
	writeTestFile(t, existing, "existing skill\n")

	preview, err := app.previewCreate(tmp, createRequest{
		Target:       createTargetSkill,
		Name:         "Insight Builder",
		Description:  "Create repo-local creation helpers.",
		Instructions: "Use this artifact to keep creation flows explicit and safe.",
	})
	if err != nil {
		t.Fatalf("preview create: %v", err)
	}

	if !preview.OverwriteImpact.Required {
		t.Fatalf("expected overwrite requirement, got %+v", preview)
	}
	if strings.Join(preview.OverwriteImpact.Paths, "\n") != ".agents/skills/insight-builder/SKILL.md" {
		t.Fatalf("unexpected overwrite paths: %+v", preview.OverwriteImpact.Paths)
	}
}

func TestPreviewCreateDigestChangesWhenPlanChanges(t *testing.T) {
	t.Parallel()

	tmp, app := prepareCreateProject(t)

	baseReq := createRequest{
		Target:       createTargetSkill,
		Name:         "Insight Builder",
		Description:  "Create repo-local creation helpers.",
		Instructions: "Use this artifact to keep creation flows explicit and safe.",
	}
	basePreview, err := app.previewCreate(tmp, baseReq)
	if err != nil {
		t.Fatalf("preview create: %v", err)
	}

	changedPreview, err := app.previewCreate(tmp, createRequest{
		Target:       createTargetSkill,
		Name:         "Insight Builder",
		Description:  "Create repo-local creation helpers.",
		Instructions: "Use this artifact to keep creation flows explicit, safe, and deterministic.",
	})
	if err != nil {
		t.Fatalf("preview create with changed instructions: %v", err)
	}

	if basePreview.PreviewDigest == changedPreview.PreviewDigest {
		t.Fatalf("expected preview digest to change when durable plan changes: base=%q changed=%q", basePreview.PreviewDigest, changedPreview.PreviewDigest)
	}
}

func TestPreviewCreateRejectsUnsafeInstructions(t *testing.T) {
	t.Parallel()

	tmp, app := prepareCreateProject(t)

	tests := []struct {
		name         string
		instructions string
		wantErr      string
	}{
		{
			name:         "claude path",
			instructions: "Write the durable workflow under .claude/commands/create.md.",
			wantErr:      "Claude-only paths",
		},
		{
			name:         "claude runtime primitive",
			instructions: "Use TeamCreate and SendMessage to coordinate the durable workflow.",
			wantErr:      "Claude-only runtime primitives",
		},
		{
			name:         "legacy codex skills mirror",
			instructions: "Write the generated output into .codex/skills/insight-builder.",
			wantErr:      "Deprecated .codex/skills mirror",
		},
		{
			name:         "legacy codex skills mirror windows path",
			instructions: `Write the generated output into .codex\skills\insight-builder.`,
			wantErr:      "Deprecated .codex/skills mirror",
		},
		{
			name:         "opus requirement",
			instructions: "Require model: \"opus\" for every generated run.",
			wantErr:      "Claude-only model requirements",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := app.previewCreate(tmp, createRequest{
				Target:       createTargetSkill,
				Name:         "Insight Builder",
				Description:  "Create repo-local creation helpers.",
				Instructions: tc.instructions,
			})
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestPreviewCreateRejectsManagedTargets(t *testing.T) {
	t.Parallel()

	tmp, app := prepareCreateProject(t)

	for _, tc := range []struct {
		name    string
		req     createRequest
		wantErr string
	}{
		{
			name: "managed skill",
			req: createRequest{
				Target:       createTargetSkill,
				Name:         "namba-run",
				Description:  "Invalid",
				Instructions: "Invalid",
			},
			wantErr: "reserved for Namba-managed skills",
		},
		{
			name: "managed agent",
			req: createRequest{
				Target:       createTargetAgent,
				Name:         "namba-reviewer",
				Description:  "Invalid",
				Instructions: "Invalid",
			},
			wantErr: "reserved for Namba-managed agents",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := app.previewCreate(tmp, tc.req); err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestApplyCreateRequiresConfirmation(t *testing.T) {
	t.Parallel()

	tmp, app := prepareCreateProject(t)

	if _, err := app.applyCreate(tmp, createRequest{
		Target:       createTargetSkill,
		Name:         "Insight Builder",
		Description:  "Create repo-local creation helpers.",
		Instructions: "Use this artifact to keep creation flows explicit and safe.",
	}); err == nil || !strings.Contains(err.Error(), "confirmation") {
		t.Fatalf("expected confirmation error, got %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmp, ".agents", "skills", "insight-builder", "SKILL.md")); !os.IsNotExist(err) {
		t.Fatalf("apply without confirmation should not create output, stat err=%v", err)
	}
}

func TestApplyCreateWritesSkillOnlyAndTracksUserOwnership(t *testing.T) {
	t.Parallel()

	tmp, app := prepareCreateProject(t)

	req := createRequest{
		Target:       createTargetSkill,
		Name:         "Insight Builder",
		Description:  "Create repo-local creation helpers.",
		Instructions: "Use this artifact to keep creation flows explicit and safe.",
	}
	req.PreviewDigest = mustPreviewCreateDigest(t, app, tmp, req)

	result, err := app.applyCreate(tmp, createRequest{
		Target:        req.Target,
		Name:          req.Name,
		Description:   req.Description,
		Instructions:  req.Instructions,
		PreviewDigest: req.PreviewDigest,
		Confirmed:     true,
	})
	if err != nil {
		t.Fatalf("apply create: %v", err)
	}

	skillPath := filepath.Join(tmp, ".agents", "skills", "insight-builder", "SKILL.md")
	if got := mustReadFile(t, skillPath); !strings.Contains(got, "name: insight-builder") || !strings.Contains(got, "Create repo-local creation helpers.") || !strings.Contains(got, "Use this artifact to keep creation flows explicit and safe.") {
		t.Fatalf("unexpected skill content: %q", got)
	}
	for _, rel := range []string{".codex/agents/insight-builder.toml", ".codex/agents/insight-builder.md"} {
		if _, err := os.Stat(filepath.Join(tmp, filepath.FromSlash(rel))); !os.IsNotExist(err) {
			t.Fatalf("skill-only create should not write %s, stat err=%v", rel, err)
		}
	}
	if strings.Join(result.WrittenPaths, "\n") != ".agents/skills/insight-builder/SKILL.md" {
		t.Fatalf("unexpected written paths: %+v", result.WrittenPaths)
	}

	manifest, err := app.readManifest(tmp)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	entry, ok := findManifestEntry(manifest, ".agents/skills/insight-builder/SKILL.md")
	if !ok {
		t.Fatalf("expected manifest entry for created skill")
	}
	if entry.Owner != "" {
		t.Fatalf("expected user-authored ownership, got %+v", entry)
	}
	if notice := mustReadFile(t, filepath.Join(tmp, ".namba", "logs", "session-refresh-required.json")); !strings.Contains(notice, ".agents/skills/insight-builder/SKILL.md") {
		t.Fatalf("expected session refresh notice to mention created skill, got %q", notice)
	}
}

func TestApplyCreateWritesAgentPair(t *testing.T) {
	t.Parallel()

	tmp, app := prepareCreateProject(t)

	req := createRequest{
		Target:               createTargetAgent,
		Name:                 "Insight Builder",
		Description:          "Review and generate repo-local creation artifacts.",
		Instructions:         "Change only the files assigned by the main session.",
		SandboxMode:          "read-only",
		Model:                "gpt-5.4",
		ModelReasoningEffort: "high",
	}
	req.PreviewDigest = mustPreviewCreateDigest(t, app, tmp, req)

	result, err := app.applyCreate(tmp, createRequest{
		Target:               req.Target,
		Name:                 req.Name,
		Description:          req.Description,
		Instructions:         req.Instructions,
		SandboxMode:          req.SandboxMode,
		Model:                req.Model,
		ModelReasoningEffort: req.ModelReasoningEffort,
		PreviewDigest:        req.PreviewDigest,
		Confirmed:            true,
	})
	if err != nil {
		t.Fatalf("apply create: %v", err)
	}

	tomlPath := filepath.Join(tmp, ".codex", "agents", "insight-builder.toml")
	markdownPath := filepath.Join(tmp, ".codex", "agents", "insight-builder.md")
	if got := mustReadFile(t, tomlPath); !strings.Contains(got, `name = "insight-builder"`) || !strings.Contains(got, `sandbox_mode = "read-only"`) || !strings.Contains(got, `model = "gpt-5.4"`) || !strings.Contains(got, `model_reasoning_effort = "high"`) {
		t.Fatalf("unexpected agent toml content: %q", got)
	}
	if got := mustReadFile(t, markdownPath); !strings.Contains(got, "# Insight Builder") || !strings.Contains(got, "Change only the files assigned by the main session.") {
		t.Fatalf("unexpected agent markdown content: %q", got)
	}
	if _, err := os.Stat(filepath.Join(tmp, ".agents", "skills", "insight-builder", "SKILL.md")); !os.IsNotExist(err) {
		t.Fatalf("agent-only create should not write skill output, stat err=%v", err)
	}
	wantPaths := []string{".codex/agents/insight-builder.toml", ".codex/agents/insight-builder.md"}
	if strings.Join(result.WrittenPaths, "\n") != strings.Join(wantPaths, "\n") {
		t.Fatalf("unexpected written paths: %+v", result.WrittenPaths)
	}
}

func TestRenderUserAgentTOMLEscapesUserFields(t *testing.T) {
	t.Parallel()

	description := "He said \"hello\"\nSecond line."
	sandboxMode := "workspace-write\"\nunsafe"
	model := "gpt-5.4\"\nmini"
	reasoning := "high\"\ncareful"
	instructions := "Use C:\\work\\repo during dry-runs.\nKeep \"\"\" markers literal."

	got := renderUserAgentTOML(
		"insight-builder",
		description,
		instructions,
		sandboxMode,
		model,
		reasoning,
	)

	for key, value := range map[string]string{
		"description":            description,
		"sandbox_mode":           sandboxMode,
		"model":                  model,
		"model_reasoning_effort": reasoning,
	} {
		want := key + " = " + strconv.Quote(value)
		if !strings.Contains(got, want) {
			t.Fatalf("expected escaped %s line %q in %q", key, want, got)
		}
	}

	wantInstructions := strings.Join(splitCreateInstructions(withCreateAgentPreamble("insight-builder", instructions)), "\n")
	wantInstructionLine := "developer_instructions = " + strconv.Quote(wantInstructions)
	if !strings.Contains(got, wantInstructionLine) {
		t.Fatalf("expected escaped developer instructions line %q in %q", wantInstructionLine, got)
	}
	if strings.Contains(got, `developer_instructions = """`) {
		t.Fatalf("expected developer instructions to avoid multiline basic TOML strings, got %q", got)
	}
}

func TestPreviewCreateRejectsTargetsResolvingOutsideRootViaSymlink(t *testing.T) {
	t.Parallel()

	tmp, app := prepareCreateProject(t)

	externalDir := t.TempDir()
	mustCreateSymlink(t, externalDir, filepath.Join(tmp, ".agents", "skills", "insight-builder"))

	_, err := app.previewCreate(tmp, createRequest{
		Target:       createTargetSkill,
		Name:         "Insight Builder",
		Description:  "Create repo-local creation helpers.",
		Instructions: "Use this artifact to keep creation flows explicit and safe.",
	})
	if err == nil || !strings.Contains(err.Error(), "resolves outside the project root") {
		t.Fatalf("expected symlink root escape error, got %v", err)
	}
}

func TestApplyCreateRejectsSymlinkedAgentTargetAfterPreview(t *testing.T) {
	t.Parallel()

	tmp, app := prepareCreateProject(t)

	req := createRequest{
		Target:       createTargetAgent,
		Name:         "Insight Builder",
		Description:  "Create repo-local creation helpers.",
		Instructions: "Use this artifact to keep creation flows explicit and safe.",
	}
	req.PreviewDigest = mustPreviewCreateDigest(t, app, tmp, req)

	externalFile := filepath.Join(t.TempDir(), "outside-agent.toml")
	writeTestFile(t, externalFile, "outside\n")
	mustCreateSymlink(t, externalFile, filepath.Join(tmp, ".codex", "agents", "insight-builder.toml"))

	_, err := app.applyCreate(tmp, createRequest{
		Target:        req.Target,
		Name:          req.Name,
		Description:   req.Description,
		Instructions:  req.Instructions,
		PreviewDigest: req.PreviewDigest,
		Confirmed:     true,
	})
	if err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("expected symlink rejection, got %v", err)
	}
}

func TestApplyCreateWritesBothAndSurvivesRegen(t *testing.T) {
	t.Parallel()

	tmp, app := prepareCreateProject(t)

	req := createRequest{
		Target:       createTargetBoth,
		Name:         "Insight Builder",
		Description:  "Create repo-local creation helpers.",
		Instructions: "Use this artifact to keep creation flows explicit and safe.",
	}
	req.PreviewDigest = mustPreviewCreateDigest(t, app, tmp, req)

	if _, err := app.applyCreate(tmp, createRequest{
		Target:        req.Target,
		Name:          req.Name,
		Description:   req.Description,
		Instructions:  req.Instructions,
		PreviewDigest: req.PreviewDigest,
		Confirmed:     true,
	}); err != nil {
		t.Fatalf("apply create: %v", err)
	}

	restore := chdirExecution(t, tmp)
	defer restore()

	if err := app.Run(context.Background(), []string{"regen"}); err != nil {
		t.Fatalf("regen failed: %v", err)
	}

	for _, rel := range []string{
		".agents/skills/insight-builder/SKILL.md",
		".codex/agents/insight-builder.toml",
		".codex/agents/insight-builder.md",
	} {
		if _, err := os.Stat(filepath.Join(tmp, filepath.FromSlash(rel))); err != nil {
			t.Fatalf("expected create output %s to survive regen, stat err=%v", rel, err)
		}
		entry, ok := findManifestEntry(mustReadManifest(t, app, tmp), rel)
		if !ok {
			t.Fatalf("expected manifest entry for %s after regen", rel)
		}
		if entry.Owner != "" {
			t.Fatalf("expected %s to remain user-authored after regen, got %+v", rel, entry)
		}
	}
}

func TestApplyCreateRejectsUnsafeInputsAndPartialState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		prepare func(t *testing.T, root string)
		req     createRequest
		digest  bool
		wantErr string
	}{
		{
			name: "path traversal",
			req: createRequest{
				Target:       createTargetSkill,
				Name:         "../escape",
				Description:  "Invalid",
				Instructions: "Invalid",
			},
			wantErr: "path traversal",
		},
		{
			name: "invalid slug",
			req: createRequest{
				Target:       createTargetSkill,
				Name:         "!!!",
				Description:  "Invalid",
				Instructions: "Invalid",
			},
			wantErr: "invalid slug",
		},
		{
			name: "incomplete agent mirror",
			prepare: func(t *testing.T, root string) {
				if err := os.MkdirAll(filepath.Join(root, ".codex", "agents"), 0o755); err != nil {
					t.Fatalf("mkdir agent dir: %v", err)
				}
				writeTestFile(t, filepath.Join(root, ".codex", "agents", "insight-builder.toml"), `name = "insight-builder"`+"\n")
			},
			req: createRequest{
				Target:       createTargetAgent,
				Name:         "Insight Builder",
				Description:  "Invalid",
				Instructions: "Invalid",
			},
			wantErr: "incomplete agent mirror",
		},
		{
			name: "overwrite refusal",
			prepare: func(t *testing.T, root string) {
				if err := os.MkdirAll(filepath.Join(root, ".agents", "skills", "insight-builder"), 0o755); err != nil {
					t.Fatalf("mkdir skill dir: %v", err)
				}
				writeTestFile(t, filepath.Join(root, ".agents", "skills", "insight-builder", "SKILL.md"), "existing\n")
			},
			req: createRequest{
				Target:       createTargetSkill,
				Name:         "Insight Builder",
				Description:  "Invalid",
				Instructions: "Invalid",
				Confirmed:    true,
			},
			digest:  true,
			wantErr: "overwrite",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp, app := prepareCreateProject(t)
			if tt.prepare != nil {
				tt.prepare(t, tmp)
			}
			if tt.digest {
				tt.req.PreviewDigest = mustPreviewCreateDigest(t, app, tmp, createRequest{
					Target:               tt.req.Target,
					Name:                 tt.req.Name,
					Description:          tt.req.Description,
					Instructions:         tt.req.Instructions,
					SandboxMode:          tt.req.SandboxMode,
					Model:                tt.req.Model,
					ModelReasoningEffort: tt.req.ModelReasoningEffort,
				})
			}

			if _, err := app.applyCreate(tmp, tt.req); err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestApplyCreateRejectsPreviewDigestMismatch(t *testing.T) {
	t.Parallel()

	tmp, app := prepareCreateProject(t)

	req := createRequest{
		Target:       createTargetSkill,
		Name:         "Insight Builder",
		Description:  "Create repo-local creation helpers.",
		Instructions: "Use this artifact to keep creation flows explicit and safe.",
	}
	previewDigest := mustPreviewCreateDigest(t, app, tmp, req)

	_, err := app.applyCreate(tmp, createRequest{
		Target:        req.Target,
		Name:          req.Name,
		Description:   req.Description,
		Instructions:  "Use this artifact to keep creation flows explicit, safe, and deterministic.",
		PreviewDigest: previewDigest,
		Confirmed:     true,
	})
	if err == nil || !strings.Contains(err.Error(), "preview digest mismatch") {
		t.Fatalf("expected preview digest mismatch, got %v", err)
	}
}

func TestApplyCreateRollsBackOnPairedWriteFailure(t *testing.T) {
	t.Parallel()

	tmp, app := prepareCreateProject(t)

	failPath := filepath.Join(tmp, ".codex", "agents", "insight-builder.md")
	failed := false
	originalWriteFile := app.writeFile
	app.writeFile = func(path string, data []byte, perm os.FileMode) error {
		if !failed && path == failPath {
			failed = true
			return errors.New("forced paired write failure")
		}
		return originalWriteFile(path, data, perm)
	}

	req := createRequest{
		Target:       createTargetAgent,
		Name:         "Insight Builder",
		Description:  "Create repo-local creation helpers.",
		Instructions: "Use this artifact to keep creation flows explicit and safe.",
	}
	req.PreviewDigest = mustPreviewCreateDigest(t, app, tmp, req)

	if _, err := app.applyCreate(tmp, createRequest{
		Target:        req.Target,
		Name:          req.Name,
		Description:   req.Description,
		Instructions:  req.Instructions,
		PreviewDigest: req.PreviewDigest,
		Confirmed:     true,
	}); err == nil || !strings.Contains(err.Error(), "forced paired write failure") {
		t.Fatalf("expected paired write failure, got %v", err)
	}

	for _, rel := range []string{".codex/agents/insight-builder.toml", ".codex/agents/insight-builder.md"} {
		if _, err := os.Stat(filepath.Join(tmp, filepath.FromSlash(rel))); !os.IsNotExist(err) {
			t.Fatalf("expected rollback to remove %s, stat err=%v", rel, err)
		}
	}
	if _, ok := findManifestEntry(mustReadManifest(t, app, tmp), ".codex/agents/insight-builder.toml"); ok {
		t.Fatalf("expected rollback to remove manifest entry for paired agent write")
	}
}

func TestApplyCreateRollsBackOnManifestPersistenceFailure(t *testing.T) {
	t.Parallel()

	tmp, app := prepareCreateProject(t)

	failPath := filepath.Join(tmp, manifestPath)
	failed := false
	originalWriteFile := app.writeFile
	app.writeFile = func(path string, data []byte, perm os.FileMode) error {
		if !failed && path == failPath {
			failed = true
			return errors.New("forced manifest failure")
		}
		return originalWriteFile(path, data, perm)
	}

	req := createRequest{
		Target:       createTargetBoth,
		Name:         "Insight Builder",
		Description:  "Create repo-local creation helpers.",
		Instructions: "Use this artifact to keep creation flows explicit and safe.",
	}
	req.PreviewDigest = mustPreviewCreateDigest(t, app, tmp, req)

	if _, err := app.applyCreate(tmp, createRequest{
		Target:        req.Target,
		Name:          req.Name,
		Description:   req.Description,
		Instructions:  req.Instructions,
		PreviewDigest: req.PreviewDigest,
		Confirmed:     true,
	}); err == nil || !strings.Contains(err.Error(), "forced manifest failure") {
		t.Fatalf("expected manifest persistence failure, got %v", err)
	}

	for _, rel := range []string{
		".agents/skills/insight-builder/SKILL.md",
		".codex/agents/insight-builder.toml",
		".codex/agents/insight-builder.md",
	} {
		if _, err := os.Stat(filepath.Join(tmp, filepath.FromSlash(rel))); !os.IsNotExist(err) {
			t.Fatalf("expected rollback to remove %s after manifest failure, stat err=%v", rel, err)
		}
		if _, ok := findManifestEntry(mustReadManifest(t, app, tmp), rel); ok {
			t.Fatalf("expected rollback to remove manifest entry for %s", rel)
		}
	}
}

func prepareCreateProject(t *testing.T) (string, *App) {
	t.Helper()

	tmp := t.TempDir()
	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	writeTestFile(t, filepath.Join(tmp, ".namba", "config", "sections", "quality.yaml"), "development_mode: tdd\ntest_command: go test ./...\nlint_command: gofmt -l \"cmd\" \"internal\" \"namba_test.go\"\ntypecheck_command: go vet ./...\nbuild_command: none\nmigration_dry_run_command: none\nsmoke_start_command: none\noutput_contract_command: none\n")
	return tmp, app
}

func mustReadManifest(t *testing.T, app *App, root string) Manifest {
	t.Helper()

	manifest, err := app.readManifest(root)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	return manifest
}

func mustPreviewCreateDigest(t *testing.T, app *App, root string, req createRequest) string {
	t.Helper()

	preview, err := app.previewCreate(root, req)
	if err != nil {
		t.Fatalf("preview create for digest: %v", err)
	}
	if strings.TrimSpace(preview.PreviewDigest) == "" {
		t.Fatalf("expected preview digest, got %+v", preview)
	}
	return preview.PreviewDigest
}

func mustCreateSymlink(t *testing.T, target, link string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(link), err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink not supported: %v", err)
	}
}
