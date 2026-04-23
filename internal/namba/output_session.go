package namba

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type managedOutputSession struct {
	app      *App
	root     string
	manifest Manifest
	report   outputWriteReport
	now      string
	changed  bool
}

func (a *App) beginManagedOutputSession(root string) (*managedOutputSession, error) {
	return a.beginManagedOutputSessionWithOptions(root, false)
}

func (a *App) beginManagedOutputSessionAllowMalformedManifest(root string) (*managedOutputSession, error) {
	return a.beginManagedOutputSessionWithOptions(root, true)
}

func (a *App) beginManagedOutputSessionWithOptions(root string, allowMalformedManifest bool) (*managedOutputSession, error) {
	manifest, err := a.readManifest(root)
	if err != nil {
		if !allowMalformedManifest || !isRecoverableManifestError(err) {
			return nil, err
		}
		manifest = Manifest{}
	}
	return &managedOutputSession{
		app:      a,
		root:     root,
		manifest: manifest,
		now:      a.now().Format(time.RFC3339),
	}, nil
}

func isRecoverableManifestError(err error) bool {
	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		return true
	}
	var typeErr *json.UnmarshalTypeError
	return errors.As(err, &typeErr)
}

func (s *managedOutputSession) replaceManagedOutputs(outputs map[string]string, managed func(string) bool, ownedManaged func(ManifestEntry) bool) error {
	filtered := s.manifest.Entries[:0]
	for _, entry := range s.manifest.Entries {
		if manifestEntryIsManaged(entry, managed, ownedManaged) {
			if _, keep := outputs[entry.Path]; keep {
				filtered = append(filtered, entry)
				continue
			}
			if err := os.RemoveAll(filepath.Join(s.root, filepath.FromSlash(entry.Path))); err != nil && !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("remove obsolete generated file %s: %w", entry.Path, err)
			}
			s.changed = true
			continue
		}
		filtered = append(filtered, entry)
	}
	s.manifest.Entries = filtered
	return s.writeOutputs(outputs)
}

func (s *managedOutputSession) writeOutputs(outputs map[string]string) error {
	for _, rel := range sortedOutputPaths(outputs) {
		content := outputs[rel]
		abs := filepath.Join(s.root, filepath.FromSlash(rel))
		wrote, err := writeFileIfChanged(abs, content)
		if err != nil {
			return err
		}
		entry := ManifestEntry{
			Path:     rel,
			Kind:     manifestKind(rel),
			Owner:    manifestOwnerManaged,
			Checksum: checksum(content),
		}
		if existing, ok := findManifestEntry(s.manifest, rel); ok && !wrote &&
			existing.Kind == entry.Kind &&
			existing.Owner == entry.Owner &&
			existing.Checksum == entry.Checksum {
			continue
		}
		entry.UpdatedAt = s.now
		s.manifest = upsertManifest(s.manifest, entry)
		s.report.ChangedPaths = append(s.report.ChangedPaths, rel)
		if isInstructionSurfacePath(rel) {
			s.report.InstructionSurfacePaths = append(s.report.InstructionSurfacePaths, rel)
		}
		s.changed = true
	}
	return nil
}

func (s *managedOutputSession) commit() (outputWriteReport, error) {
	if !s.changed {
		return s.report, nil
	}
	if err := s.app.writeManifest(s.root, s.manifest); err != nil {
		return s.report, err
	}
	if err := s.app.writeSessionRefreshNotice(s.root, s.report); err != nil {
		return s.report, err
	}
	return s.report, nil
}

func sortedOutputPaths(outputs map[string]string) []string {
	paths := make([]string, 0, len(outputs))
	for rel := range outputs {
		paths = append(paths, rel)
	}
	sort.Strings(paths)
	return paths
}
