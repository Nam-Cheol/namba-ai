package namba

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	versionAdvisorySource   = "official GitHub Release metadata"
	versionAdvisoryCacheTTL = 24 * time.Hour
	versionAdvisoryEndpoint = "https://api.github.com/repos/" + updateRepo + "/releases/latest"
)

type versionAdvisoryCacheRecord struct {
	Source        string `json:"source"`
	Repo          string `json:"repo"`
	CheckedAt     string `json:"checked_at"`
	LatestVersion string `json:"latest_version"`
	Status        string `json:"status"`
}

type versionAdvisoryState struct {
	Source           string
	Repo             string
	CheckedAt        time.Time
	InstalledVersion string
	LatestVersion    string
	Status           string
	Detail           string
}

type nambaSemver struct {
	Major int
	Minor int
	Patch int
}

var nambaReleaseVersionPattern = regexp.MustCompile(`^v([0-9]+)\.([0-9]+)\.([0-9]+)$`)

func (a *App) readCachedVersionAdvisory() versionAdvisoryState {
	path, err := a.versionAdvisoryCachePath()
	if err != nil {
		return versionAdvisoryState{Status: "miss", Detail: err.Error()}
	}
	data, err := a.readFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return versionAdvisoryState{Status: "miss"}
		}
		return versionAdvisoryState{Status: "cache_read_failed", Detail: err.Error()}
	}
	record, err := parseVersionAdvisoryCache(data)
	if err != nil {
		return versionAdvisoryState{Status: "parse_failed", Detail: err.Error()}
	}
	checkedAt, err := time.Parse(time.RFC3339, record.CheckedAt)
	if err != nil {
		return versionAdvisoryState{Status: "parse_failed", Detail: err.Error()}
	}
	if a.now().Sub(checkedAt) > versionAdvisoryCacheTTL {
		return versionAdvisoryState{
			Source:        record.Source,
			Repo:          record.Repo,
			CheckedAt:     checkedAt,
			LatestVersion: record.LatestVersion,
			Status:        "stale",
		}
	}
	state := evaluateVersionAdvisory(Version(), record.LatestVersion)
	state.Source = record.Source
	state.Repo = record.Repo
	state.CheckedAt = checkedAt
	return state
}

func (a *App) refreshVersionAdvisory(ctx context.Context) versionAdvisoryState {
	now := a.now()
	data, err := a.downloadURL(ctx, versionAdvisoryEndpoint)
	if err != nil {
		state := versionAdvisoryState{Source: versionAdvisorySource, Repo: updateRepo, CheckedAt: now, Status: "lookup_failed", Detail: err.Error()}
		a.writeVersionAdvisoryCache(state)
		return state
	}
	latest, err := parseLatestReleaseVersion(data)
	if err != nil {
		state := versionAdvisoryState{Source: versionAdvisorySource, Repo: updateRepo, CheckedAt: now, Status: "parse_failed", Detail: err.Error()}
		a.writeVersionAdvisoryCache(state)
		return state
	}
	state := evaluateVersionAdvisory(Version(), latest)
	state.Source = versionAdvisorySource
	state.Repo = updateRepo
	state.CheckedAt = now
	a.writeVersionAdvisoryCache(state)
	return state
}

func (a *App) writeVersionAdvisoryCache(state versionAdvisoryState) {
	path, err := a.versionAdvisoryCachePath()
	if err != nil {
		return
	}
	if err := a.mkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	record := versionAdvisoryCacheRecord{
		Source:        versionAdvisorySource,
		Repo:          updateRepo,
		CheckedAt:     state.CheckedAt.Format(time.RFC3339),
		LatestVersion: state.LatestVersion,
		Status:        state.Status,
	}
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return
	}
	_ = a.writeFile(path, append(data, '\n'), 0o644)
}

func (a *App) versionAdvisoryCachePath() (string, error) {
	cacheRoot, err := a.userCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheRoot, "namba", "version-advisory.json"), nil
}

