package namba

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func (a *App) writeOutputs(root string, outputs map[string]string) (outputWriteReport, error) {
	session, err := a.beginManagedOutputSessionAllowMalformedManifest(root)
	if err != nil {
		return outputWriteReport{}, err
	}
	if err := session.writeOutputs(outputs); err != nil {
		return outputWriteReport{}, err
	}
	return session.commit()
}

func (a *App) writeManifest(root string, manifest Manifest) error {
	if a.writeManifestOverride != nil {
		return a.writeManifestOverride(root, manifest)
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(root, manifestPath)
	if err := a.mkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	existing, err := a.readFile(path)
	if err == nil && string(existing) == string(data) {
		return nil
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return a.writeFile(path, data, 0o644)
}

func (a *App) readManifest(root string) (Manifest, error) {
	data, err := a.readFile(filepath.Join(root, manifestPath))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Manifest{}, nil
		}
		return Manifest{}, err
	}
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func firstExisting(root string, candidates ...string) string {
	for _, candidate := range candidates {
		if exists(filepath.Join(root, candidate)) {
			return candidate
		}
	}
	return ""
}

func checksum(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}

func manifestKind(rel string) string {
	switch {
	case strings.HasPrefix(rel, ".agents/skills/"), strings.HasPrefix(rel, ".codex/skills/"):
		return "skill"
	case strings.HasPrefix(rel, ".codex/agents/"):
		return "agent"
	case strings.HasPrefix(rel, ".namba/specs/"):
		return "spec"
	case strings.HasPrefix(rel, ".namba/project/"):
		return "project-doc"
	case strings.HasSuffix(rel, ".yaml"):
		return "config"
	default:
		return "asset"
	}
}

func findManifestEntry(manifest Manifest, path string) (ManifestEntry, bool) {
	for _, entry := range manifest.Entries {
		if entry.Path == path {
			return entry, true
		}
	}
	return ManifestEntry{}, false
}

func upsertManifest(manifest Manifest, entry ManifestEntry) Manifest {
	found := false
	for i := range manifest.Entries {
		if manifest.Entries[i].Path == entry.Path {
			manifest.Entries[i] = entry
			found = true
			break
		}
	}
	if !found {
		manifest.Entries = append(manifest.Entries, entry)
	}
	sort.Slice(manifest.Entries, func(i, j int) bool { return manifest.Entries[i].Path < manifest.Entries[j].Path })
	manifest.GeneratedAt = entry.UpdatedAt
	return manifest
}

func manifestEntryIsManaged(entry ManifestEntry, managed func(string) bool, ownedManaged func(ManifestEntry) bool) bool {
	if strings.TrimSpace(entry.Owner) != "" {
		if ownedManaged != nil {
			return ownedManaged(entry)
		}
		return entry.Owner == manifestOwnerManaged && managed(entry.Path)
	}
	return managed(entry.Path)
}
