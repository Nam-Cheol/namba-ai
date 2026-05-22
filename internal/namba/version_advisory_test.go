package namba

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestEvaluateVersionAdvisoryStates(t *testing.T) {
	cases := []struct {
		name      string
		installed string
		latest    string
		want      string
	}{
		{name: "behind", installed: "v0.5.0", latest: "v0.6.0", want: "behind"},
		{name: "current", installed: "v0.6.0", latest: "v0.6.0", want: "current"},
		{name: "newer local", installed: "v0.7.0", latest: "v0.6.0", want: "current"},
		{name: "dev", installed: "dev", latest: "v0.6.0", want: "dev"},
		{name: "blank", installed: "", latest: "v0.6.0", want: "dev"},
		{name: "annotated local", installed: "v0.6.0-local", latest: "v0.7.0", want: "dev"},
		{name: "bad latest", installed: "v0.6.0", latest: "latest", want: "parse_failed"},
	}
	for _, tc := range cases {
		got := evaluateVersionAdvisory(tc.installed, tc.latest)
		if got.Status != tc.want {
			t.Fatalf("%s: status=%q want %q", tc.name, got.Status, tc.want)
		}
	}
}

func TestParseLatestReleaseVersion(t *testing.T) {
	if got, err := parseLatestReleaseVersion([]byte(`{"tag_name":"v0.6.1","prerelease":false}`)); err != nil || got != "v0.6.1" {
		t.Fatalf("parse valid release = %q, %v", got, err)
	}
	for _, payload := range []string{
		`{"tag_name":"0.6.1","prerelease":false}`,
		`{"tag_name":"v0.6.1-beta.1","prerelease":false}`,
		`{"tag_name":"v0.6.1","prerelease":true}`,
		`{bad`,
	} {
		if _, err := parseLatestReleaseVersion([]byte(payload)); err == nil {
			t.Fatalf("expected parse error for %s", payload)
		}
	}
}

func TestVersionAdvisoryCacheStates(t *testing.T) {
	now := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	cacheRoot := canonicalTempDir(t)
	withCLIVersion(t, "v0.5.0")
	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	app.now = func() time.Time { return now }
	app.userCacheDir = func() (string, error) { return cacheRoot, nil }

	if got := app.readCachedVersionAdvisory(); got.Status != "miss" {
		t.Fatalf("empty cache status=%q want miss", got.Status)
	}

	writeVersionAdvisoryTestCache(t, cacheRoot, versionAdvisoryCacheRecord{
		Source:        versionAdvisorySource,
		Repo:          updateRepo,
		CheckedAt:     now.Add(-time.Hour).Format(time.RFC3339),
		LatestVersion: "v0.6.0",
		Status:        "behind",
	})
	if got := app.readCachedVersionAdvisory(); got.Status != "behind" || got.LatestVersion != "v0.6.0" {
		t.Fatalf("fresh cache = %+v, want behind latest v0.6.0", got)
	}

	writeVersionAdvisoryTestCache(t, cacheRoot, versionAdvisoryCacheRecord{
		Source:        versionAdvisorySource,
		Repo:          updateRepo,
		CheckedAt:     now.Add(-versionAdvisoryCacheTTL - time.Minute).Format(time.RFC3339),
		LatestVersion: "v0.6.0",
		Status:        "behind",
	})
	if got := app.readCachedVersionAdvisory(); got.Status != "stale" {
		t.Fatalf("stale cache status=%q want stale", got.Status)
	}

	mustWrite(t, filepath.Join(cacheRoot, "namba", "version-advisory.json"), "{bad")
	if got := app.readCachedVersionAdvisory(); got.Status != "parse_failed" {
		t.Fatalf("bad cache status=%q want parse_failed", got.Status)
	}

	app.readFile = func(string) ([]byte, error) { return nil, errors.New("permission denied") }
	if got := app.readCachedVersionAdvisory(); got.Status != "cache_read_failed" {
		t.Fatalf("read failure status=%q want cache_read_failed", got.Status)
	}
}

func TestDoctorCheckUpdateRefreshesAdvisoryWithoutRunningUpdate(t *testing.T) {
	root := newReportFixture(t)
	cacheRoot := canonicalTempDir(t)
	withCLIVersion(t, "v0.5.0")
	stdout := &bytes.Buffer{}
	app := NewApp(stdout, &bytes.Buffer{})
	app.getwd = func() (string, error) { return root, nil }
	app.now = func() time.Time { return time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC) }
	app.userCacheDir = func() (string, error) { return cacheRoot, nil }
	app.lookPath = func(name string) (string, error) { return "/bin/" + name, nil }
	app.runCmd = func(context.Context, string, []string, string) (string, error) { return "codex 0.131.0", nil }
	calls := 0
	app.downloadURL = func(_ context.Context, url string) ([]byte, error) {
		calls++
		if url != versionAdvisoryEndpoint {
			t.Fatalf("unexpected URL %s", url)
		}
		return []byte(`{"tag_name":"v0.6.0","prerelease":false}`), nil
	}

	if err := app.Run(context.Background(), []string{"doctor", "--check-update"}); err != nil {
		t.Fatalf("doctor --check-update failed: %v\n%s", err, stdout.String())
	}
	out := stdout.String()
	for _, want := range []string{"NambaAI version advisory", "appears behind latest v0.6.0", "Ask the user before running `namba update`", "not upstream Codex"} {
		if !strings.Contains(out, want) {
			t.Fatalf("doctor output missing %q:\n%s", want, out)
		}
	}
	if calls != 1 {
		t.Fatalf("download calls=%d want 1", calls)
	}
}