func parseVersionAdvisoryCache(data []byte) (versionAdvisoryCacheRecord, error) {
	var record versionAdvisoryCacheRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return versionAdvisoryCacheRecord{}, err
	}
	if strings.TrimSpace(record.Source) == "" || strings.TrimSpace(record.Repo) == "" || strings.TrimSpace(record.CheckedAt) == "" {
		return versionAdvisoryCacheRecord{}, errors.New("version advisory cache is missing required fields")
	}
	return record, nil
}

func parseLatestReleaseVersion(data []byte) (string, error) {
	var payload struct {
		TagName    string `json:"tag_name"`
		Prerelease bool   `json:"prerelease"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return "", err
	}
	if payload.Prerelease {
		return "", errors.New("latest release metadata points to a prerelease")
	}
	tag := strings.TrimSpace(payload.TagName)
	if _, ok := parseNambaReleaseVersion(tag); !ok {
		return "", fmt.Errorf("unsupported latest release tag %q", tag)
	}
	return tag, nil
}

func evaluateVersionAdvisory(installed, latest string) versionAdvisoryState {
	installed = strings.TrimSpace(installed)
	latest = strings.TrimSpace(latest)
	state := versionAdvisoryState{
		Source:           versionAdvisorySource,
		Repo:             updateRepo,
		InstalledVersion: installed,
		LatestVersion:    latest,
	}
	if installed == "" || installed == devVersionLabel {
		state.Status = "dev"
		return state
	}
	installedVersion, ok := parseNambaReleaseVersion(installed)
	if !ok {
		state.Status = "dev"
		return state
	}
	latestVersion, ok := parseNambaReleaseVersion(latest)
	if !ok {
		state.Status = "parse_failed"
		return state
	}
	switch compareNambaSemver(installedVersion, latestVersion) {
	case -1:
		state.Status = "behind"
	case 0:
		state.Status = "current"
	default:
		state.Status = "current"
	}
	return state
}

func parseNambaReleaseVersion(raw string) (nambaSemver, bool) {
	matches := nambaReleaseVersionPattern.FindStringSubmatch(strings.TrimSpace(raw))
	if matches == nil {
		return nambaSemver{}, false
	}
	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch, _ := strconv.Atoi(matches[3])
	return nambaSemver{Major: major, Minor: minor, Patch: patch}, true
}

func compareNambaSemver(a, b nambaSemver) int {
	switch {
	case a.Major != b.Major:
		return compareInt(a.Major, b.Major)
	case a.Minor != b.Minor:
		return compareInt(a.Minor, b.Minor)
	default:
		return compareInt(a.Patch, b.Patch)
	}
}

func compareInt(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func formatVersionAdvisory(state versionAdvisoryState, explicit bool) string {
	switch state.Status {
	case "behind":
		return fmt.Sprintf("NambaAI version advisory: installed %s appears behind latest %s from %s. Ask the user before running `namba update`; `namba update` updates only NambaAI, not upstream Codex.\n", state.InstalledVersion, state.LatestVersion, state.Source)
	case "current":
		if explicit {
			return fmt.Sprintf("NambaAI version advisory: installed %s is current against latest %s from %s. No `namba update` prompt is needed.\n", state.InstalledVersion, state.LatestVersion, state.Source)
		}
	case "dev":
		if explicit {
			return "NambaAI version advisory: local CLI version is dev or locally annotated, so no update prompt is shown.\n"
		}
	case "lookup_failed", "parse_failed", "cache_read_failed", "miss", "stale":
		if explicit {
			return fmt.Sprintf("NambaAI version advisory: update metadata is unavailable (%s). This is advisory only; no update was run.\n", state.Status)
		}
	}
	return ""
}

func (a *App) printCachedVersionAdvisory() {
	if text := formatVersionAdvisory(a.readCachedVersionAdvisory(), false); text != "" {
		fmt.Fprint(a.stdout, text)
	}
}
