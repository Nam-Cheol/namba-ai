package namba

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
)

func TestInitTUIModelSupportsNavigationConfirmBackAndQuit(t *testing.T) {
	t.Parallel()

	model := newInitTUIModel(testInitTUIProfile())
	model = updateInitTUIModel(t, model, tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}))
	if model.selected != 1 {
		t.Fatalf("down selected = %d, want 1", model.selected)
	}

	model = updateInitTUIModel(t, model, tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	if model.profile.ConversationLanguage != "en" {
		t.Fatalf("language = %q, want en", model.profile.ConversationLanguage)
	}
	if model.step != wizardStepProjectType {
		t.Fatalf("step = %d, want project type", model.step)
	}

	model = updateInitTUIModel(t, model, tea.KeyPressMsg(tea.Key{Code: tea.KeyLeft}))
	if model.step != wizardStepHumanLanguage {
		t.Fatalf("left should navigate back to language step, got %d", model.step)
	}

	model = updateInitTUIModel(t, model, tea.KeyPressMsg(tea.Key{Code: 'q', Text: "q"}))
	if !model.canceled {
		t.Fatal("q should cancel the init TUI model")
	}
}

func TestInitTUIModelHandlesTextInputAndAltScreenView(t *testing.T) {
	t.Parallel()

	model := newInitTUIModel(testInitTUIProfile())
	model.step = wizardStepProjectDetails
	model.resetInput()
	model = updateInitTUIModel(t, model, tea.KeyPressMsg(tea.Key{Code: 'x', Text: "x"}))
	model = updateInitTUIModel(t, model, tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))

	if model.profile.ProjectName != "demox" {
		t.Fatalf("project name = %q, want demox", model.profile.ProjectName)
	}
	if model.step != wizardStepDevelopmentMode {
		t.Fatalf("step = %d, want development mode", model.step)
	}

	view := model.View()
	if !view.AltScreen {
		t.Fatal("init TUI view must request AltScreen")
	}
}

func TestInitTUIModelAllowsQInTextInput(t *testing.T) {
	t.Parallel()

	model := newInitTUIModel(testInitTUIProfile())
	model.step = wizardStepProjectDetails
	model.resetInput()
	model = updateInitTUIModel(t, model, tea.KeyPressMsg(tea.Key{Code: 'q', Text: "q"}))

	if model.canceled {
		t.Fatal("q should be text input on text steps, not a quit shortcut")
	}
	if model.input != "demoq" {
		t.Fatalf("input = %q, want demoq", model.input)
	}
}

func TestInitTUIModelBackspaceRemovesLastRune(t *testing.T) {
	t.Parallel()

	model := newInitTUIModel(testInitTUIProfile())
	model.step = wizardStepProjectDetails
	model.resetInput()
	model.input = "프로젝트"
	model = updateInitTUIModel(t, model, tea.KeyPressMsg(tea.Key{Code: tea.KeyBackspace}))

	if model.input != "프로젝" {
		t.Fatalf("input = %q, want 프로젝", model.input)
	}
	if !utf8.ValidString(model.input) {
		t.Fatalf("input should remain valid UTF-8 after backspace: %q", model.input)
	}
}

func TestRunInitNonTTYFallbackPrintsCreatedAndSkippedSummary(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	stdout := &bytes.Buffer{}
	app := NewApp(stdout, &bytes.Buffer{})
	if err := app.Run(context.Background(), []string{"init", tmp, "--yes", "--name", "demo"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	first := stdout.String()
	if !strings.Contains(first, "Files:") || !strings.Contains(first, "created") {
		t.Fatalf("expected created file summary, got %q", first)
	}

	stdout.Reset()
	if err := app.Run(context.Background(), []string{"init", tmp, "--yes", "--name", "demo"}); err != nil {
		t.Fatalf("second init failed: %v", err)
	}
	second := stdout.String()
	if !strings.Contains(second, "skipped") {
		t.Fatalf("expected skipped file summary on identical rerun, got %q", second)
	}
}

func TestRunRegenNonTTYFallbackKeepsPlainResultText(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	stdout := &bytes.Buffer{}
	app := NewApp(stdout, &bytes.Buffer{})
	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	restore := chdirExecution(t, tmp)
	defer restore()
	stdout.Reset()
	if err := app.Run(context.Background(), []string{"regen"}); err != nil {
		t.Fatalf("regen failed: %v", err)
	}
	got := stdout.String()
	if !strings.Contains(got, "Regenerated NambaAI AGENTS") {
		t.Fatalf("expected plain regen fallback output, got %q", got)
	}
}

func TestRunUpdateNonTTYFallbackKeepsPlainProgressText(t *testing.T) {
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
	got := stdout.String()
	if !strings.Contains(got, "Downloading latest release using namba_Linux_x86_64.tar.gz for linux/amd64") {
		t.Fatalf("expected plain update fallback progress, got %q", got)
	}
}

func updateInitTUIModel(t *testing.T, model initTUIModel, msg tea.Msg) initTUIModel {
	t.Helper()

	updated, _ := model.Update(msg)
	next, ok := updated.(initTUIModel)
	if !ok {
		t.Fatalf("Update returned %T, want initTUIModel", updated)
	}
	return next
}

func testInitTUIProfile() initProfile {
	return initProfile{
		ProjectName:           "demo",
		ProjectType:           "existing",
		Language:              "go",
		Framework:             "cobra",
		DevelopmentMode:       "tdd",
		ConversationLanguage:  "ko",
		DocumentationLanguage: "ko",
		CommentLanguage:       "ko",
		ApprovalPolicy:        "on-request",
		SandboxMode:           "workspace-write",
		GitMode:               "manual",
		GitProvider:           "github",
		GitLabInstanceURL:     "https://gitlab.com",
		BranchPerWork:         true,
		BranchBase:            "main",
		SpecBranchPrefix:      "spec/",
		TaskBranchPrefix:      "task/",
		PRBaseBranch:          "main",
		PRLanguage:            "ko",
		CodexReviewComment:    "@codex review",
		AgentMode:             "single",
		StatusLinePreset:      "namba",
		UserName:              "Developer",
		CreatedAt:             "2026-04-23T10:00:00Z",
	}
}