func TestStatusJSONDoesNotLookupOrEmitVersionAdvisory(t *testing.T) {
	root := newReportFixture(t)
	cacheRoot := canonicalTempDir(t)
	withCLIVersion(t, "v0.5.0")
	writeVersionAdvisoryTestCache(t, cacheRoot, versionAdvisoryCacheRecord{
		Source:        versionAdvisorySource,
		Repo:          updateRepo,
		CheckedAt:     time.Now().Format(time.RFC3339),
		LatestVersion: "v0.6.0",
		Status:        "behind",
	})
	stdout := &bytes.Buffer{}
	app := NewApp(stdout, &bytes.Buffer{})
	app.getwd = func() (string, error) { return root, nil }
	app.userCacheDir = func() (string, error) { return cacheRoot, nil }
	app.downloadURL = func(context.Context, string) ([]byte, error) {
		t.Fatal("status --json must not perform live update lookup")
		return nil, nil
	}

	if err := app.Run(context.Background(), []string{"status", "--json"}); err != nil {
		t.Fatalf("status --json failed: %v", err)
	}
	if strings.Contains(stdout.String(), "version advisory") || strings.Contains(stdout.String(), "namba update") {
		t.Fatalf("status --json emitted advisory chatter:\n%s", stdout.String())
	}
}

func TestHumanMaintenanceCommandsAppendBehindAdvisoryOnlyFromFreshCache(t *testing.T) {
	root := canonicalTempDir(t)
	if err := NewApp(&bytes.Buffer{}, &bytes.Buffer{}).Run(context.Background(), []string{"init", root, "--yes", "--name", "demo"}); err != nil {
		t.Fatalf("init fixture failed: %v", err)
	}
	cacheRoot := canonicalTempDir(t)
	withCLIVersion(t, "v0.5.0")
	writeVersionAdvisoryTestCache(t, cacheRoot, versionAdvisoryCacheRecord{
		Source:        versionAdvisorySource,
		Repo:          updateRepo,
		CheckedAt:     time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC).Format(time.RFC3339),
		LatestVersion: "v0.6.0",
		Status:        "behind",
	})

	for _, command := range []string{"status", "regen", "project", "sync"} {
		stdout := &bytes.Buffer{}
		app := NewApp(stdout, &bytes.Buffer{})
		app.getwd = func() (string, error) { return root, nil }
		app.now = func() time.Time { return time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC) }
		app.userCacheDir = func() (string, error) { return cacheRoot, nil }
		app.downloadURL = func(context.Context, string) ([]byte, error) {
			t.Fatalf("%s must not perform live update lookup", command)
			return nil, nil
		}
		app.lookPath = func(name string) (string, error) { return "/bin/" + name, nil }
		app.runCmdWithInput = func(context.Context, string, []string, string, string) (string, string, error) { return "", "", nil }
		app.runCodexCmdWithInput = app.runCmdWithInput

		if err := app.Run(context.Background(), []string{command}); err != nil {
			t.Fatalf("%s failed: %v\n%s", command, err, stdout.String())
		}
		out := stdout.String()
		if !strings.Contains(out, "Ask the user before running `namba update`") || !strings.Contains(out, "not upstream Codex") {
			t.Fatalf("%s missing behind advisory:\n%s", command, out)
		}
		if count := strings.Count(out, "NambaAI version advisory:"); count != 1 {
			t.Fatalf("%s emitted advisory %d times, want exactly once:\n%s", command, count, out)
		}
	}
}

func TestHumanMaintenanceCommandsStayQuietForStaleCurrentAndDevCache(t *testing.T) {
	for _, tc := range []struct {
		name      string
		installed string
		checkedAt time.Time
		latest    string
	}{
		{name: "current", installed: "v0.6.0", checkedAt: time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC), latest: "v0.6.0"},
		{name: "dev", installed: "dev", checkedAt: time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC), latest: "v0.6.0"},
		{name: "stale", installed: "v0.5.0", checkedAt: time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC), latest: "v0.6.0"},
	} {
		root := newReportFixture(t)
		cacheRoot := canonicalTempDir(t)
		withCLIVersion(t, tc.installed)
		writeVersionAdvisoryTestCache(t, cacheRoot, versionAdvisoryCacheRecord{
			Source:        versionAdvisorySource,
			Repo:          updateRepo,
			CheckedAt:     tc.checkedAt.Format(time.RFC3339),
			LatestVersion: tc.latest,
			Status:        "behind",
		})
		stdout := &bytes.Buffer{}
		app := NewApp(stdout, &bytes.Buffer{})
		app.getwd = func() (string, error) { return root, nil }
		app.now = func() time.Time { return time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC) }
		app.userCacheDir = func() (string, error) { return cacheRoot, nil }
		if err := app.Run(context.Background(), []string{"status"}); err != nil {
			t.Fatalf("%s status failed: %v", tc.name, err)
		}
		if strings.Contains(stdout.String(), "Ask the user before running `namba update`") {
			t.Fatalf("%s should not prompt update:\n%s", tc.name, stdout.String())
		}
	}
}

func writeVersionAdvisoryTestCache(t *testing.T, cacheRoot string, record versionAdvisoryCacheRecord) {
	t.Helper()
	path := filepath.Join(cacheRoot, "namba", "version-advisory.json")
	mustMkdir(t, filepath.Dir(path))
	data := []byte(`{"source":"` + record.Source + `","repo":"` + record.Repo + `","checked_at":"` + record.CheckedAt + `","latest_version":"` + record.LatestVersion + `","status":"` + record.Status + `"}`)
	mustWrite(t, path, string(data))
}

func withCLIVersion(t *testing.T, version string) {
	t.Helper()
	old := cliVersion
	cliVersion = version
	t.Cleanup(func() { cliVersion = old })
}
