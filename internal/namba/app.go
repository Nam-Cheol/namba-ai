package namba

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const (
	nambaDir     = ".namba"
	specsDir     = ".namba/specs"
	projectDir   = ".namba/project"
	codemapsDir  = ".namba/project/codemaps"
	configDir    = ".namba/config/sections"
	logsDir      = ".namba/logs"
	worktreesDir = ".namba/worktrees"
	manifestPath = ".namba/manifest.json"

	manifestOwnerManaged = "namba-managed"

	noReviewPlanningFlag = "--no-review"
)

type App struct {
	stdin                   io.Reader
	stdout                  io.Writer
	stderr                  io.Writer
	now                     func() time.Time
	getenv                  func(string) string
	getwd                   func() (string, error)
	readFile                func(string) ([]byte, error)
	writeFile               func(string, []byte, fs.FileMode) error
	mkdirAll                func(string, fs.FileMode) error
	lookPath                func(string) (string, error)
	detectCodexCapabilities func(context.Context, string, executionRequest) (codexCapabilityMatrix, error)
	runCmd                  func(context.Context, string, []string, string) (string, error)
	runCmdWithInput         func(context.Context, string, []string, string, string) (string, string, error)
	runCodexCmdWithInput    func(context.Context, string, []string, string, string) (string, string, error)
	startCmd                func(string, []string, string) error
	downloadURL             func(context.Context, string) ([]byte, error)
	executablePath          func() (string, error)
	userCacheDir            func() (string, error)
	writeManifestOverride   func(string, Manifest) error
	newParallelProgressSink func(parallelProgressSinkConfig) (parallelProgressSink, error)
	goos                    string
	goarch                  string
}

type Manifest struct {
	GeneratedAt string          `json:"generated_at"`
	Entries     []ManifestEntry `json:"entries"`
}

type ManifestEntry struct {
	Path      string `json:"path"`
	Kind      string `json:"kind"`
	Owner     string `json:"owner,omitempty"`
	Checksum  string `json:"checksum"`
	UpdatedAt string `json:"updated_at"`
}

type projectConfig struct {
	Name        string
	ProjectType string
	Language    string
	Framework   string
}

type qualityConfig struct {
	DevelopmentMode        string
	TestCommand            string
	LintCommand            string
	TypecheckCommand       string
	BuildCommand           string
	MigrationDryRunCommand string
	SmokeStartCommand      string
	OutputContractCommand  string
}

type docsConfig struct {
	ManageReadme           bool
	ReadmeProfile          string
	DefaultLanguage        string
	AdditionalLanguages    []string
	AdditionalLanguagesSet bool
	HeroImage              string
}

type packageManifest struct {
	Name             string            `json:"name"`
	Dependencies     map[string]string `json:"dependencies"`
	DevDependencies  map[string]string `json:"devDependencies"`
	PeerDependencies map[string]string `json:"peerDependencies"`
}

type specPackage struct {
	ID          string
	Description string
	Path        string
}

func NewApp(stdout, stderr io.Writer) *App {
	return &App{
		stdin:        os.Stdin,
		stdout:       stdout,
		stderr:       stderr,
		now:          time.Now,
		getenv:       os.Getenv,
		getwd:        os.Getwd,
		readFile:     os.ReadFile,
		writeFile:    os.WriteFile,
		mkdirAll:     os.MkdirAll,
		lookPath:     exec.LookPath,
		userCacheDir: os.UserCacheDir,
		newParallelProgressSink: func(cfg parallelProgressSinkConfig) (parallelProgressSink, error) {
			return newJSONLParallelProgressSink(cfg)
		},
		executablePath: os.Executable,
		goos:           runtime.GOOS,
		goarch:         runtime.GOARCH,
		runCmd: func(ctx context.Context, name string, args []string, dir string) (string, error) {
			cmd := exec.CommandContext(ctx, name, args...)
			cmd.Dir = dir
			output, err := cmd.CombinedOutput()
			return strings.TrimSpace(string(output)), err
		},
		runCmdWithInput: func(ctx context.Context, name string, args []string, dir, input string) (string, string, error) {
			cmd := exec.CommandContext(ctx, name, args...)
			cmd.Dir = dir
			if input != "" {
				cmd.Stdin = strings.NewReader(input)
			}
			var stdout strings.Builder
			var stderr strings.Builder
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			err := cmd.Run()
			return stdout.String(), stderr.String(), err
		},
		runCodexCmdWithInput: func(ctx context.Context, name string, args []string, dir, input string) (string, string, error) {
			cmd := exec.CommandContext(ctx, name, args...)
			cmd.Dir = dir
			cmd.Stdin = strings.NewReader(input)
			var stdout strings.Builder
			var stderr strings.Builder
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			err := cmd.Run()
			return stdout.String(), stderr.String(), err
		},
		startCmd: func(name string, args []string, dir string) error {
			cmd := exec.Command(name, args...)
			cmd.Dir = dir
			return cmd.Start()
		},
		downloadURL: func(ctx context.Context, url string) ([]byte, error) {
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				return nil, err
			}
			req.Header.Set("User-Agent", "NambaAI-Updater")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
				return nil, fmt.Errorf("download %s failed with status %d: %s", url, resp.StatusCode, strings.TrimSpace(string(body)))
			}
			return io.ReadAll(resp.Body)
		},
	}
}
