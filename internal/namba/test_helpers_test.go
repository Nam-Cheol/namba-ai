package namba

import (
	"path/filepath"
	"testing"
)

func canonicalTempDir(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	resolved, err := filepath.EvalSymlinks(dir)
	if err == nil && resolved != "" {
		return resolved
	}

	abs, err := filepath.Abs(dir)
	if err == nil && abs != "" {
		return filepath.Clean(abs)
	}

	return filepath.Clean(dir)
}
